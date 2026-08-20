package app

import (
	"clinicos/api/internal/auth"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strings"
	"time"
)

type cashShiftRequest struct {
	BranchID       string `json:"branchId"`
	OpeningCashUZS int64  `json:"openingCashUzs"`
	Notes          string `json:"notes"`
}
type cashCloseRequest struct {
	ClosingCashUZS int64  `json:"closingCashUzs"`
	Notes          string `json:"notes"`
}
type cashTransactionRequest struct {
	ShiftID              string `json:"shiftId"`
	PatientID            string `json:"patientId"`
	ServiceID            string `json:"serviceId"`
	RelatedTransactionID string `json:"relatedTransactionId"`
	Type                 string `json:"type"`
	Method               string `json:"method"`
	AmountUZS            int64  `json:"amountUzs"`
	Description          string `json:"description"`
	PaymentPurpose       string `json:"paymentPurpose"`
	Allocations          []struct {
		ChargeID string `json:"chargeId"`
		AmountUZS int64 `json:"amountUzs"`
	} `json:"allocations"`
}

func (a *App) getCashDesk(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	branch := r.URL.Query().Get("branchId")
	if branch != "" {
		if _, e := uuid.Parse(branch); e != nil {
			errorJSON(w, 422, "invalid_branch", "Invalid branch", requestID(r))
			return
		}
	}
	shifts := []map[string]any{}
	rows, e := a.db.Query(r.Context(), `SELECT s.id,s.status,s.opening_cash_uzs,s.closing_cash_uzs,s.expected_cash_uzs,s.variance_uzs,s.notes,s.opened_at,s.closed_at,b.id,b.name,concat_ws(' ',eo.last_name,eo.first_name),concat_ws(' ',ec.last_name,ec.first_name) FROM cash_shifts s JOIN branches b ON b.id=s.branch_id LEFT JOIN employees eo ON eo.user_id=s.opened_by LEFT JOIN employees ec ON ec.user_id=s.closed_by WHERE s.organization_id=$1 AND ($2='' OR s.branch_id=NULLIF($2,'')::uuid) ORDER BY (s.status='OPEN') DESC,s.opened_at DESC LIMIT 100`, p.OrganizationID, branch)
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not load shifts", requestID(r))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, status, bid, bname, opener, closer string
		var opening int64
		var closing, expected, variance *int64
		var notes *string
		var opened time.Time
		var closed *time.Time
		if rows.Scan(&id, &status, &opening, &closing, &expected, &variance, &notes, &opened, &closed, &bid, &bname, &opener, &closer) == nil {
			shifts = append(shifts, map[string]any{"id": id, "status": status, "openingCashUzs": opening, "closingCashUzs": closing, "expectedCashUzs": expected, "varianceUzs": variance, "notes": notes, "openedAt": opened, "closedAt": closed, "branchId": bid, "branch": bname, "openedBy": opener, "closedBy": closer})
		}
	}
	transactions := []map[string]any{}
	trs, e := a.db.Query(r.Context(), `SELECT t.id,t.shift_id,t.transaction_type,t.payment_method,t.payment_purpose,t.amount_uzs,t.receipt_number,t.description,t.created_at,t.related_transaction_id,p.id,concat_ws(' ',p.last_name,p.first_name,p.middle_name),s.id,s.name,b.id,b.name,concat_ws(' ',e.last_name,e.first_name) FROM cash_transactions t JOIN branches b ON b.id=t.branch_id LEFT JOIN patients p ON p.id=t.patient_id LEFT JOIN services s ON s.id=t.service_id LEFT JOIN employees e ON e.user_id=t.created_by WHERE t.organization_id=$1 AND ($2='' OR t.branch_id=NULLIF($2,'')::uuid) ORDER BY t.created_at DESC LIMIT 300`, p.OrganizationID, branch)
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not load transactions", requestID(r))
		return
	}
	defer trs.Close()
	for trs.Next() {
		var id, shift, kind, method, purpose, receipt, desc, bid, bname, cashier string
		var amount int64
		var created time.Time
		var related, pid, pname, sid, sname *string
		if trs.Scan(&id, &shift, &kind, &method, &purpose, &amount, &receipt, &desc, &created, &related, &pid, &pname, &sid, &sname, &bid, &bname, &cashier) == nil {
			transactions = append(transactions, map[string]any{"id": id, "shiftId": shift, "type": kind, "method": method, "paymentPurpose": purpose, "amountUzs": amount, "receiptNumber": receipt, "description": desc, "createdAt": created, "relatedTransactionId": related, "patientId": pid, "patient": pname, "serviceId": sid, "service": sname, "branchId": bid, "branch": bname, "cashier": cashier})
		}
	}
	var payments, cash, card, transfer, refunds, expenses int64
	_ = a.db.QueryRow(r.Context(), `SELECT COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='PAYMENT'),0),COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='PAYMENT' AND payment_method='CASH'),0),COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='PAYMENT' AND payment_method='CARD'),0),COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='PAYMENT' AND payment_method='TRANSFER'),0),COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='REFUND'),0),COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='EXPENSE'),0) FROM cash_transactions WHERE organization_id=$1 AND created_at>=date_trunc('day',now() AT TIME ZONE 'Asia/Tashkent') AT TIME ZONE 'Asia/Tashkent' AND ($2='' OR branch_id=NULLIF($2,'')::uuid)`, p.OrganizationID, branch).Scan(&payments, &cash, &card, &transfer, &refunds, &expenses)
	writeJSON(w, 200, map[string]any{"shifts": shifts, "transactions": transactions, "today": map[string]int64{"payments": payments, "cash": cash, "card": card, "transfer": transfer, "refunds": refunds, "expenses": expenses}})
}

func (a *App) openCashShift(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var in cashShiftRequest
	if !decode(w, r, &in) {
		return
	}
	if _, e := uuid.Parse(in.BranchID); e != nil || in.OpeningCashUZS < 0 {
		errorJSON(w, 422, "validation_error", "Invalid shift", requestID(r))
		return
	}
	id := uuid.NewString()
	tag, e := a.db.Exec(r.Context(), `INSERT INTO cash_shifts(id,organization_id,branch_id,opened_by,opening_cash_uzs,notes) SELECT $1,$2,id,$3,$4,NULLIF($5,'') FROM branches WHERE id=$6 AND organization_id=$2 AND deleted_at IS NULL`, id, p.OrganizationID, p.UserID, in.OpeningCashUZS, strings.TrimSpace(in.Notes), in.BranchID)
	if e != nil {
		if strings.Contains(e.Error(), "cash_shift_one_open_idx") {
			errorJSON(w, 409, "shift_already_open", "Shift already open", requestID(r))
			return
		}
		errorJSON(w, 500, "database_error", "Could not open shift", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "branch_not_found", "Branch not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'CASH_SHIFT_OPENED','cash_shift',$4,'success')`, p.OrganizationID, in.BranchID, p.UserID, id)
	writeJSON(w, 201, map[string]string{"id": id})
}

func (a *App) closeCashShift(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	var in cashCloseRequest
	if !decode(w, r, &in) {
		return
	}
	if _, e := uuid.Parse(id); e != nil || in.ClosingCashUZS < 0 {
		errorJSON(w, 422, "validation_error", "Invalid closing amount", requestID(r))
		return
	}
	tx, e := a.db.Begin(r.Context())
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not close shift", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branch string
	var opening int64
	e = tx.QueryRow(r.Context(), `SELECT branch_id,opening_cash_uzs FROM cash_shifts WHERE id=$1 AND organization_id=$2 AND status='OPEN' FOR UPDATE`, id, p.OrganizationID).Scan(&branch, &opening)
	if e == pgx.ErrNoRows {
		errorJSON(w, 409, "shift_not_open", "Shift is not open", requestID(r))
		return
	}
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not close shift", requestID(r))
		return
	}
	var movement int64
	_ = tx.QueryRow(r.Context(), `SELECT COALESCE(sum(CASE WHEN transaction_type='PAYMENT' AND payment_method='CASH' THEN amount_uzs WHEN transaction_type IN ('REFUND','EXPENSE') AND payment_method='CASH' THEN -amount_uzs ELSE 0 END),0) FROM cash_transactions WHERE shift_id=$1`, id).Scan(&movement)
	expected := opening + movement
	variance := in.ClosingCashUZS - expected
	_, e = tx.Exec(r.Context(), `UPDATE cash_shifts SET status='CLOSED',closed_by=$1,closing_cash_uzs=$2,expected_cash_uzs=$3,variance_uzs=$4,notes=concat_ws(E'\n',notes,NULLIF($5,'')),closed_at=now() WHERE id=$6`, p.UserID, in.ClosingCashUZS, expected, variance, strings.TrimSpace(in.Notes), id)
	if e == nil {
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'CASH_SHIFT_CLOSED','cash_shift',$4,'success')`, p.OrganizationID, branch, p.UserID, id)
	}
	if e != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not close shift", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]int64{"expectedCashUzs": expected, "closingCashUzs": in.ClosingCashUZS, "varianceUzs": variance})
}

func (a *App) createCashTransaction(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var in cashTransactionRequest
	if !decode(w, r, &in) {
		return
	}
	in.Type = strings.ToUpper(in.Type)
	in.Method = strings.ToUpper(in.Method)
	in.PaymentPurpose = strings.ToUpper(strings.TrimSpace(in.PaymentPurpose))
	if in.PaymentPurpose == "" {
		in.PaymentPurpose = "SERVICE"
	}
	in.Description = strings.TrimSpace(in.Description)
	if in.Type == "DEBT" {
		in.Method = "DEBT"
	}
	if _, e := uuid.Parse(in.ShiftID); e != nil || in.AmountUZS <= 0 || !cashOption(in.Type, "PAYMENT", "REFUND", "EXPENSE", "DEBT") || !cashOption(in.Method, "CASH", "CARD", "TRANSFER", "DEBT") || !cashOption(in.PaymentPurpose, "SERVICE", "LABORATORY", "MEDICINE", "SUPPLEMENT", "INPATIENT", "DIAGNOSTICS", "OTHER") || in.Description == "" {
		errorJSON(w, 422, "validation_error", "Invalid transaction", requestID(r))
		return
	}
	if (in.Type == "PAYMENT" || in.Type == "DEBT") && in.PatientID == "" {
		errorJSON(w, 422, "patient_required", "Patient required", requestID(r))
		return
	}
	tx, e := a.db.Begin(r.Context())
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not save transaction", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branch string
	if e = tx.QueryRow(r.Context(), `SELECT branch_id FROM cash_shifts WHERE id=$1 AND organization_id=$2 AND status='OPEN' FOR UPDATE`, in.ShiftID, p.OrganizationID).Scan(&branch); e != nil {
		errorJSON(w, 409, "shift_not_open", "Shift is not open", requestID(r))
		return
	}
	nullable := func(v string) (any, error) {
		if v == "" {
			return nil, nil
		}
		if _, e := uuid.Parse(v); e != nil {
			return nil, e
		}
		return v, nil
	}
	patient, e := nullable(in.PatientID)
	if e != nil {
		errorJSON(w, 422, "invalid_patient", "Invalid patient", requestID(r))
		return
	}
	service, e := nullable(in.ServiceID)
	if e != nil {
		errorJSON(w, 422, "invalid_service", "Invalid service", requestID(r))
		return
	}
	related, e := nullable(in.RelatedTransactionID)
	if e != nil {
		errorJSON(w, 422, "invalid_related", "Invalid payment", requestID(r))
		return
	}
	if in.Type == "REFUND" {
		if related == nil {
			errorJSON(w, 422, "related_payment_required", "Payment required", requestID(r))
			return
		}
		var original, returned int64
		e = tx.QueryRow(r.Context(), `SELECT amount_uzs,COALESCE((SELECT sum(r.amount_uzs) FROM cash_transactions r WHERE r.related_transaction_id=t.id AND r.transaction_type='REFUND'),0) FROM cash_transactions t WHERE t.id=$1 AND t.organization_id=$2 AND t.transaction_type='PAYMENT'`, related, p.OrganizationID).Scan(&original, &returned)
		if e != nil || returned+in.AmountUZS > original {
			errorJSON(w, 422, "refund_exceeds_payment", "Refund exceeds payment", requestID(r))
			return
		}
	}
	id := uuid.NewString()
	receipt := fmt.Sprintf("K-%s-%s", time.Now().In(time.FixedZone("UZT", 5*3600)).Format("060102"), strings.ToUpper(id[:8]))
	tag, e := tx.Exec(r.Context(), `INSERT INTO cash_transactions(id,organization_id,branch_id,shift_id,patient_id,service_id,related_transaction_id,transaction_type,payment_method,payment_purpose,amount_uzs,receipt_number,description,created_by) SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14 WHERE ($5::uuid IS NULL OR EXISTS(SELECT 1 FROM patients WHERE id=$5 AND organization_id=$2 AND deleted_at IS NULL)) AND ($6::uuid IS NULL OR EXISTS(SELECT 1 FROM services WHERE id=$6 AND organization_id=$2))`, id, p.OrganizationID, branch, in.ShiftID, patient, service, related, in.Type, in.Method, in.PaymentPurpose, in.AmountUZS, receipt, in.Description, p.UserID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_reference", "Patient or service not found", requestID(r))
		return
	}
	if in.Type == "PAYMENT" && len(in.Allocations) > 0 {
		var allocated int64
		for _, allocation := range in.Allocations {
			if _, parseErr := uuid.Parse(allocation.ChargeID); parseErr != nil || allocation.AmountUZS <= 0 {
				errorJSON(w, 422, "invalid_allocation", "Invalid charge allocation", requestID(r))
				return
			}
			allocationTag, allocationErr := tx.Exec(r.Context(), `UPDATE patient_charges SET paid_uzs=paid_uzs+$1,status=CASE WHEN paid_uzs+$1=amount_uzs THEN 'PAID' ELSE 'PARTIAL' END WHERE id=$2 AND organization_id=$3 AND patient_id=$4 AND status IN ('UNPAID','PARTIAL') AND paid_uzs+$1<=amount_uzs`, allocation.AmountUZS, allocation.ChargeID, p.OrganizationID, patient)
			if allocationErr != nil || allocationTag.RowsAffected() == 0 {
				errorJSON(w, 422, "invalid_allocation", "Charge is already paid or amount is too high", requestID(r))
				return
			}
			_, e = tx.Exec(r.Context(), `INSERT INTO cash_payment_allocations(transaction_id,charge_id,amount_uzs) VALUES($1,$2,$3)`, id, allocation.ChargeID, allocation.AmountUZS)
			if e != nil {
				errorJSON(w, 422, "invalid_allocation", "Could not allocate payment", requestID(r))
				return
			}
			allocated += allocation.AmountUZS
		}
		if allocated != in.AmountUZS {
			errorJSON(w, 422, "allocation_total_mismatch", "Payment total does not match selected charges", requestID(r))
			return
		}
	}
	_, e = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'CASH_TRANSACTION_CREATED','cash_transaction',$4,'success')`, p.OrganizationID, branch, p.UserID, id)
	if e == nil && in.Type == "DEBT" {
		_, e = tx.Exec(r.Context(), `INSERT INTO accounting_obligations(id,organization_id,branch_id,obligation_type,patient_id,source_transaction_id,counterparty,description,amount_uzs,due_on,created_by) SELECT $1,$2,$3,'RECEIVABLE',p.id,$4,concat_ws(' ',p.last_name,p.first_name,p.middle_name),$5,$6,current_date+30,$7 FROM patients p WHERE p.id=$8 AND p.organization_id=$2 AND p.deleted_at IS NULL`, uuid.NewString(), p.OrganizationID, branch, id, in.Description, in.AmountUZS, p.UserID, patient)
	}
	if e != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not save transaction", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]string{"id": id, "receiptNumber": receipt})
}
func cashOption(value string, values ...string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}
