package app

import (
	"clinicos/api/internal/auth"
	"github.com/google/uuid"
	"net/http"
)

func (a *App) getTodayPatientDebts(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT c.id,c.description,c.amount_uzs-c.paid_uzs,c.source_type,c.charged_at,p.id,concat_ws(' ',p.last_name,p.first_name,p.middle_name),pgp_sym_decrypt(p.phone_encrypted,$2),CASE WHEN p.guardian_phone_encrypted IS NULL THEN NULL ELSE pgp_sym_decrypt(p.guardian_phone_encrypted,$2) END,p.telegram_url FROM patient_charges c JOIN patients p ON p.id=c.patient_id WHERE c.organization_id=$1 AND c.status IN ('UNPAID','PARTIAL') AND c.charged_at>=date_trunc('day',now() AT TIME ZONE 'Asia/Tashkent') AT TIME ZONE 'Asia/Tashkent' AND c.charged_at<(date_trunc('day',now() AT TIME ZONE 'Asia/Tashkent')+interval '1 day') AT TIME ZONE 'Asia/Tashkent' ORDER BY p.last_name,p.first_name,c.charged_at`, p.OrganizationID, a.cfg.RefreshSecret)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load today debts", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, description, source, patientID, patient, phone string
		var guardianPhone, telegram *string
		var remaining int64
		var chargedAt any
		if rows.Scan(&id, &description, &remaining, &source, &chargedAt, &patientID, &patient, &phone, &guardianPhone, &telegram) == nil {
			items = append(items, map[string]any{"id": id, "description": description, "remainingUzs": remaining, "sourceType": source, "chargedAt": chargedAt, "patientId": patientID, "patient": patient, "phone": phone, "guardianPhone": guardianPhone, "telegram": telegram})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (a *App) getPatientAccount(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	patientID := r.URL.Query().Get("patientId")
	if _, err := uuid.Parse(patientID); err != nil {
		errorJSON(w, 422, "invalid_patient", "Invalid patient", requestID(r))
		return
	}
	var patient string
	if err := a.db.QueryRow(r.Context(), `SELECT concat_ws(' ',last_name,first_name,middle_name) FROM patients WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL`, patientID, p.OrganizationID).Scan(&patient); err != nil {
		errorJSON(w, 404, "patient_not_found", "Patient not found", requestID(r))
		return
	}
	rows, err := a.db.Query(r.Context(), `SELECT id,source_type,description,amount_uzs,paid_uzs,status,charged_at,branch_id FROM patient_charges WHERE organization_id=$1 AND patient_id=$2 AND status IN ('UNPAID','PARTIAL') ORDER BY charged_at`, p.OrganizationID, patientID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load patient account", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	var total int64
	for rows.Next() {
		var id, source, description, status, branchID string
		var amount, paid int64
		var chargedAt any
		if rows.Scan(&id, &source, &description, &amount, &paid, &status, &chargedAt, &branchID) == nil {
			remaining := amount - paid
			total += remaining
			items = append(items, map[string]any{"id": id, "sourceType": source, "description": description, "amountUzs": amount, "paidUzs": paid, "remainingUzs": remaining, "status": status, "chargedAt": chargedAt, "branchId": branchID})
		}
	}
	writeJSON(w, 200, map[string]any{"patientId": patientID, "patient": patient, "items": items, "totalDueUzs": total})
}
