package app

import (
	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

func (a *App) getAccounting(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	branch := r.URL.Query().Get("branchId")
	if from == "" {
		from = time.Now().Format("2006-01") + "-01"
	}
	if to == "" {
		to = time.Now().AddDate(0, 1, 0).Format("2006-01") + "-01"
	}
	if _, e := time.Parse("2006-01-02", from); e != nil {
		errorJSON(w, 422, "invalid_period", "Invalid period", requestID(r))
		return
	}
	if _, e := time.Parse("2006-01-02", to); e != nil {
		errorJSON(w, 422, "invalid_period", "Invalid period", requestID(r))
		return
	}
	cats := []map[string]any{}
	rows, _ := a.db.Query(r.Context(), `SELECT id,name,entry_type,is_active FROM accounting_categories WHERE organization_id=$1 ORDER BY entry_type,name`, p.OrganizationID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id, name, kind string
			var active bool
			if rows.Scan(&id, &name, &kind, &active) == nil {
				cats = append(cats, map[string]any{"id": id, "name": name, "type": kind, "isActive": active})
			}
		}
	}
	entries := []map[string]any{}
	erows, _ := a.db.Query(r.Context(), `SELECT id,entry_type,amount_uzs,payment_method,occurred_on,description,document_number,counterparty,b.id,b.name,c.name,concat_ws(' ',e.last_name,e.first_name) FROM accounting_entries ae JOIN branches b ON b.id=ae.branch_id LEFT JOIN accounting_categories c ON c.id=ae.category_id LEFT JOIN employees e ON e.user_id=ae.created_by WHERE ae.organization_id=$1 AND occurred_on>=$2::date AND occurred_on<$3::date AND ($4='' OR ae.branch_id=NULLIF($4,'')::uuid) ORDER BY occurred_on DESC,ae.created_at DESC LIMIT 500`, p.OrganizationID, from, to, branch)
	if erows != nil {
		defer erows.Close()
		for erows.Next() {
			var id, kind, method, date, desc, bid, bname, creator string
			var amount int64
			var document, counterparty, category *string
			if erows.Scan(&id, &kind, &amount, &method, &date, &desc, &document, &counterparty, &bid, &bname, &category, &creator) == nil {
				entries = append(entries, map[string]any{"id": id, "type": kind, "amountUzs": amount, "method": method, "occurredOn": date, "description": desc, "documentNumber": document, "counterparty": counterparty, "branchId": bid, "branch": bname, "category": category, "createdBy": creator})
			}
		}
	}
	obligations := []map[string]any{}
	orows, _ := a.db.Query(r.Context(), `SELECT o.id,o.obligation_type,o.description,o.amount_uzs,o.paid_uzs,o.due_on,o.status,COALESCE(NULLIF(o.counterparty,''),concat_ws(' ',pt.last_name,pt.first_name,pt.middle_name)),b.id,b.name,e.id,concat_ws(' ',e.last_name,e.first_name) FROM accounting_obligations o JOIN branches b ON b.id=o.branch_id LEFT JOIN employees e ON e.id=o.employee_id LEFT JOIN patients pt ON pt.id=o.patient_id WHERE o.organization_id=$1 AND ($2='' OR o.branch_id=NULLIF($2,'')::uuid) ORDER BY (o.status IN ('OPEN','PARTIAL')) DESC,o.due_on NULLS LAST,o.created_at DESC LIMIT 300`, p.OrganizationID, branch)
	if orows != nil {
		defer orows.Close()
		for orows.Next() {
			var id, kind, desc, status, bid, bname string
			var amount, paid int64
			var due, counterparty, eid, employee *string
			if orows.Scan(&id, &kind, &desc, &amount, &paid, &due, &status, &counterparty, &bid, &bname, &eid, &employee) == nil {
				obligations = append(obligations, map[string]any{"id": id, "type": kind, "description": desc, "amountUzs": amount, "paidUzs": paid, "dueOn": due, "status": status, "counterparty": counterparty, "branchId": bid, "branch": bname, "employeeId": eid, "employee": employee})
			}
		}
	}
	payroll := []map[string]any{}
	prows, _ := a.db.Query(r.Context(), `SELECT pa.id,pa.period_month,pa.revenue_base_uzs,pa.fixed_amount_uzs,pa.percent_amount_uzs,pa.bonus_uzs,pa.deduction_uzs,pa.total_uzs,pa.paid_uzs,pa.status,pa.notes,e.id,concat_ws(' ',e.last_name,e.first_name),b.name FROM payroll_accruals pa JOIN employees e ON e.id=pa.employee_id LEFT JOIN branches b ON b.id=pa.branch_id WHERE pa.organization_id=$1 ORDER BY pa.period_month DESC,e.last_name LIMIT 300`, p.OrganizationID)
	if prows != nil {
		defer prows.Close()
		for prows.Next() {
			var id, month, status, eid, employee string
			var revenue, fixed, percent, bonus, deduction, total, paid int64
			var notes, branchName *string
			if prows.Scan(&id, &month, &revenue, &fixed, &percent, &bonus, &deduction, &total, &paid, &status, &notes, &eid, &employee, &branchName) == nil {
				payroll = append(payroll, map[string]any{"id": id, "periodMonth": month, "revenueBaseUzs": revenue, "fixedAmountUzs": fixed, "percentAmountUzs": percent, "bonusUzs": bonus, "deductionUzs": deduction, "totalUzs": total, "paidUzs": paid, "status": status, "notes": notes, "employeeId": eid, "employee": employee, "branch": branchName})
			}
		}
	}
	var manualIncome, manualExpense, cashIncome, cashOut, receivable, payable, payrollDebt int64
	_ = a.db.QueryRow(r.Context(), `SELECT COALESCE(sum(amount_uzs) FILTER(WHERE entry_type='INCOME'),0),COALESCE(sum(amount_uzs) FILTER(WHERE entry_type='EXPENSE'),0) FROM accounting_entries WHERE organization_id=$1 AND occurred_on>=$2::date AND occurred_on<$3::date AND ($4='' OR branch_id=NULLIF($4,'')::uuid)`, p.OrganizationID, from, to, branch).Scan(&manualIncome, &manualExpense)
	_ = a.db.QueryRow(r.Context(), `SELECT COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type='PAYMENT'),0),COALESCE(sum(amount_uzs) FILTER(WHERE transaction_type IN ('REFUND','EXPENSE')),0) FROM cash_transactions WHERE organization_id=$1 AND created_at>=$2::date AND created_at<$3::date AND ($4='' OR branch_id=NULLIF($4,'')::uuid)`, p.OrganizationID, from, to, branch).Scan(&cashIncome, &cashOut)
	_ = a.db.QueryRow(r.Context(), `SELECT COALESCE(sum(amount_uzs-paid_uzs) FILTER(WHERE obligation_type='RECEIVABLE' AND status<>'CANCELLED'),0),COALESCE(sum(amount_uzs-paid_uzs) FILTER(WHERE obligation_type='PAYABLE' AND status<>'CANCELLED'),0),COALESCE(sum(amount_uzs-paid_uzs) FILTER(WHERE obligation_type='EMPLOYEE' AND status<>'CANCELLED'),0) FROM accounting_obligations WHERE organization_id=$1`, p.OrganizationID).Scan(&receivable, &payable, &payrollDebt)
	writeJSON(w, 200, map[string]any{"categories": cats, "entries": entries, "obligations": obligations, "payroll": payroll, "summary": map[string]int64{"income": manualIncome + cashIncome, "expense": manualExpense + cashOut, "profit": manualIncome + cashIncome - manualExpense - cashOut, "receivable": receivable, "payable": payable, "employeeDebt": payrollDebt, "cashIncome": cashIncome}})
}

type accountingEntryInput struct {
	BranchID       string `json:"branchId"`
	CategoryID     string `json:"categoryId"`
	Type           string `json:"type"`
	Method         string `json:"method"`
	OccurredOn     string `json:"occurredOn"`
	Description    string `json:"description"`
	DocumentNumber string `json:"documentNumber"`
	Counterparty   string `json:"counterparty"`
	AmountUZS      int64  `json:"amountUzs"`
}

func (a *App) createAccountingEntry(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var in accountingEntryInput
	if !decode(w, r, &in) {
		return
	}
	in.Type = strings.ToUpper(in.Type)
	in.Method = strings.ToUpper(in.Method)
	if in.AmountUZS <= 0 || !cashOption(in.Type, "INCOME", "EXPENSE") || !cashOption(in.Method, "CASH", "CARD", "TRANSFER") || strings.TrimSpace(in.Description) == "" {
		errorJSON(w, 422, "validation_error", "Invalid entry", requestID(r))
		return
	}
	id := uuid.NewString()
	tag, e := a.db.Exec(r.Context(), `INSERT INTO accounting_entries(id,organization_id,branch_id,category_id,entry_type,amount_uzs,payment_method,occurred_on,description,document_number,counterparty,created_by) SELECT $1,$2,b.id,c.id,$5,$6,$7,$8::date,$9,NULLIF($10,''),NULLIF($11,''),$12 FROM branches b LEFT JOIN accounting_categories c ON c.id=NULLIF($4,'')::uuid AND c.organization_id=$2 WHERE b.id=$3 AND b.organization_id=$2`, id, p.OrganizationID, in.BranchID, in.CategoryID, in.Type, in.AmountUZS, in.Method, in.OccurredOn, strings.TrimSpace(in.Description), strings.TrimSpace(in.DocumentNumber), strings.TrimSpace(in.Counterparty), p.UserID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_reference", "Invalid branch or category", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'ACCOUNTING_ENTRY_CREATED','accounting_entry',$4,'success')`, p.OrganizationID, in.BranchID, p.UserID, id)
	writeJSON(w, 201, map[string]string{"id": id})
}

type obligationInput struct {
	BranchID     string `json:"branchId"`
	Type         string `json:"type"`
	EmployeeID   string `json:"employeeId"`
	Counterparty string `json:"counterparty"`
	Description  string `json:"description"`
	DueOn        string `json:"dueOn"`
	AmountUZS    int64  `json:"amountUzs"`
}

func (a *App) createObligation(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var in obligationInput
	if !decode(w, r, &in) {
		return
	}
	in.Type = strings.ToUpper(in.Type)
	if in.AmountUZS <= 0 || !cashOption(in.Type, "RECEIVABLE", "PAYABLE", "EMPLOYEE") || strings.TrimSpace(in.Description) == "" {
		errorJSON(w, 422, "validation_error", "Invalid obligation", requestID(r))
		return
	}
	id := uuid.NewString()
	tag, e := a.db.Exec(r.Context(), `INSERT INTO accounting_obligations(id,organization_id,branch_id,obligation_type,employee_id,counterparty,description,amount_uzs,due_on,created_by) SELECT $1,$2,b.id,$4,e.id,NULLIF($6,''),$7,$8,NULLIF($9,'')::date,$10 FROM branches b LEFT JOIN employees e ON e.id=NULLIF($5,'')::uuid AND e.organization_id=$2 WHERE b.id=$3 AND b.organization_id=$2`, id, p.OrganizationID, in.BranchID, in.Type, in.EmployeeID, strings.TrimSpace(in.Counterparty), strings.TrimSpace(in.Description), in.AmountUZS, in.DueOn, p.UserID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_reference", "Invalid obligation", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'ACCOUNTING_OBLIGATION_CREATED','accounting_obligation',$4,'success')`, p.OrganizationID, in.BranchID, p.UserID, id)
	writeJSON(w, 201, map[string]string{"id": id})
}

type paymentInput struct {
	AmountUZS int64 `json:"amountUzs"`
}

func (a *App) payObligation(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	var in paymentInput
	if !decode(w, r, &in) || in.AmountUZS <= 0 {
		return
	}
	tag, e := a.db.Exec(r.Context(), `UPDATE accounting_obligations SET paid_uzs=paid_uzs+$1,status=CASE WHEN paid_uzs+$1=amount_uzs THEN 'PAID' ELSE 'PARTIAL' END,updated_at=now() WHERE id=$2 AND organization_id=$3 AND status IN ('OPEN','PARTIAL') AND paid_uzs+$1<=amount_uzs`, in.AmountUZS, id, p.OrganizationID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_payment", "Payment exceeds remaining debt", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'ACCOUNTING_OBLIGATION_PAID','accounting_obligation',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]string{"id": id})
}

type payrollInput struct {
	BranchID       string  `json:"branchId"`
	EmployeeID     string  `json:"employeeId"`
	PeriodMonth    string  `json:"periodMonth"`
	Notes          string  `json:"notes"`
	RevenueBaseUZS int64   `json:"revenueBaseUzs"`
	FixedAmountUZS int64   `json:"fixedAmountUzs"`
	PercentRate    float64 `json:"percentRate"`
	BonusUZS       int64   `json:"bonusUzs"`
	DeductionUZS   int64   `json:"deductionUzs"`
}

func (a *App) createPayroll(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var in payrollInput
	if !decode(w, r, &in) {
		return
	}
	if in.PercentRate < 0 || in.PercentRate > 100 || in.FixedAmountUZS < 0 || in.RevenueBaseUZS < 0 || in.BonusUZS < 0 || in.DeductionUZS < 0 {
		errorJSON(w, 422, "validation_error", "Invalid payroll", requestID(r))
		return
	}
	percent := int64(float64(in.RevenueBaseUZS) * in.PercentRate / 100)
	total := in.FixedAmountUZS + percent + in.BonusUZS - in.DeductionUZS
	if total < 0 {
		errorJSON(w, 422, "invalid_total", "Payroll total cannot be negative", requestID(r))
		return
	}
	id := uuid.NewString()
	tag, e := a.db.Exec(r.Context(), `INSERT INTO payroll_accruals(id,organization_id,branch_id,employee_id,period_month,revenue_base_uzs,fixed_amount_uzs,percent_amount_uzs,bonus_uzs,deduction_uzs,total_uzs,notes,created_by) SELECT $1,$2,b.id,e.id,$5::date,$6,$7,$8,$9,$10,$11,NULLIF($12,''),$13 FROM employees e JOIN branches b ON b.id=$3 AND b.organization_id=$2 WHERE e.id=$4 AND e.organization_id=$2 ON CONFLICT(organization_id,employee_id,period_month) DO UPDATE SET revenue_base_uzs=EXCLUDED.revenue_base_uzs,fixed_amount_uzs=EXCLUDED.fixed_amount_uzs,percent_amount_uzs=EXCLUDED.percent_amount_uzs,bonus_uzs=EXCLUDED.bonus_uzs,deduction_uzs=EXCLUDED.deduction_uzs,total_uzs=EXCLUDED.total_uzs,notes=EXCLUDED.notes RETURNING id`, id, p.OrganizationID, in.BranchID, in.EmployeeID, in.PeriodMonth, in.RevenueBaseUZS, in.FixedAmountUZS, percent, in.BonusUZS, in.DeductionUZS, total, strings.TrimSpace(in.Notes), p.UserID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_reference", "Invalid employee or period", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'PAYROLL_ACCRUED','payroll_accrual',$4,'success')`, p.OrganizationID, in.BranchID, p.UserID, id)
	writeJSON(w, 201, map[string]any{"id": id, "totalUzs": total})
}
func (a *App) payPayroll(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	var in paymentInput
	if !decode(w, r, &in) || in.AmountUZS <= 0 {
		return
	}
	tag, e := a.db.Exec(r.Context(), `UPDATE payroll_accruals SET paid_uzs=paid_uzs+$1,status=CASE WHEN paid_uzs+$1=total_uzs THEN 'PAID' ELSE 'PARTIAL' END WHERE id=$2 AND organization_id=$3 AND paid_uzs+$1<=total_uzs`, in.AmountUZS, id, p.OrganizationID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_payment", "Payment exceeds payroll debt", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'PAYROLL_PAID','payroll_accrual',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]string{"id": id})
}
