package app

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (a *App) getClinicalRecord(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(patientID); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid patient id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var exists bool
	_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, patientID, p.OrganizationID).Scan(&exists)
	if !exists {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	blood := "unknown"
	var height, weight *float64
	_ = a.db.QueryRow(r.Context(), `SELECT blood_group,height_cm,weight_kg FROM patient_clinical_profiles WHERE patient_id=$1 AND organization_id=$2`, patientID, p.OrganizationID).Scan(&blood, &height, &weight)
	allergies := []map[string]any{}
	rows, err := a.db.Query(r.Context(), `SELECT id,allergen,reaction,severity,is_active,created_at FROM patient_allergies WHERE patient_id=$1 AND organization_id=$2 ORDER BY is_active DESC,created_at DESC`, patientID, p.OrganizationID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, allergen, severity string
			var reaction *string
			var active bool
			var created any
			if rows.Scan(&id, &allergen, &reaction, &severity, &active, &created) == nil {
				allergies = append(allergies, map[string]any{"id": id, "allergen": allergen, "reaction": reaction, "severity": severity, "isActive": active, "createdAt": created})
			}
		}
	}
	history := []map[string]any{}
	hrows, err := a.db.Query(r.Context(), `SELECT h.id,h.entry_type,h.occurred_at,h.complaints,h.diagnosis,h.treatment,h.notes,concat_ws(' ',e.first_name,e.last_name),b.name FROM clinical_history_entries h LEFT JOIN employees e ON e.user_id=h.author_id LEFT JOIN branches b ON b.id=h.branch_id WHERE h.patient_id=$1 AND h.organization_id=$2 ORDER BY h.occurred_at DESC LIMIT 200`, patientID, p.OrganizationID)
	if err == nil {
		defer hrows.Close()
		for hrows.Next() {
			var id, kind, author string
			var occurred time.Time
			var complaints, diagnosis, treatment, notes, branch *string
			if hrows.Scan(&id, &kind, &occurred, &complaints, &diagnosis, &treatment, &notes, &author, &branch) == nil {
				history = append(history, map[string]any{"id": id, "type": kind, "occurredAt": occurred, "complaints": complaints, "diagnosis": diagnosis, "treatment": treatment, "notes": notes, "author": author, "branch": branch})
			}
		}
	}
	vaccinations := []map[string]any{}
	vrows, err := a.db.Query(r.Context(), `SELECT id,vaccine_name,administered_on,dose,batch_number,notes FROM patient_vaccinations WHERE patient_id=$1 AND organization_id=$2 ORDER BY administered_on DESC,created_at DESC`, patientID, p.OrganizationID)
	if err == nil {
		defer vrows.Close()
		for vrows.Next() {
			var id, name string
			var administered time.Time
			var dose, batch, notes *string
			if vrows.Scan(&id, &name, &administered, &dose, &batch, &notes) == nil {
				vaccinations = append(vaccinations, map[string]any{"id": id, "name": name, "administeredOn": administered.Format("2006-01-02"), "dose": dose, "batchNumber": batch, "notes": notes})
			}
		}
	}
	labResults := []map[string]any{}
	lrows, err := a.db.Query(r.Context(), `SELECT lr.id,lr.analysis_type,lr.collected_on,lr.test_name,lr.result_value,lr.unit,lr.reference_range,lr.notes,la.id,la.original_name FROM patient_lab_results lr LEFT JOIN patient_lab_attachments la ON la.id=lr.attachment_id AND la.organization_id=lr.organization_id WHERE lr.patient_id=$1 AND lr.organization_id=$2 ORDER BY lr.collected_on DESC,lr.created_at DESC`, patientID, p.OrganizationID)
	if err == nil {
		defer lrows.Close()
		for lrows.Next() {
			var id, kind, name, result string
			var collected time.Time
			var attachmentID, attachmentName *string
			var unit, reference, notes *string
			if lrows.Scan(&id, &kind, &collected, &name, &result, &unit, &reference, &notes, &attachmentID, &attachmentName) == nil {
				labResults = append(labResults, map[string]any{"id": id, "type": kind, "collectedOn": collected.Format("2006-01-02"), "testName": name, "result": result, "unit": unit, "referenceRange": reference, "notes": notes, "attachmentId": attachmentID, "attachmentName": attachmentName})
			}
		}
	}
	imagingStudies := []map[string]any{}
	irows, e := a.db.Query(r.Context(), `SELECT s.id,s.modality,s.performed_on,s.body_area,s.diagnosis,s.conclusion,s.notes,s.original_name,concat_ws(' ',e.first_name,e.last_name) FROM patient_imaging_studies s LEFT JOIN employees e ON e.user_id=s.created_by WHERE s.patient_id=$1 AND s.organization_id=$2 ORDER BY s.performed_on DESC,s.created_at DESC`, patientID, p.OrganizationID)
	if e == nil {
		defer irows.Close()
		for irows.Next() {
			var id, modality, bodyArea, diagnosis, fileName, author string
			var performed time.Time
			var conclusion, notes *string
			if irows.Scan(&id, &modality, &performed, &bodyArea, &diagnosis, &conclusion, &notes, &fileName, &author) == nil {
				imagingStudies = append(imagingStudies, map[string]any{"id": id, "modality": modality, "performedOn": performed.Format("2006-01-02"), "bodyArea": bodyArea, "diagnosis": diagnosis, "conclusion": conclusion, "notes": notes, "fileName": fileName, "author": author})
			}
		}
	}
	diagnoses := []map[string]any{}
	drows, e := a.db.Query(r.Context(), `SELECT d.id,d.diagnosis_name,d.icd10_code,d.diagnosis_type,d.certainty,d.status,d.diagnosed_on,d.notes,concat_ws(' ',e.first_name,e.last_name),b.name FROM patient_diagnoses d LEFT JOIN employees e ON e.user_id=d.created_by LEFT JOIN branches b ON b.id=d.branch_id WHERE d.patient_id=$1 AND d.organization_id=$2 ORDER BY d.diagnosed_on DESC,d.created_at DESC`, patientID, p.OrganizationID)
	if e == nil {
		defer drows.Close()
		for drows.Next() {
			var id, name, kind, certainty, status, author string
			var code, notes, branch *string
			var date time.Time
			if drows.Scan(&id, &name, &code, &kind, &certainty, &status, &date, &notes, &author, &branch) == nil {
				diagnoses = append(diagnoses, map[string]any{"id": id, "name": name, "icd10Code": code, "diagnosisType": kind, "certainty": certainty, "status": status, "diagnosedOn": date.Format("2006-01-02"), "notes": notes, "author": author, "branch": branch})
			}
		}
	}
	orders := []map[string]any{}
	orows, e := a.db.Query(r.Context(), `SELECT o.id,o.diagnosis_id,o.order_type,o.title,o.dosage,o.frequency,o.duration_days,o.instructions,o.start_on,o.end_on,o.status,concat_ws(' ',e.first_name,e.last_name) FROM patient_clinical_orders o LEFT JOIN employees e ON e.user_id=o.created_by WHERE o.patient_id=$1 AND o.organization_id=$2 ORDER BY CASE o.status WHEN 'ACTIVE' THEN 0 ELSE 1 END,o.start_on DESC,o.created_at DESC`, patientID, p.OrganizationID)
	if e == nil {
		defer orows.Close()
		for orows.Next() {
			var id, diagnosisID, kind, title, instructions, status, author string
			var dosage, frequency *string
			var duration *int
			var start time.Time
			var end *time.Time
			if orows.Scan(&id, &diagnosisID, &kind, &title, &dosage, &frequency, &duration, &instructions, &start, &end, &status, &author) == nil {
				var endText any = nil
				if end != nil {
					endText = end.Format("2006-01-02")
				}
				orders = append(orders, map[string]any{"id": id, "diagnosisId": diagnosisID, "orderType": kind, "title": title, "dosage": dosage, "frequency": frequency, "durationDays": duration, "instructions": instructions, "startOn": start.Format("2006-01-02"), "endOn": endText, "status": status, "author": author})
			}
		}
	}
	writeJSON(w, 200, map[string]any{"bloodGroup": blood, "heightCm": height, "weightKg": weight, "allergies": allergies, "history": history, "vaccinations": vaccinations, "labResults": labResults, "imagingStudies": imagingStudies, "diagnoses": diagnoses, "orders": orders})
}

type clinicalProfileRequest struct {
	BloodGroup string  `json:"bloodGroup"`
	HeightCM   float64 `json:"heightCm"`
	WeightKG   float64 `json:"weightKg"`
}

func (a *App) updateClinicalProfile(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in clinicalProfileRequest
	if !decode(w, r, &in) {
		return
	}
	allowed := map[string]bool{"O+": true, "O-": true, "A+": true, "A-": true, "B+": true, "B-": true, "AB+": true, "AB-": true, "unknown": true}
	if _, err := uuid.Parse(patientID); err != nil || !allowed[in.BloodGroup] || in.HeightCM < 30 || in.HeightCM > 250 || in.WeightKG < 0.5 || in.WeightKG > 500 {
		errorJSON(w, 422, "validation_error", "Clinical profile is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tag, err := a.db.Exec(r.Context(), `INSERT INTO patient_clinical_profiles(patient_id,organization_id,blood_group,height_cm,weight_kg,updated_by) SELECT id,$2,$3,$4,$5,$6 FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL ON CONFLICT(patient_id) DO UPDATE SET blood_group=EXCLUDED.blood_group,height_cm=EXCLUDED.height_cm,weight_kg=EXCLUDED.weight_kg,updated_by=EXCLUDED.updated_by,updated_at=now(),version=patient_clinical_profiles.version+1`, patientID, p.OrganizationID, in.BloodGroup, in.HeightCM, in.WeightKG, p.UserID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'CLINICAL_PROFILE_UPDATED','patient',$3,'success',jsonb_build_object('bloodGroup',$4::text,'heightCm',$5::numeric,'weightKg',$6::numeric))`, p.OrganizationID, p.UserID, patientID, in.BloodGroup, in.HeightCM, in.WeightKG)
	writeJSON(w, 200, map[string]any{"bloodGroup": in.BloodGroup, "heightCm": in.HeightCM, "weightKg": in.WeightKG})
}

type vaccinationRequest struct {
	Name           string `json:"name"`
	AdministeredOn string `json:"administeredOn"`
	Dose           string `json:"dose"`
	BatchNumber    string `json:"batchNumber"`
	Notes          string `json:"notes"`
}

func (a *App) addVaccination(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in vaccinationRequest
	if !decode(w, r, &in) {
		return
	}
	date, err := time.Parse("2006-01-02", in.AdministeredOn)
	in.Name = strings.TrimSpace(in.Name)
	if _, e := uuid.Parse(patientID); e != nil || err != nil || len(in.Name) < 2 || date.After(time.Now()) {
		errorJSON(w, 422, "validation_error", "Vaccination is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	tag, err := a.db.Exec(r.Context(), `INSERT INTO patient_vaccinations(id,patient_id,organization_id,vaccine_name,administered_on,dose,batch_number,notes,recorded_by) SELECT $1,id,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),$9 FROM patients WHERE id=$2 AND organization_id=$3 AND deleted_at IS NULL`, id, patientID, p.OrganizationID, in.Name, date, strings.TrimSpace(in.Dose), strings.TrimSpace(in.BatchNumber), strings.TrimSpace(in.Notes), p.UserID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'VACCINATION_ADDED','patient',$3,'success',jsonb_build_object('vaccinationId',$4::text))`, p.OrganizationID, p.UserID, patientID, id)
	writeJSON(w, 201, map[string]any{"id": id, "name": in.Name, "administeredOn": in.AdministeredOn, "dose": strings.TrimSpace(in.Dose), "batchNumber": strings.TrimSpace(in.BatchNumber), "notes": strings.TrimSpace(in.Notes)})
}

type labResultRequest struct {
	Type           string               `json:"type"`
	CollectedOn    string               `json:"collectedOn"`
	TestName       string               `json:"testName"`
	Result         string               `json:"result"`
	Unit           string               `json:"unit"`
	ReferenceRange string               `json:"referenceRange"`
	Notes          string               `json:"notes"`
	Attachment     labAttachmentRequest `json:"attachment"`
}

func (a *App) addLabResult(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in labResultRequest
	if !decode(w, r, &in) {
		return
	}
	date, err := time.Parse("2006-01-02", in.CollectedOn)
	attachmentData, attachmentOK := validateLabAttachment(in.Attachment)
	in.TestName = strings.TrimSpace(in.TestName)
	in.Result = strings.TrimSpace(in.Result)
	if _, e := uuid.Parse(patientID); e != nil || err != nil || (in.Type != "blood" && in.Type != "urine") || len(in.TestName) < 2 || len(in.Result) < 1 || date.After(time.Now()) || !attachmentOK {
		errorJSON(w, 422, "validation_error", "Lab result is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save lab result", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var patientOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, patientID, p.OrganizationID).Scan(&patientOK)
	if !patientOK {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	attachmentID, err := saveLabAttachment(r, tx, patientID, in.Attachment, attachmentData, p.OrganizationID, p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save lab attachment", requestID(r))
		return
	}
	id := uuid.NewString()
	tag, err := tx.Exec(r.Context(), `INSERT INTO patient_lab_results(id,patient_id,organization_id,analysis_type,collected_on,test_name,result_value,unit,reference_range,notes,recorded_by,attachment_id) VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),$11,$12)`, id, patientID, p.OrganizationID, in.Type, date, in.TestName, in.Result, strings.TrimSpace(in.Unit), strings.TrimSpace(in.ReferenceRange), strings.TrimSpace(in.Notes), p.UserID, attachmentID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'LAB_RESULT_ADDED','patient',$3,'success',jsonb_build_object('labResultId',$4::text,'type',$5::text))`, p.OrganizationID, p.UserID, patientID, id, in.Type)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not save lab result", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "type": in.Type, "collectedOn": in.CollectedOn, "testName": in.TestName, "result": in.Result, "unit": strings.TrimSpace(in.Unit), "referenceRange": strings.TrimSpace(in.ReferenceRange), "notes": strings.TrimSpace(in.Notes), "attachmentId": attachmentID, "attachmentName": in.Attachment.Name})
}

type urinePanelItem struct {
	TestName       string `json:"testName"`
	Result         string `json:"result"`
	Unit           string `json:"unit"`
	ReferenceRange string `json:"referenceRange"`
}
type urinePanelRequest struct {
	CollectedOn string               `json:"collectedOn"`
	Items       []urinePanelItem     `json:"items"`
	Attachment  labAttachmentRequest `json:"attachment"`
}

type labAttachmentRequest struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
	Quality     string `json:"quality"`
}

func validateLabAttachment(in labAttachmentRequest) ([]byte, bool) {
	in.Name = strings.TrimSpace(in.Name)
	allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "application/pdf": true}
	data, err := base64.StdEncoding.DecodeString(in.Data)
	return data, err == nil && len(in.Name) >= 1 && len(in.Name) <= 180 && allowed[in.ContentType] && len(data) > 0 && len(data) <= 10*1024*1024 && (in.ContentType == "application/pdf" || in.Quality == "accepted")
}

func saveLabAttachment(r *http.Request, tx pgx.Tx, patientID string, in labAttachmentRequest, data []byte, organizationID, userID string) (string, error) {
	id := uuid.NewString()
	quality := "not_checked"
	if in.ContentType != "application/pdf" {
		quality = "accepted"
	}
	_, err := tx.Exec(r.Context(), `INSERT INTO patient_lab_attachments(id,patient_id,organization_id,original_name,content_type,byte_size,content,image_quality,uploaded_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, id, patientID, organizationID, strings.TrimSpace(in.Name), in.ContentType, len(data), data, quality, userID)
	return id, err
}

func (a *App) getLabAttachment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid attachment id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var name, contentType string
	var content []byte
	err := a.db.QueryRow(r.Context(), `SELECT original_name,content_type,content FROM patient_lab_attachments WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID).Scan(&name, &contentType, &content)
	if err != nil {
		errorJSON(w, 404, "not_found", "Attachment not found", requestID(r))
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `inline; filename="analysis-document"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write(content)
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'LAB_ATTACHMENT_VIEWED','lab_attachment',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, name)
}

type imagingStudyRequest struct {
	Modality    string               `json:"modality"`
	PerformedOn string               `json:"performedOn"`
	BodyArea    string               `json:"bodyArea"`
	Diagnosis   string               `json:"diagnosis"`
	Conclusion  string               `json:"conclusion"`
	Notes       string               `json:"notes"`
	Attachment  labAttachmentRequest `json:"attachment"`
}

func (a *App) addImagingStudy(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in imagingStudyRequest
	if !decode(w, r, &in) {
		return
	}
	in.Modality = strings.ToUpper(strings.TrimSpace(in.Modality))
	in.BodyArea = strings.TrimSpace(in.BodyArea)
	in.Diagnosis = strings.TrimSpace(in.Diagnosis)
	in.Conclusion = strings.TrimSpace(in.Conclusion)
	in.Notes = strings.TrimSpace(in.Notes)
	in.Attachment.Name = strings.TrimSpace(in.Attachment.Name)
	date, err := time.Parse("2006-01-02", in.PerformedOn)
	allowedModality := map[string]bool{"MRI": true, "CT": true, "XRAY": true}
	allowedType := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "application/pdf": true, "application/dicom": true}
	data, decodeErr := base64.StdEncoding.DecodeString(in.Attachment.Data)
	if _, e := uuid.Parse(patientID); e != nil || err != nil || date.After(time.Now()) || !allowedModality[in.Modality] || len(in.BodyArea) < 2 || len(in.Diagnosis) < 2 || len(in.Attachment.Name) < 1 || len(in.Attachment.Name) > 180 || !allowedType[in.Attachment.ContentType] || decodeErr != nil || len(data) < 1 || len(data) > 25*1024*1024 || (strings.HasPrefix(in.Attachment.ContentType, "image/") && in.Attachment.Quality != "accepted") {
		errorJSON(w, 422, "validation_error", "Imaging study or file is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	tag, err := a.db.Exec(r.Context(), `INSERT INTO patient_imaging_studies(id,patient_id,organization_id,modality,performed_on,body_area,diagnosis,conclusion,notes,original_name,content_type,byte_size,content,created_by) SELECT $1,id,$3,$4,$5,$6,$7,NULLIF($8,''),NULLIF($9,''),$10,$11,$12,$13,$14 FROM patients WHERE id=$2 AND organization_id=$3 AND deleted_at IS NULL`, id, patientID, p.OrganizationID, in.Modality, date, in.BodyArea, in.Diagnosis, in.Conclusion, in.Notes, in.Attachment.Name, in.Attachment.ContentType, len(data), data, p.UserID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found or study could not be saved", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'IMAGING_STUDY_ADDED','patient',$3,'success',jsonb_build_object('studyId',$4::text,'modality',$5::text,'diagnosis',$6::text))`, p.OrganizationID, p.UserID, patientID, id, in.Modality, in.Diagnosis)
	writeJSON(w, 201, map[string]any{"id": id, "modality": in.Modality, "performedOn": in.PerformedOn, "bodyArea": in.BodyArea, "diagnosis": in.Diagnosis, "conclusion": in.Conclusion, "notes": in.Notes, "fileName": in.Attachment.Name, "author": ""})
}

func (a *App) getImagingFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid study id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var name, contentType string
	var content []byte
	err := a.db.QueryRow(r.Context(), `SELECT original_name,content_type,content FROM patient_imaging_studies WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID).Scan(&name, &contentType, &content)
	if err != nil {
		errorJSON(w, 404, "not_found", "Imaging file not found", requestID(r))
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `inline; filename="imaging-study"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write(content)
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'IMAGING_FILE_VIEWED','imaging_study',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, name)
}

func (a *App) addUrinePanel(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in urinePanelRequest
	if !decode(w, r, &in) {
		return
	}
	date, err := time.Parse("2006-01-02", in.CollectedOn)
	attachmentData, attachmentOK := validateLabAttachment(in.Attachment)
	if _, e := uuid.Parse(patientID); e != nil || err != nil || date.After(time.Now()) || len(in.Items) == 0 || len(in.Items) > 30 || !attachmentOK {
		errorJSON(w, 422, "validation_error", "Urine panel is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save urine panel", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var patientOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, patientID, p.OrganizationID).Scan(&patientOK)
	if !patientOK {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	attachmentID, err := saveLabAttachment(r, tx, patientID, in.Attachment, attachmentData, p.OrganizationID, p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save lab attachment", requestID(r))
		return
	}
	created := make([]map[string]any, 0, len(in.Items))
	for _, item := range in.Items {
		item.TestName = strings.TrimSpace(item.TestName)
		item.Result = strings.TrimSpace(item.Result)
		item.Unit = strings.TrimSpace(item.Unit)
		item.ReferenceRange = strings.TrimSpace(item.ReferenceRange)
		if len(item.TestName) < 2 || len(item.Result) < 1 {
			errorJSON(w, 422, "validation_error", "Urine panel item is invalid", requestID(r))
			return
		}
		id := uuid.NewString()
		_, err = tx.Exec(r.Context(), `INSERT INTO patient_lab_results(id,patient_id,organization_id,analysis_type,collected_on,test_name,result_value,unit,reference_range,recorded_by,attachment_id) VALUES($1,$2,$3,'urine',$4,$5,$6,NULLIF($7,''),NULLIF($8,''),$9,$10)`, id, patientID, p.OrganizationID, date, item.TestName, item.Result, item.Unit, item.ReferenceRange, p.UserID, attachmentID)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not save urine panel", requestID(r))
			return
		}
		created = append(created, map[string]any{"id": id, "type": "urine", "collectedOn": in.CollectedOn, "testName": item.TestName, "result": item.Result, "unit": item.Unit, "referenceRange": item.ReferenceRange, "notes": "", "attachmentId": attachmentID, "attachmentName": in.Attachment.Name})
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'URINE_PANEL_ADDED','patient',$3,'success',jsonb_build_object('items',$4::int))`, p.OrganizationID, p.UserID, patientID, len(created))
	if err != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit urine panel", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"items": created})
}

func (a *App) addBloodPanel(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in urinePanelRequest
	if !decode(w, r, &in) {
		return
	}
	date, err := time.Parse("2006-01-02", in.CollectedOn)
	attachmentData, attachmentOK := validateLabAttachment(in.Attachment)
	if _, e := uuid.Parse(patientID); e != nil || err != nil || date.After(time.Now()) || len(in.Items) == 0 || len(in.Items) > 30 || !attachmentOK {
		errorJSON(w, 422, "validation_error", "Blood panel is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save blood panel", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var patientOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, patientID, p.OrganizationID).Scan(&patientOK)
	if !patientOK {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	attachmentID, err := saveLabAttachment(r, tx, patientID, in.Attachment, attachmentData, p.OrganizationID, p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save lab attachment", requestID(r))
		return
	}
	created := make([]map[string]any, 0, len(in.Items))
	for _, item := range in.Items {
		item.TestName = strings.TrimSpace(item.TestName)
		item.Result = strings.TrimSpace(item.Result)
		item.Unit = strings.TrimSpace(item.Unit)
		item.ReferenceRange = strings.TrimSpace(item.ReferenceRange)
		if len(item.TestName) < 2 || len(item.Result) < 1 {
			errorJSON(w, 422, "validation_error", "Blood panel item is invalid", requestID(r))
			return
		}
		id := uuid.NewString()
		_, err = tx.Exec(r.Context(), `INSERT INTO patient_lab_results(id,patient_id,organization_id,analysis_type,collected_on,test_name,result_value,unit,reference_range,recorded_by,attachment_id) VALUES($1,$2,$3,'blood',$4,$5,$6,NULLIF($7,''),NULLIF($8,''),$9,$10)`, id, patientID, p.OrganizationID, date, item.TestName, item.Result, item.Unit, item.ReferenceRange, p.UserID, attachmentID)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not save blood panel", requestID(r))
			return
		}
		created = append(created, map[string]any{"id": id, "type": "blood", "collectedOn": in.CollectedOn, "testName": item.TestName, "result": item.Result, "unit": item.Unit, "referenceRange": item.ReferenceRange, "notes": "", "attachmentId": attachmentID, "attachmentName": in.Attachment.Name})
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'BLOOD_PANEL_ADDED','patient',$3,'success',jsonb_build_object('items',$4::int))`, p.OrganizationID, p.UserID, patientID, len(created))
	if err != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit blood panel", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"items": created})
}

type bloodGroupRequest struct {
	BloodGroup string `json:"bloodGroup"`
}

func (a *App) updateBloodGroup(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in bloodGroupRequest
	if !decode(w, r, &in) {
		return
	}
	allowed := map[string]bool{"O+": true, "O-": true, "A+": true, "A-": true, "B+": true, "B-": true, "AB+": true, "AB-": true, "unknown": true}
	if _, err := uuid.Parse(patientID); err != nil || !allowed[in.BloodGroup] {
		errorJSON(w, 422, "validation_error", "Blood group is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tag, err := a.db.Exec(r.Context(), `INSERT INTO patient_clinical_profiles(patient_id,organization_id,blood_group,updated_by) SELECT id,$2,$3,$4 FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL ON CONFLICT(patient_id) DO UPDATE SET blood_group=EXCLUDED.blood_group,updated_by=EXCLUDED.updated_by,updated_at=now(),version=patient_clinical_profiles.version+1`, patientID, p.OrganizationID, in.BloodGroup, p.UserID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'BLOOD_GROUP_UPDATED','patient',$3,'success',jsonb_build_object('bloodGroup',$4::text))`, p.OrganizationID, p.UserID, patientID, in.BloodGroup)
	writeJSON(w, 200, map[string]string{"bloodGroup": in.BloodGroup})
}

type allergyRequest struct {
	Allergen string `json:"allergen"`
	Reaction string `json:"reaction"`
	Severity string `json:"severity"`
}

func (a *App) addAllergy(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in allergyRequest
	if !decode(w, r, &in) {
		return
	}
	in.Allergen = strings.TrimSpace(in.Allergen)
	in.Reaction = strings.TrimSpace(in.Reaction)
	allowed := map[string]bool{"mild": true, "moderate": true, "severe": true, "unknown": true}
	if _, err := uuid.Parse(patientID); err != nil || len(in.Allergen) < 2 || !allowed[in.Severity] {
		errorJSON(w, 422, "validation_error", "Allergy is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	tag, err := a.db.Exec(r.Context(), `INSERT INTO patient_allergies(id,patient_id,organization_id,allergen,reaction,severity,recorded_by) SELECT $1,id,$3,$4,NULLIF($5,''),$6,$7 FROM patients WHERE id=$2 AND organization_id=$3 AND deleted_at IS NULL`, id, patientID, p.OrganizationID, in.Allergen, in.Reaction, in.Severity, p.UserID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'ALLERGY_ADDED','patient',$3,'success',jsonb_build_object('allergyId',$4::text,'severity',$5::text))`, p.OrganizationID, p.UserID, patientID, id, in.Severity)
	writeJSON(w, 201, map[string]any{"id": id, "allergen": in.Allergen, "reaction": in.Reaction, "severity": in.Severity, "isActive": true})
}

type historyRequest struct {
	Type       string `json:"type"`
	OccurredAt string `json:"occurredAt"`
	BranchID   string `json:"branchId"`
	Complaints string `json:"complaints"`
	Diagnosis  string `json:"diagnosis"`
	Treatment  string `json:"treatment"`
	Notes      string `json:"notes"`
}

func (a *App) addClinicalHistory(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	var in historyRequest
	if !decode(w, r, &in) {
		return
	}
	occurred, err := time.Parse(time.RFC3339, in.OccurredAt)
	allowed := map[string]bool{"visit": true, "diagnosis": true, "prescription": true, "procedure": true, "note": true}
	if _, e := uuid.Parse(patientID); e != nil || err != nil || !allowed[in.Type] || occurred.After(time.Now().Add(5*time.Minute)) {
		errorJSON(w, 422, "validation_error", "History entry is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	tag, err := a.db.Exec(r.Context(), `INSERT INTO clinical_history_entries(id,patient_id,organization_id,branch_id,author_id,entry_type,occurred_at,complaints,diagnosis,treatment,notes) SELECT $1,p.id,$3,b.id,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,'') FROM patients p JOIN branches b ON b.id=$11 AND b.organization_id=$3 AND b.deleted_at IS NULL WHERE p.id=$2 AND p.organization_id=$3 AND p.deleted_at IS NULL`, id, patientID, p.OrganizationID, p.UserID, in.Type, occurred, strings.TrimSpace(in.Complaints), strings.TrimSpace(in.Diagnosis), strings.TrimSpace(in.Treatment), strings.TrimSpace(in.Notes), in.BranchID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_reference", "Patient or branch is invalid", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'CLINICAL_HISTORY_ADDED','patient',$4,'success',jsonb_build_object('historyId',$5::text,'type',$6::text))`, p.OrganizationID, in.BranchID, p.UserID, patientID, id, in.Type)
	writeJSON(w, 201, map[string]any{"id": id, "type": in.Type, "occurredAt": occurred})
}
