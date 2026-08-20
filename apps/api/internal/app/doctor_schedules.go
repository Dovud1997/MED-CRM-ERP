package app

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type scheduleDayRequest struct {
	Weekday   int    `json:"weekday"`
	IsWorking bool   `json:"isWorking"`
	Start     string `json:"start"`
	End       string `json:"end"`
	BreakFrom string `json:"breakFrom"`
	BreakTo   string `json:"breakTo"`
}

type scheduleRequest struct {
	Days []scheduleDayRequest `json:"days"`
}

func validClock(value string) bool { _, err := time.Parse("15:04", value); return err == nil }

func validScheduleDays(days []scheduleDayRequest) bool {
	if len(days) != 7 {
		return false
	}
	seen := map[int]bool{}
	for _, day := range days {
		if day.Weekday < 1 || day.Weekday > 7 || seen[day.Weekday] {
			return false
		}
		seen[day.Weekday] = true
		if !day.IsWorking {
			continue
		}
		if !validClock(day.Start) || !validClock(day.End) || day.End <= day.Start {
			return false
		}
		if (day.BreakFrom == "") != (day.BreakTo == "") {
			return false
		}
		if day.BreakFrom != "" && (!validClock(day.BreakFrom) || !validClock(day.BreakTo) || day.BreakTo <= day.BreakFrom || day.BreakFrom < day.Start || day.BreakTo > day.End) {
			return false
		}
	}
	return true
}

func (a *App) getDoctorSchedule(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeId")
	if _, err := uuid.Parse(employeeID); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid employee id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT weekday,is_working,COALESCE(to_char(start_time,'HH24:MI'),''),COALESCE(to_char(end_time,'HH24:MI'),''),COALESCE(to_char(break_start,'HH24:MI'),''),COALESCE(to_char(break_end,'HH24:MI'),'') FROM doctor_schedules WHERE organization_id=$1 AND employee_id=$2 ORDER BY weekday`, p.OrganizationID, employeeID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load doctor schedule", requestID(r))
		return
	}
	defer rows.Close()
	days := []scheduleDayRequest{}
	for rows.Next() {
		var day scheduleDayRequest
		if rows.Scan(&day.Weekday, &day.IsWorking, &day.Start, &day.End, &day.BreakFrom, &day.BreakTo) == nil {
			days = append(days, day)
		}
	}
	writeJSON(w, 200, map[string]any{"employeeId": employeeID, "configured": len(days) > 0, "days": days})
}

func (a *App) saveDoctorSchedule(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeId")
	if _, err := uuid.Parse(employeeID); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid employee id", requestID(r))
		return
	}
	var in scheduleRequest
	if !decode(w, r, &in) {
		return
	}
	for i := range in.Days {
		in.Days[i].Start = strings.TrimSpace(in.Days[i].Start)
		in.Days[i].End = strings.TrimSpace(in.Days[i].End)
		in.Days[i].BreakFrom = strings.TrimSpace(in.Days[i].BreakFrom)
		in.Days[i].BreakTo = strings.TrimSpace(in.Days[i].BreakTo)
	}
	if !validScheduleDays(in.Days) {
		errorJSON(w, 422, "validation_error", "Doctor schedule is invalid", requestID(r))
		return
	}
	sort.Slice(in.Days, func(i, j int) bool { return in.Days[i].Weekday < in.Days[j].Weekday })
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save doctor schedule", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branchID string
	err = tx.QueryRow(r.Context(), `SELECT e.branch_id FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id=$1 AND e.organization_id=$2 AND e.branch_id IS NOT NULL AND u.deleted_at IS NULL AND u.is_active=true`, employeeID, p.OrganizationID).Scan(&branchID)
	if err != nil {
		errorJSON(w, 404, "employee_not_found", "Active employee not found", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `DELETE FROM doctor_schedules WHERE organization_id=$1 AND employee_id=$2`, p.OrganizationID, employeeID); err != nil {
		errorJSON(w, 500, "database_error", "Could not replace doctor schedule", requestID(r))
		return
	}
	for _, day := range in.Days {
		var start, end, breakFrom, breakTo any
		if day.IsWorking {
			start, end = day.Start, day.End
			if day.BreakFrom != "" {
				breakFrom, breakTo = day.BreakFrom, day.BreakTo
			}
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO doctor_schedules(organization_id,branch_id,employee_id,weekday,is_working,start_time,end_time,break_start,break_end) VALUES($1,$2,$3,$4,$5,$6::time,$7::time,$8::time,$9::time)`, p.OrganizationID, branchID, employeeID, day.Weekday, day.IsWorking, start, end, breakFrom, breakTo)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not save schedule day", requestID(r))
			return
		}
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'DOCTOR_SCHEDULE_UPDATED','employee',$4,'success',jsonb_build_object('days',7))`, p.OrganizationID, branchID, p.UserID, employeeID)
	if err != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit doctor schedule", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]any{"employeeId": employeeID, "saved": true})
}
