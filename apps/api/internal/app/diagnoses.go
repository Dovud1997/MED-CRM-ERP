package app

import (
	"net/http"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type diagnosisRequest struct{ Name, ICD10Code, DiagnosisType, Certainty, DiagnosedOn, Notes, BranchID string }

func (a *App) addPatientDiagnosis(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in diagnosisRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.ICD10Code = strings.ToUpper(strings.TrimSpace(in.ICD10Code))
	in.DiagnosisType = strings.ToUpper(in.DiagnosisType)
	in.Certainty = strings.ToUpper(in.Certainty)
	date, err := time.Parse("2006-01-02", in.DiagnosedOn)
	if _, e := uuid.Parse(patientID); e != nil || err != nil || date.After(time.Now()) || len(in.Name) < 2 || (in.DiagnosisType != "PRIMARY" && in.DiagnosisType != "SECONDARY") || (in.Certainty != "PRELIMINARY" && in.Certainty != "CONFIRMED") {
		errorJSON(w, 422, "validation_error", "Diagnosis is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var branch any = nil
	if in.BranchID != "" {
		if _, e := uuid.Parse(in.BranchID); e != nil {
			errorJSON(w, 422, "invalid_branch", "Invalid branch", requestID(r))
			return
		}
		var ok bool
		_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, in.BranchID, p.OrganizationID).Scan(&ok)
		if !ok {
			errorJSON(w, 422, "invalid_branch", "Invalid branch", requestID(r))
			return
		}
		branch = in.BranchID
	}
	id := uuid.NewString()
	tag, err := a.db.Exec(r.Context(), `INSERT INTO patient_diagnoses(id,patient_id,organization_id,branch_id,diagnosis_name,icd10_code,diagnosis_type,certainty,diagnosed_on,notes,created_by) SELECT $1,id,$3,$4,$5,NULLIF($6,''),$7,$8,$9,NULLIF($10,''),$11 FROM patients WHERE id=$2 AND organization_id=$3 AND deleted_at IS NULL`, id, patientID, p.OrganizationID, branch, in.Name, in.ICD10Code, in.DiagnosisType, in.Certainty, date, strings.TrimSpace(in.Notes), p.UserID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'DIAGNOSIS_ADDED','patient',$4,'success',jsonb_build_object('diagnosisId',$5::text,'name',$6::text,'certainty',$7::text))`, p.OrganizationID, branch, p.UserID, patientID, id, in.Name, in.Certainty)
	writeJSON(w, 201, map[string]any{"id": id, "name": in.Name, "icd10Code": in.ICD10Code, "diagnosisType": in.DiagnosisType, "certainty": in.Certainty, "status": "ACTIVE", "diagnosedOn": in.DiagnosedOn, "notes": strings.TrimSpace(in.Notes), "author": "", "branch": ""})
}

func (a *App) updateDiagnosisStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Status = strings.ToUpper(in.Status)
	if in.Status != "ACTIVE" && in.Status != "RESOLVED" && in.Status != "RULED_OUT" {
		errorJSON(w, 422, "validation_error", "Invalid diagnosis status", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag, err := a.db.Exec(r.Context(), `UPDATE patient_diagnoses SET status=$1,updated_at=now() WHERE id=$2 AND organization_id=$3`, in.Status, id, p.OrganizationID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Diagnosis not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'DIAGNOSIS_STATUS_CHANGED','diagnosis',$3,'success',jsonb_build_object('status',$4::text))`, p.OrganizationID, p.UserID, id, in.Status)
	writeJSON(w, 200, map[string]string{"status": in.Status})
}

type clinicalOrderRequest struct {
	DiagnosisID, BranchID, OrderType, Title, Dosage, Frequency, Instructions, StartOn, EndOn string
	DurationDays                                                                             *int `json:"durationDays"`
}

func (a *App) addClinicalOrder(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in clinicalOrderRequest
	if !decode(w, r, &in) {
		return
	}
	in.OrderType = strings.ToUpper(in.OrderType)
	in.Title = strings.TrimSpace(in.Title)
	in.Instructions = strings.TrimSpace(in.Instructions)
	start, err := time.Parse("2006-01-02", in.StartOn)
	var end any = nil
	if in.EndOn != "" {
		parsed, e := time.Parse("2006-01-02", in.EndOn)
		if e != nil || parsed.Before(start) {
			errorJSON(w, 422, "validation_error", "Invalid end date", requestID(r))
			return
		}
		end = parsed
	}
	allowed := map[string]bool{"MEDICATION": true, "LAB": true, "IMAGING": true, "REFERRAL": true, "PROCEDURE": true, "FOLLOW_UP": true}
	if _, e := uuid.Parse(patientID); e != nil || err != nil || !allowed[in.OrderType] || len(in.Title) < 2 || len(in.Instructions) < 2 || in.DurationDays != nil && (*in.DurationDays < 1 || *in.DurationDays > 3650) {
		errorJSON(w, 422, "validation_error", "Clinical order is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var diagnosisOK bool
	_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patient_diagnoses WHERE id=$1 AND patient_id=$2 AND organization_id=$3)`, in.DiagnosisID, patientID, p.OrganizationID).Scan(&diagnosisOK)
	if !diagnosisOK {
		errorJSON(w, 422, "invalid_diagnosis", "Select a patient diagnosis", requestID(r))
		return
	}
	var branch any = nil
	if in.BranchID != "" {
		var ok bool
		_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, in.BranchID, p.OrganizationID).Scan(&ok)
		if !ok {
			errorJSON(w, 422, "invalid_branch", "Invalid branch", requestID(r))
			return
		}
		branch = in.BranchID
	}
	id := uuid.NewString()
	_, err = a.db.Exec(r.Context(), `INSERT INTO patient_clinical_orders(id,patient_id,organization_id,diagnosis_id,branch_id,order_type,title,dosage,frequency,duration_days,instructions,start_on,end_on,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),NULLIF($9,''),$10,$11,$12,$13,$14)`, id, patientID, p.OrganizationID, in.DiagnosisID, branch, in.OrderType, in.Title, strings.TrimSpace(in.Dosage), strings.TrimSpace(in.Frequency), in.DurationDays, in.Instructions, start, end, p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save clinical order", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'CLINICAL_ORDER_ADDED','patient',$4,'success',jsonb_build_object('orderId',$5::text,'diagnosisId',$6::text,'type',$7::text))`, p.OrganizationID, branch, p.UserID, patientID, id, in.DiagnosisID, in.OrderType)
	writeJSON(w, 201, map[string]any{"id": id, "diagnosisId": in.DiagnosisID, "orderType": in.OrderType, "title": in.Title, "dosage": strings.TrimSpace(in.Dosage), "frequency": strings.TrimSpace(in.Frequency), "durationDays": in.DurationDays, "instructions": in.Instructions, "startOn": in.StartOn, "endOn": in.EndOn, "status": "ACTIVE", "author": ""})
}

func (a *App) updateClinicalOrderStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Status = strings.ToUpper(in.Status)
	if in.Status != "ACTIVE" && in.Status != "COMPLETED" && in.Status != "CANCELLED" {
		errorJSON(w, 422, "validation_error", "Invalid order status", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag, err := a.db.Exec(r.Context(), `UPDATE patient_clinical_orders SET status=$1,updated_at=now() WHERE id=$2 AND organization_id=$3`, in.Status, id, p.OrganizationID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Clinical order not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'CLINICAL_ORDER_STATUS_CHANGED','clinical_order',$3,'success',jsonb_build_object('status',$4::text))`, p.OrganizationID, p.UserID, id, in.Status)
	writeJSON(w, 200, map[string]string{"status": in.Status})
}
