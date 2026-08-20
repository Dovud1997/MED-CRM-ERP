package app

import (
	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type diagnosisCatalogRequest struct {
	SpecialtyID string `json:"specialtyId"`
	Name string `json:"name"`
	ICD10Code string `json:"icd10Code"`
	Description string `json:"description"`
	IsActive *bool `json:"isActive"`
}

func (a *App) listDiagnosisCatalog(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	specialty := r.URL.Query().Get("specialtyId")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := a.db.Query(r.Context(), `SELECT d.id,d.specialty_id,s.name,d.name,d.icd10_code,d.description,d.is_active FROM diagnosis_catalog d JOIN specialty_catalog s ON s.id=d.specialty_id WHERE d.organization_id=$1 AND ($2='' OR d.specialty_id::text=$2) AND ($3='' OR d.name ILIKE '%'||$3||'%' OR COALESCE(d.icd10_code,'') ILIKE '%'||$3||'%') ORDER BY s.name,d.name`, p.OrganizationID, specialty, q)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load diagnosis catalog", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, sid, sname, name string
		var code, description *string
		var active bool
		if rows.Scan(&id, &sid, &sname, &name, &code, &description, &active) == nil {
			items = append(items, map[string]any{"id": id, "specialtyId": sid, "specialtyName": sname, "name": name, "icd10Code": code, "description": description, "isActive": active})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func validDiagnosisCatalog(in diagnosisCatalogRequest) bool {
	in.Name = strings.TrimSpace(in.Name)
	return len(in.Name) >= 2 && len(in.Name) <= 240
}
func (a *App) createDiagnosisCatalogItem(w http.ResponseWriter, r *http.Request) {
	var in diagnosisCatalogRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.ICD10Code = strings.ToUpper(strings.TrimSpace(in.ICD10Code))
	if !validDiagnosisCatalog(in) {
		errorJSON(w, 422, "validation_error", "Diagnosis is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var ok bool
	_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM specialty_catalog WHERE id=$1 AND organization_id=$2)`, in.SpecialtyID, p.OrganizationID).Scan(&ok)
	if !ok {
		errorJSON(w, 422, "invalid_specialty", "Invalid specialty", requestID(r))
		return
	}
	id := uuid.NewString()
	_, err := a.db.Exec(r.Context(), `INSERT INTO diagnosis_catalog(id,organization_id,specialty_id,name,icd10_code,description) VALUES($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''))`, id, p.OrganizationID, in.SpecialtyID, in.Name, in.ICD10Code, strings.TrimSpace(in.Description))
	if err != nil {
		errorJSON(w, 409, "diagnosis_exists", "Diagnosis already exists", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'DIAGNOSIS_CATALOG_CREATED','diagnosis_catalog',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, in.Name)
	writeJSON(w, 201, map[string]string{"id": id})
}

func (a *App) updateDiagnosisCatalogItem(w http.ResponseWriter, r *http.Request) {
	var in diagnosisCatalogRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.ICD10Code = strings.ToUpper(strings.TrimSpace(in.ICD10Code))
	if !validDiagnosisCatalog(in) {
		errorJSON(w, 422, "validation_error", "Diagnosis is invalid", requestID(r))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	var specialtyExists bool
	_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM specialty_catalog WHERE id=$1 AND organization_id=$2)`, in.SpecialtyID, p.OrganizationID).Scan(&specialtyExists)
	if !specialtyExists {
		errorJSON(w, 422, "invalid_specialty", "Invalid specialty", requestID(r))
		return
	}
	tag, err := a.db.Exec(r.Context(), `UPDATE diagnosis_catalog SET specialty_id=$1,name=$2,icd10_code=NULLIF($3,''),description=NULLIF($4,''),is_active=$5,updated_at=now() WHERE id=$6 AND organization_id=$7`, in.SpecialtyID, in.Name, in.ICD10Code, strings.TrimSpace(in.Description), active, id, p.OrganizationID)
	if err != nil {
		errorJSON(w, 409, "diagnosis_exists", "Diagnosis already exists", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Diagnosis not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'DIAGNOSIS_CATALOG_UPDATED','diagnosis_catalog',$3,'success',jsonb_build_object('name',$4::text,'active',$5::boolean))`, p.OrganizationID, p.UserID, id, in.Name, active)
	writeJSON(w, 200, map[string]bool{"success": true})
}

func (a *App) deleteDiagnosisCatalogItem(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid diagnosis id", requestID(r))
		return
	}
	tag, err := a.db.Exec(r.Context(), `DELETE FROM diagnosis_catalog WHERE id=$1 AND organization_id=$2 AND NOT EXISTS(SELECT 1 FROM patient_diagnoses WHERE catalog_id=$1)`, id, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not delete diagnosis", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 409, "diagnosis_in_use", "Diagnosis is used or not found; disable it instead", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'DIAGNOSIS_CATALOG_DELETED','diagnosis_catalog',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]bool{"success": true})
}
