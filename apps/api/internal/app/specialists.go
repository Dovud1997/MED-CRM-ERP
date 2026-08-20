package app

import (
	"net/http"
	"strings"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type specialtyRequest struct {
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive"`
}

func (a *App) listSpecialties(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT sc.id,sc.name,sc.is_active,count(DISTINCT e.id),count(DISTINCT ss.service_id) FROM specialty_catalog sc LEFT JOIN employees e ON e.organization_id=sc.organization_id AND lower(e.specialty)=lower(sc.name) LEFT JOIN service_specialties ss ON ss.organization_id=sc.organization_id AND lower(ss.specialty)=lower(sc.name) WHERE sc.organization_id=$1 GROUP BY sc.id ORDER BY sc.name`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load specialists", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name string
		var active bool
		var employees, services int
		if rows.Scan(&id, &name, &active, &employees, &services) == nil {
			items = append(items, map[string]any{"id": id, "name": name, "isActive": active, "employeeCount": employees, "serviceCount": services})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func validSpecialtyName(value string) bool { return len(value) >= 2 && len(value) <= 120 }

func (a *App) createSpecialty(w http.ResponseWriter, r *http.Request) {
	var in specialtyRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if !validSpecialtyName(in.Name) {
		errorJSON(w, 422, "validation_error", "Specialty name is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	_, err := a.db.Exec(r.Context(), `INSERT INTO specialty_catalog(id,organization_id,name) VALUES($1,$2,$3)`, id, p.OrganizationID, in.Name)
	if err != nil {
		errorJSON(w, 409, "specialty_exists", "Specialty already exists", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'SPECIALTY_CREATED','specialty',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, in.Name)
	writeJSON(w, 201, map[string]any{"id": id, "name": in.Name, "isActive": true, "employeeCount": 0, "serviceCount": 0})
}

func (a *App) updateSpecialty(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid specialty id", requestID(r))
		return
	}
	var in specialtyRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if !validSpecialtyName(in.Name) {
		errorJSON(w, 422, "validation_error", "Specialty name is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not update specialty", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var oldName string
	if err = tx.QueryRow(r.Context(), `SELECT name FROM specialty_catalog WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID).Scan(&oldName); err != nil {
		errorJSON(w, 404, "not_found", "Specialty not found", requestID(r))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	if _, err = tx.Exec(r.Context(), `UPDATE specialty_catalog SET name=$1,is_active=$2,updated_at=now() WHERE id=$3`, in.Name, active, id); err != nil {
		errorJSON(w, 409, "specialty_exists", "Specialty already exists", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `UPDATE employees SET specialty=$1,updated_at=now() WHERE organization_id=$2 AND lower(specialty)=lower($3)`, in.Name, p.OrganizationID, oldName); err != nil {
		errorJSON(w, 500, "database_error", "Could not update employees", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `UPDATE service_specialties SET specialty=$1 WHERE organization_id=$2 AND lower(specialty)=lower($3)`, in.Name, p.OrganizationID, oldName); err != nil {
		errorJSON(w, 409, "specialty_conflict", "Service already has this specialty", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'SPECIALTY_UPDATED','specialty',$3,'success',jsonb_build_object('oldName',$4::text,'name',$5::text,'active',$6::boolean))`, p.OrganizationID, p.UserID, id, oldName, in.Name, active)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit specialty", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}

func (a *App) deleteSpecialty(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p := auth.PrincipalFrom(r.Context())
	var name string
	if err := a.db.QueryRow(r.Context(), `SELECT name FROM specialty_catalog WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID).Scan(&name); err != nil {
		errorJSON(w, 404, "not_found", "Specialty not found", requestID(r))
		return
	}
	var count int
	if err := a.db.QueryRow(r.Context(), `SELECT count(*) FROM employees WHERE organization_id=$1 AND lower(specialty)=lower($2)`, p.OrganizationID, name).Scan(&count); err != nil {
		errorJSON(w, 500, "database_error", "Could not check specialty", requestID(r))
		return
	}
	if count > 0 {
		errorJSON(w, 409, "specialty_in_use", "Move employees to another specialty before deleting", requestID(r))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not delete specialty", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	_, _ = tx.Exec(r.Context(), `DELETE FROM service_specialties WHERE organization_id=$1 AND lower(specialty)=lower($2)`, p.OrganizationID, name)
	if _, err = tx.Exec(r.Context(), `DELETE FROM specialty_catalog WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID); err != nil {
		errorJSON(w, 500, "database_error", "Could not delete specialty", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'SPECIALTY_DELETED','specialty',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, name)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit deletion", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}
