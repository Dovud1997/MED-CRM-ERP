package app

import (
	"net/http"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type appointmentRequest struct {
	PatientID  string `json:"patientId"`
	EmployeeID string `json:"employeeId"`
	BranchID   string `json:"branchId"`
	StartsAt   string `json:"startsAt"`
	Duration   int    `json:"durationMinutes"`
	Reason     string `json:"reason"`
	Notes      string `json:"notes"`
}

func (a *App) listAppointments(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	if r.URL.Query().Get("scope") == "active" {
		rows, err := a.db.Query(r.Context(), `SELECT a.id,a.starts_at,a.ends_at,a.status,a.reason,a.notes,p.id,p.last_name,p.first_name,p.middle_name,pgp_sym_decrypt(p.phone_encrypted,$2),CASE WHEN p.guardian_phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(p.guardian_phone_encrypted,$2) END,p.telegram_url,e.id,e.user_id,e.last_name,e.first_name,e.position,e.specialty,CASE WHEN e.phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(e.phone_encrypted,$2) END,e.public_email,e.telegram_url,b.id,b.name FROM appointments a JOIN patients p ON p.id=a.patient_id JOIN employees e ON e.id=a.employee_id JOIN branches b ON b.id=a.branch_id WHERE a.organization_id=$1 AND a.starts_at>=now() AND a.status NOT IN ('completed','cancelled','no_show') ORDER BY a.starts_at LIMIT 100`, p.OrganizationID, a.cfg.RefreshSecret)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not load active appointments", requestID(r))
			return
		}
		defer rows.Close()
		writeAppointmentRows(w, rows, "active")
		return
	}
	date := r.URL.Query().Get("date")
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		day = time.Now()
	}
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.FixedZone("UZT", 5*60*60))
	end := start.Add(24 * time.Hour)
	rows, err := a.db.Query(r.Context(), `SELECT a.id,a.starts_at,a.ends_at,a.status,a.reason,a.notes,p.id,p.last_name,p.first_name,p.middle_name,pgp_sym_decrypt(p.phone_encrypted,$4),CASE WHEN p.guardian_phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(p.guardian_phone_encrypted,$4) END,p.telegram_url,e.id,e.user_id,e.last_name,e.first_name,e.position,e.specialty,CASE WHEN e.phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(e.phone_encrypted,$4) END,e.public_email,e.telegram_url,b.id,b.name FROM appointments a JOIN patients p ON p.id=a.patient_id JOIN employees e ON e.id=a.employee_id JOIN branches b ON b.id=a.branch_id WHERE a.organization_id=$1 AND a.starts_at>=$2 AND a.starts_at<$3 ORDER BY a.starts_at`, p.OrganizationID, start, end, a.cfg.RefreshSecret)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load appointments", requestID(r))
		return
	}
	defer rows.Close()
	writeAppointmentRows(w, rows, start.Format("2006-01-02"))
}

type appointmentRows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
}

func writeAppointmentRows(w http.ResponseWriter, rows appointmentRows, date string) {
	items := []map[string]any{}
	for rows.Next() {
		var id, status, patientID, patientLast, patientFirst, employeeID, employeeUserID, employeeLast, employeeFirst, position, branchID, branchName string
		var starts, ends time.Time
		var reason, notes, patientMiddle, guardianPhone, telegramURL, specialty, employeePhone, employeeEmail, employeeTelegram *string
		var phone string
		if rows.Scan(&id, &starts, &ends, &status, &reason, &notes, &patientID, &patientLast, &patientFirst, &patientMiddle, &phone, &guardianPhone, &telegramURL, &employeeID, &employeeUserID, &employeeLast, &employeeFirst, &position, &specialty, &employeePhone, &employeeEmail, &employeeTelegram, &branchID, &branchName) == nil {
			items = append(items, map[string]any{"id": id, "startsAt": starts, "endsAt": ends, "status": status, "reason": reason, "notes": notes, "patient": map[string]any{"id": patientID, "name": strings.TrimSpace(patientLast + " " + patientFirst + " " + valueOrEmpty(patientMiddle)), "phone": phone, "secondPhone": guardianPhone, "telegram": telegramURL}, "employee": map[string]any{"id": employeeID, "userId": employeeUserID, "name": employeeLast + " " + employeeFirst, "position": position, "specialty": specialty, "phone": employeePhone, "email": employeeEmail, "telegram": employeeTelegram}, "branch": map[string]any{"id": branchID, "name": branchName}})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items, "date": date})
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (a *App) getAppointmentsDashboard(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var today, waiting, completed, doctorsOnDuty int
	err := a.db.QueryRow(r.Context(), `SELECT count(*),count(*) FILTER (WHERE status IN ('scheduled','confirmed','arrived')),count(*) FILTER (WHERE status='completed') FROM appointments WHERE organization_id=$1 AND (starts_at AT TIME ZONE 'Asia/Tashkent')::date=(now() AT TIME ZONE 'Asia/Tashkent')::date`, p.OrganizationID).Scan(&today, &waiting, &completed)
	if err == nil {
		err = a.db.QueryRow(r.Context(), `SELECT count(DISTINCT employee_id) FROM doctor_schedules WHERE organization_id=$1 AND weekday=extract(isodow FROM (now() AT TIME ZONE 'Asia/Tashkent'))::int AND is_working=true AND (now() AT TIME ZONE 'Asia/Tashkent')::time BETWEEN start_time AND end_time`, p.OrganizationID).Scan(&doctorsOnDuty)
	}
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load dashboard summary", requestID(r))
		return
	}
	rows, err := a.db.Query(r.Context(), `SELECT a.id,a.starts_at,a.status,a.reason,p.last_name,p.first_name,p.middle_name,pgp_sym_decrypt(p.phone_encrypted,$2),CASE WHEN p.guardian_phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(p.guardian_phone_encrypted,$2) END,p.telegram_url,e.user_id,e.last_name,e.first_name,e.position,e.specialty,CASE WHEN e.phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(e.phone_encrypted,$2) END,e.public_email,e.telegram_url,b.name FROM appointments a JOIN patients p ON p.id=a.patient_id JOIN employees e ON e.id=a.employee_id JOIN branches b ON b.id=a.branch_id WHERE a.organization_id=$1 AND a.starts_at>=now() AND a.status NOT IN ('cancelled','no_show') ORDER BY a.starts_at LIMIT 8`, p.OrganizationID, a.cfg.RefreshSecret)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load upcoming appointments", requestID(r))
		return
	}
	defer rows.Close()
	upcoming := []map[string]any{}
	for rows.Next() {
		var id, status, patientLast, patientFirst, employeeUserID, employeeLast, employeeFirst, position, branch string
		var starts time.Time
		var reason, patientMiddle, guardianPhone, telegramURL, specialty, employeePhone, employeeEmail, employeeTelegram *string
		var phone string
		if rows.Scan(&id, &starts, &status, &reason, &patientLast, &patientFirst, &patientMiddle, &phone, &guardianPhone, &telegramURL, &employeeUserID, &employeeLast, &employeeFirst, &position, &specialty, &employeePhone, &employeeEmail, &employeeTelegram, &branch) == nil {
			upcoming = append(upcoming, map[string]any{"id": id, "startsAt": starts, "status": status, "reason": reason, "patient": strings.TrimSpace(patientLast + " " + patientFirst + " " + valueOrEmpty(patientMiddle)), "phone": phone, "secondPhone": guardianPhone, "telegram": telegramURL, "employeeUserId": employeeUserID, "employee": employeeLast + " " + employeeFirst, "employeePhone": employeePhone, "employeeEmail": employeeEmail, "employeeTelegram": employeeTelegram, "position": position, "specialty": specialty, "branch": branch})
		}
	}
	writeJSON(w, 200, map[string]any{"today": today, "waiting": waiting, "completed": completed, "doctorsOnDuty": doctorsOnDuty, "upcoming": upcoming})
}

func (a *App) createAppointment(w http.ResponseWriter, r *http.Request) {
	var in appointmentRequest
	if !decode(w, r, &in) {
		return
	}
	starts, err := time.Parse(time.RFC3339, in.StartsAt)
	if _, e := uuid.Parse(in.PatientID); e != nil {
		err = e
	}
	if _, e := uuid.Parse(in.EmployeeID); e != nil {
		err = e
	}
	if _, e := uuid.Parse(in.BranchID); e != nil {
		err = e
	}
	in.Reason = strings.TrimSpace(in.Reason)
	in.Notes = strings.TrimSpace(in.Notes)
	if err != nil || in.Duration < 15 || in.Duration > 240 {
		errorJSON(w, 422, "validation_error", "Appointment fields are invalid", requestID(r))
		return
	}
	if starts.Before(time.Now().Add(-5 * time.Minute)) {
		errorJSON(w, 422, "appointment_in_past", "Appointment cannot be created in the past", requestID(r))
		return
	}
	ends := starts.Add(time.Duration(in.Duration) * time.Minute)
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create appointment", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var refsOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patients p JOIN employees e ON e.id=$2 AND e.organization_id=$4 AND e.branch_id=$3 JOIN users u ON u.id=e.user_id AND u.deleted_at IS NULL AND u.is_active=true JOIN branches b ON b.id=$3 AND b.organization_id=$4 WHERE p.id=$1 AND p.organization_id=$4 AND p.deleted_at IS NULL AND b.deleted_at IS NULL AND b.is_active=true)`, in.PatientID, in.EmployeeID, in.BranchID, p.OrganizationID).Scan(&refsOK)
	if !refsOK {
		errorJSON(w, 422, "invalid_reference", "Patient, doctor or branch is invalid", requestID(r))
		return
	}
	var scheduleConfigured, withinSchedule bool
	err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM doctor_schedules WHERE organization_id=$1 AND employee_id=$2), EXISTS(SELECT 1 FROM doctor_schedules s WHERE s.organization_id=$1 AND s.employee_id=$2 AND s.weekday=extract(isodow FROM ($3::timestamptz AT TIME ZONE 'Asia/Tashkent'))::int AND s.is_working=true AND ($3::timestamptz AT TIME ZONE 'Asia/Tashkent')::time >= s.start_time AND ($4::timestamptz AT TIME ZONE 'Asia/Tashkent')::time <= s.end_time AND ($3::timestamptz AT TIME ZONE 'Asia/Tashkent')::date = ($4::timestamptz AT TIME ZONE 'Asia/Tashkent')::date AND (s.break_start IS NULL OR NOT (($3::timestamptz AT TIME ZONE 'Asia/Tashkent')::time < s.break_end AND ($4::timestamptz AT TIME ZONE 'Asia/Tashkent')::time > s.break_start)))`, p.OrganizationID, in.EmployeeID, starts, ends).Scan(&scheduleConfigured, &withinSchedule)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not verify doctor schedule", requestID(r))
		return
	}
	if scheduleConfigured && !withinSchedule {
		errorJSON(w, 422, "outside_working_hours", "Appointment is outside doctor working hours", requestID(r))
		return
	}
	id := uuid.NewString()
	_, err = tx.Exec(r.Context(), `INSERT INTO appointments(id,organization_id,branch_id,patient_id,employee_id,starts_at,ends_at,reason,notes,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),NULLIF($9,''),$10)`, id, p.OrganizationID, in.BranchID, in.PatientID, in.EmployeeID, starts, ends, in.Reason, in.Notes, p.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "appointments_doctor_no_overlap") {
			errorJSON(w, 409, "schedule_conflict", "Doctor already has an appointment at this time", requestID(r))
			return
		}
		errorJSON(w, 500, "database_error", "Could not create appointment", requestID(r))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'APPOINTMENT_CREATED','appointment',$4,'success',jsonb_build_object('startsAt',$5::timestamptz,'patientId',$6::text,'employeeId',$7::text))`, p.OrganizationID, in.BranchID, p.UserID, id, starts, in.PatientID, in.EmployeeID)
	if err != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit appointment", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "startsAt": starts, "endsAt": ends, "status": "scheduled"})
}

type appointmentStatusRequest struct {
	Status string `json:"status"`
}

func (a *App) updateAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in appointmentStatusRequest
	if !decode(w, r, &in) {
		return
	}
	allowed := map[string]bool{"scheduled": true, "confirmed": true, "arrived": true, "in_progress": true, "completed": true, "cancelled": true, "no_show": true}
	if _, err := uuid.Parse(id); err != nil || !allowed[in.Status] {
		errorJSON(w, 422, "validation_error", "Status is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tag, err := a.db.Exec(r.Context(), `UPDATE appointments SET status=$1,updated_at=now(),version=version+1 WHERE id=$2 AND organization_id=$3`, in.Status, id, p.OrganizationID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Appointment not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'APPOINTMENT_STATUS_UPDATED','appointment',$3,'success',jsonb_build_object('status',$4::text))`, p.OrganizationID, p.UserID, id, in.Status)
	writeJSON(w, 200, map[string]string{"id": id, "status": in.Status})
}
