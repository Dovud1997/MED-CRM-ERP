package app

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type patientRequest struct {
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	MiddleName         string `json:"middleName"`
	BirthDate          string `json:"birthDate"`
	Gender             string `json:"gender"`
	PhoneLocal         string `json:"phoneLocal"`
	Passport           string `json:"passport"`
	PermanentAddress   string `json:"permanentAddress"`
	GuardianName       string `json:"guardianName"`
	GuardianPhoneLocal string `json:"guardianPhoneLocal"`
	TelegramURL        string `json:"telegramUrl"`
	HomeBranchID       string `json:"homeBranchId"`
	Notes              string `json:"notes"`
}

var patientPhonePattern = regexp.MustCompile(`^[0-9]{9}$`)

func normalizePatient(in *patientRequest) error {
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.MiddleName = strings.TrimSpace(in.MiddleName)
	in.PhoneLocal = strings.ReplaceAll(strings.TrimSpace(in.PhoneLocal), " ", "")
	in.Passport = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(in.Passport), " ", ""))
	in.PermanentAddress = strings.TrimSpace(in.PermanentAddress)
	in.GuardianName = strings.TrimSpace(in.GuardianName)
	in.GuardianPhoneLocal = strings.ReplaceAll(strings.TrimSpace(in.GuardianPhoneLocal), " ", "")
	in.TelegramURL = strings.TrimSpace(in.TelegramURL)
	in.Notes = strings.TrimSpace(in.Notes)
	birth, err := time.Parse("2006-01-02", in.BirthDate)
	if err != nil || birth.After(time.Now()) || len(in.FirstName) < 2 || len(in.LastName) < 2 || !patientPhonePattern.MatchString(in.PhoneLocal) || (in.Gender != "female" && in.Gender != "male") || len(in.PermanentAddress) < 5 {
		return errInvalidPatient
	}
	if in.Passport != "" && !passportPattern.MatchString(in.Passport) {
		return errInvalidPatient
	}
	if in.GuardianPhoneLocal != "" && !patientPhonePattern.MatchString(in.GuardianPhoneLocal) {
		return errInvalidPatient
	}
	if _, err = uuid.Parse(in.HomeBranchID); err != nil {
		return errInvalidPatient
	}
	return nil
}

var errInvalidPatient = &patientValidationError{}

type patientValidationError struct{}

func (*patientValidationError) Error() string { return "invalid patient" }

func (a *App) listPatients(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := a.db.Query(r.Context(), `SELECT p.id,p.first_name,p.last_name,p.middle_name,p.birth_date,p.gender,b.name,p.created_at,EXISTS(SELECT 1 FROM profile_photos ph WHERE ph.organization_id=p.organization_id AND ph.entity_type='patient' AND ph.entity_id=p.id) FROM patients p LEFT JOIN branches b ON b.id=p.home_branch_id WHERE p.organization_id=$1 AND p.deleted_at IS NULL AND ($2='' OR NOT EXISTS (SELECT 1 FROM unnest(regexp_split_to_array($2,'\s+')) AS term WHERE concat_ws(' ',p.last_name,p.first_name,p.middle_name) NOT ILIKE '%'||term||'%')) ORDER BY p.last_name,p.first_name LIMIT 20`, p.OrganizationID, q)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load patients", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, first, last, gender string
		var birth time.Time
		var middle, branch *string
		var created any
		var hasPhoto bool
		if rows.Scan(&id, &first, &last, &middle, &birth, &gender, &branch, &created, &hasPhoto) == nil {
			items = append(items, map[string]any{"id": id, "firstName": first, "lastName": last, "middleName": middle, "birthDate": birth.Format("2006-01-02"), "gender": gender, "branch": branch, "createdAt": created, "hasPhoto": hasPhoto})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items, "total": len(items)})
}

func (a *App) createPatient(w http.ResponseWriter, r *http.Request) {
	var in patientRequest
	if !decode(w, r, &in) {
		return
	}
	if normalizePatient(&in) != nil {
		errorJSON(w, 422, "validation_error", "Patient fields are invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create patient", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branchOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, in.HomeBranchID, p.OrganizationID).Scan(&branchOK)
	if !branchOK {
		errorJSON(w, 422, "invalid_branch", "Branch is invalid", requestID(r))
		return
	}
	id := uuid.NewString()
	_, err = tx.Exec(r.Context(), `INSERT INTO patients(id,organization_id,home_branch_id,first_name,last_name,middle_name,birth_date,gender,phone_hash,phone_encrypted,passport_encrypted,permanent_address_encrypted,guardian_name,guardian_phone_encrypted,telegram_url,notes,created_by) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,encode(digest($9,'sha256'),'hex'),pgp_sym_encrypt($9,$17),CASE WHEN $10='' THEN NULL ELSE pgp_sym_encrypt($10,$17) END,pgp_sym_encrypt($11,$17),NULLIF($12,''),CASE WHEN $13='' THEN NULL ELSE pgp_sym_encrypt($13,$17) END,NULLIF($14,''),NULLIF($15,''),$16)`, id, p.OrganizationID, in.HomeBranchID, in.FirstName, in.LastName, in.MiddleName, in.BirthDate, in.Gender, "+998"+in.PhoneLocal, in.Passport, in.PermanentAddress, in.GuardianName, phoneWithPrefix(in.GuardianPhoneLocal), in.TelegramURL, in.Notes, p.UserID, a.cfg.RefreshSecret)
	if err != nil {
		if strings.Contains(err.Error(), "idx_patients_org_phone") {
			errorJSON(w, 409, "patient_exists", "Patient with this phone already exists", requestID(r))
			return
		}
		errorJSON(w, 500, "database_error", "Could not create patient", requestID(r))
		return
	}
	changes, _ := json.Marshal(map[string]any{"branchId": in.HomeBranchID})
	if _, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'PATIENT_CREATED','patient',$4,'success',$5::jsonb)`, p.OrganizationID, in.HomeBranchID, p.UserID, id, string(changes)); err != nil {
		errorJSON(w, 500, "audit_error", "Could not record patient", requestID(r))
		return
	}
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit patient", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "firstName": in.FirstName, "lastName": in.LastName, "middleName": in.MiddleName, "birthDate": in.BirthDate, "gender": in.Gender, "branchId": in.HomeBranchID})
}

func (a *App) getPatient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid patient id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var firstName, lastName, birthDate, gender, phoneLocal, permanentAddress, homeBranchID string
	var middleName, passport, guardianName, guardianPhoneLocal, telegramURL, notes *string
	var hasPhoto bool
	err := a.db.QueryRow(r.Context(), `SELECT first_name,last_name,middle_name,birth_date::text,gender,
		right(pgp_sym_decrypt(phone_encrypted,$3),9),
		CASE WHEN passport_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(passport_encrypted,$3) END,
		pgp_sym_decrypt(permanent_address_encrypted,$3),guardian_name,
		CASE WHEN guardian_phone_encrypted IS NULL THEN NULL ELSE right(pgp_sym_decrypt(guardian_phone_encrypted,$3),9) END,
		telegram_url,home_branch_id,notes,
		EXISTS(SELECT 1 FROM profile_photos ph WHERE ph.organization_id=patients.organization_id AND ph.entity_type='patient' AND ph.entity_id=patients.id)
		FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL`, id, p.OrganizationID, a.cfg.RefreshSecret).Scan(
		&firstName, &lastName, &middleName, &birthDate, &gender, &phoneLocal, &passport, &permanentAddress,
		&guardianName, &guardianPhoneLocal, &telegramURL, &homeBranchID, &notes, &hasPhoto)
	if err != nil {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "firstName": firstName, "lastName": lastName, "middleName": middleName, "birthDate": birthDate, "gender": gender, "phoneLocal": phoneLocal, "passport": passport, "permanentAddress": permanentAddress, "guardianName": guardianName, "guardianPhoneLocal": guardianPhoneLocal, "telegramUrl": telegramURL, "homeBranchId": homeBranchID, "notes": notes, "hasPhoto": hasPhoto})
}

func (a *App) updatePatient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid patient id", requestID(r))
		return
	}
	var in patientRequest
	if !decode(w, r, &in) {
		return
	}
	if normalizePatient(&in) != nil {
		errorJSON(w, 422, "validation_error", "Patient fields are invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tag, err := a.db.Exec(r.Context(), `UPDATE patients SET home_branch_id=$1,first_name=$2,last_name=$3,middle_name=NULLIF($4,''),birth_date=$5,gender=$6,phone_hash=encode(digest($7,'sha256'),'hex'),phone_encrypted=pgp_sym_encrypt($7,$15),passport_encrypted=CASE WHEN $8='' THEN NULL ELSE pgp_sym_encrypt($8,$15) END,permanent_address_encrypted=pgp_sym_encrypt($9,$15),guardian_name=NULLIF($10,''),guardian_phone_encrypted=CASE WHEN $11='' THEN NULL ELSE pgp_sym_encrypt($11,$15) END,telegram_url=NULLIF($12,''),notes=NULLIF($13,''),updated_at=now(),version=version+1 WHERE id=$14 AND organization_id=$16 AND deleted_at IS NULL`, in.HomeBranchID, in.FirstName, in.LastName, in.MiddleName, in.BirthDate, in.Gender, "+998"+in.PhoneLocal, in.Passport, in.PermanentAddress, in.GuardianName, phoneWithPrefix(in.GuardianPhoneLocal), in.TelegramURL, in.Notes, id, a.cfg.RefreshSecret, p.OrganizationID)
	if err != nil {
		if strings.Contains(err.Error(), "idx_patients_org_phone") {
			errorJSON(w, 409, "patient_exists", "Patient with this phone already exists", requestID(r))
			return
		}
		errorJSON(w, 500, "database_error", "Could not update patient", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'PATIENT_UPDATED','patient',$4,'success')`, p.OrganizationID, in.HomeBranchID, p.UserID, id)
	writeJSON(w, 200, map[string]any{"id": id, "firstName": in.FirstName, "lastName": in.LastName})
}

func (a *App) deletePatient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid patient id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not delete patient", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branchID string
	err = tx.QueryRow(r.Context(), `UPDATE patients SET deleted_at=now(),updated_at=now(),version=version+1 WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL RETURNING home_branch_id`, id, p.OrganizationID).Scan(&branchID)
	if err != nil {
		errorJSON(w, 404, "not_found", "Patient not found", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'PATIENT_ARCHIVED','patient',$4,'success',jsonb_build_object('softDelete',true))`, p.OrganizationID, branchID, p.UserID, id); err != nil {
		errorJSON(w, 500, "audit_error", "Could not record patient deletion", requestID(r))
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		errorJSON(w, 500, "database_error", "Could not delete patient", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func phoneWithPrefix(local string) string {
	if local == "" {
		return ""
	}
	return "+998" + local
}
