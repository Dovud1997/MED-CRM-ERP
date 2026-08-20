package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type laboratoryTestRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Specimen    string `json:"specimen"`
	Description string `json:"description"`
	IsActive    *bool  `json:"isActive"`
	PriceUZS    int64  `json:"priceUzs"`
}

func (a *App) listLaboratoryTests(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT id,name,code,specimen,description,is_active,price_uzs FROM laboratory_test_catalog WHERE organization_id=$1 ORDER BY is_active DESC,name`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load laboratory tests", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name, specimen string
		var code, description *string
		var active bool
		var price int64
		if rows.Scan(&id, &name, &code, &specimen, &description, &active, &price) == nil {
			items = append(items, map[string]any{"id": id, "name": name, "code": code, "specimen": specimen, "description": description, "isActive": active, "priceUzs": price})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (a *App) createLaboratoryTest(w http.ResponseWriter, r *http.Request) {
	var in laboratoryTestRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	in.Specimen = strings.TrimSpace(in.Specimen)
	if len(in.Name) < 2 || len(in.Name) > 180 || len(in.Specimen) < 2 || in.PriceUZS < 0 {
		errorJSON(w, 422, "validation_error", "Invalid laboratory test", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	_, err := a.db.Exec(r.Context(), `INSERT INTO laboratory_test_catalog(id,organization_id,name,code,specimen,description,price_uzs) VALUES($1,$2,$3,NULLIF($4,''),$5,NULLIF($6,''),$7)`, id, p.OrganizationID, in.Name, in.Code, in.Specimen, strings.TrimSpace(in.Description), in.PriceUZS)
	if err != nil {
		errorJSON(w, 409, "test_exists", "Laboratory test already exists", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'LAB_TEST_CREATED','laboratory_test',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 201, map[string]string{"id": id})
}

func (a *App) updateLaboratoryTest(w http.ResponseWriter, r *http.Request) {
	var in laboratoryTestRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	in.Specimen = strings.TrimSpace(in.Specimen)
	if len(in.Name) < 2 || len(in.Specimen) < 2 || in.PriceUZS < 0 {
		errorJSON(w, 422, "validation_error", "Invalid laboratory test", requestID(r))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag, err := a.db.Exec(r.Context(), `UPDATE laboratory_test_catalog SET name=$1,code=NULLIF($2,''),specimen=$3,description=NULLIF($4,''),is_active=$5,price_uzs=$6,updated_at=now() WHERE id=$7 AND organization_id=$8`, in.Name, in.Code, in.Specimen, strings.TrimSpace(in.Description), active, in.PriceUZS, id, p.OrganizationID)
	if err != nil {
		errorJSON(w, 409, "test_exists", "Laboratory test already exists", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Laboratory test not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'LAB_TEST_UPDATED','laboratory_test',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]bool{"success": true})
}

type laboratoryOrderRequest struct {
	PatientID    string   `json:"patientId"`
	BranchID     string   `json:"branchId"`
	EmployeeID   string   `json:"employeeId"`
	Priority     string   `json:"priority"`
	ClinicalNote string   `json:"clinicalNote"`
	TestIDs      []string `json:"testIds"`
}

func (a *App) listLaboratoryOrders(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	rows, err := a.db.Query(r.Context(), `SELECT o.id,o.status,o.priority,o.clinical_note,o.result_note,o.source,o.created_at,o.completed_at,o.attachment_id,p.id,concat_ws(' ',p.last_name,p.first_name,p.middle_name),b.id,b.name,e.id,concat_ws(' ',e.last_name,e.first_name),COALESCE(jsonb_agg(jsonb_build_object('id',i.id,'testId',t.id,'name',t.name,'code',t.code,'specimen',t.specimen,'result',i.result_value,'unit',i.unit,'referenceRange',i.reference_range) ORDER BY t.name) FILTER(WHERE i.id IS NOT NULL),'[]') FROM laboratory_orders o JOIN patients p ON p.id=o.patient_id JOIN branches b ON b.id=o.branch_id LEFT JOIN employees e ON e.id=o.ordering_employee_id LEFT JOIN laboratory_order_items i ON i.order_id=o.id LEFT JOIN laboratory_test_catalog t ON t.id=i.test_id WHERE o.organization_id=$1 AND ($2='' OR o.status=$2) GROUP BY o.id,p.id,b.id,e.id ORDER BY CASE o.priority WHEN 'URGENT' THEN 0 ELSE 1 END,o.created_at DESC LIMIT 300`, p.OrganizationID, status)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load laboratory orders", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, statusValue, priority, source, patientID, patient, branchID, branch string
		var note, resultNote, employeeID, employee, attachmentID *string
		var created time.Time
		var completed *time.Time
		var raw []byte
		if rows.Scan(&id, &statusValue, &priority, &note, &resultNote, &source, &created, &completed, &attachmentID, &patientID, &patient, &branchID, &branch, &employeeID, &employee, &raw) == nil {
			items = append(items, map[string]any{"id": id, "status": statusValue, "priority": priority, "clinicalNote": note, "resultNote": resultNote, "source": source, "createdAt": created, "completedAt": completed, "attachmentId": attachmentID, "patient": map[string]any{"id": patientID, "name": patient}, "branch": map[string]any{"id": branchID, "name": branch}, "employee": map[string]any{"id": employeeID, "name": employee}, "items": json.RawMessage(raw)})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (a *App) createLaboratoryOrder(w http.ResponseWriter, r *http.Request) {
	var in laboratoryOrderRequest
	if !decode(w, r, &in) {
		return
	}
	if in.Priority == "" {
		in.Priority = "ROUTINE"
	}
	if (in.Priority != "ROUTINE" && in.Priority != "URGENT") || len(in.TestIDs) == 0 || len(in.TestIDs) > 30 {
		errorJSON(w, 422, "validation_error", "Invalid laboratory order", requestID(r))
		return
	}
	for _, id := range append([]string{in.PatientID, in.BranchID}, in.TestIDs...) {
		if _, err := uuid.Parse(id); err != nil {
			errorJSON(w, 422, "validation_error", "Invalid reference", requestID(r))
			return
		}
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create order", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var patientOK, branchOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM patients WHERE id=$1 AND organization_id=$3 AND deleted_at IS NULL),EXISTS(SELECT 1 FROM branches WHERE id=$2 AND organization_id=$3 AND deleted_at IS NULL)`, in.PatientID, in.BranchID, p.OrganizationID).Scan(&patientOK, &branchOK)
	if !patientOK || !branchOK {
		errorJSON(w, 422, "invalid_reference", "Patient or branch is invalid", requestID(r))
		return
	}
	var employee any = nil
	if strings.TrimSpace(in.EmployeeID) != "" {
		if _, e := uuid.Parse(in.EmployeeID); e != nil {
			errorJSON(w, 422, "invalid_employee", "Invalid doctor", requestID(r))
			return
		}
		var ok bool
		_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM employees WHERE id=$1 AND organization_id=$2)`, in.EmployeeID, p.OrganizationID).Scan(&ok)
		if !ok {
			errorJSON(w, 422, "invalid_employee", "Invalid doctor", requestID(r))
			return
		}
		employee = in.EmployeeID
	}
	id := uuid.NewString()
	_, err = tx.Exec(r.Context(), `INSERT INTO laboratory_orders(id,organization_id,branch_id,patient_id,ordering_employee_id,priority,clinical_note,requested_by) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8)`, id, p.OrganizationID, in.BranchID, in.PatientID, employee, in.Priority, strings.TrimSpace(in.ClinicalNote), p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create order", requestID(r))
		return
	}
	for _, testID := range in.TestIDs {
		tag, e := tx.Exec(r.Context(), `INSERT INTO laboratory_order_items(id,order_id,test_id) SELECT $1,$2,id FROM laboratory_test_catalog WHERE id=$3 AND organization_id=$4 AND is_active=true ON CONFLICT DO NOTHING`, uuid.NewString(), id, testID, p.OrganizationID)
		if e != nil || tag.RowsAffected() == 0 {
			errorJSON(w, 422, "invalid_test", "Laboratory test is invalid", requestID(r))
			return
		}
	}
	var chargeAmount int64
	var chargeDescription string
	_ = tx.QueryRow(r.Context(), `SELECT COALESCE(sum(t.price_uzs),0),'Лаборатория: '||string_agg(t.name,', ' ORDER BY t.name) FROM laboratory_order_items i JOIN laboratory_test_catalog t ON t.id=i.test_id WHERE i.order_id=$1`, id).Scan(&chargeAmount, &chargeDescription)
	if chargeAmount > 0 {
		_, err = tx.Exec(r.Context(), `INSERT INTO patient_charges(id,organization_id,branch_id,patient_id,source_type,source_id,description,amount_uzs,created_by) VALUES($1,$2,$3,$4,'LABORATORY',$5,$6,$7,$8)`, uuid.NewString(), p.OrganizationID, in.BranchID, in.PatientID, id, chargeDescription, chargeAmount, p.UserID)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not create laboratory charge", requestID(r))
			return
		}
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'LAB_ORDER_CREATED','laboratory_order',$4,'success',jsonb_build_object('patientId',$5::text,'priority',$6::text))`, p.OrganizationID, in.BranchID, p.UserID, id, in.PatientID, in.Priority)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit order", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]string{"id": id})
}

type laboratoryStatusRequest struct {
	Status string `json:"status"`
}

func (a *App) updateLaboratoryOrderStatus(w http.ResponseWriter, r *http.Request) {
	var in laboratoryStatusRequest
	if !decode(w, r, &in) {
		return
	}
	allowed := map[string]bool{"NEW": true, "SAMPLE_COLLECTED": true, "IN_PROGRESS": true, "CANCELLED": true}
	if !allowed[in.Status] {
		errorJSON(w, 422, "invalid_status", "Invalid status", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag, err := a.db.Exec(r.Context(), `UPDATE laboratory_orders SET status=$1,updated_at=now() WHERE id=$2 AND organization_id=$3 AND status<>'COMPLETED'`, in.Status, id, p.OrganizationID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 409, "status_conflict", "Could not update status", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'LAB_ORDER_STATUS_CHANGED','laboratory_order',$3,'success',jsonb_build_object('status',$4::text))`, p.OrganizationID, p.UserID, id, in.Status)
	writeJSON(w, 200, map[string]bool{"success": true})
}

type laboratoryResultItem struct {
	ItemID         string `json:"itemId"`
	Result         string `json:"result"`
	Unit           string `json:"unit"`
	ReferenceRange string `json:"referenceRange"`
}
type laboratoryResultRequest struct {
	ResultNote string                 `json:"resultNote"`
	Attachment labAttachmentRequest   `json:"attachment"`
	Items      []laboratoryResultItem `json:"items"`
}

func (a *App) completeLaboratoryOrder(w http.ResponseWriter, r *http.Request) {
	var in laboratoryResultRequest
	if !decode(w, r, &in) {
		return
	}
	data, ok := validateLabAttachment(in.Attachment)
	if !ok || len(in.Items) == 0 {
		errorJSON(w, 422, "validation_error", "Result file and values are required", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save result", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var patientID, branchID, status string
	err = tx.QueryRow(r.Context(), `SELECT patient_id,branch_id,status FROM laboratory_orders WHERE id=$1 AND organization_id=$2 FOR UPDATE`, id, p.OrganizationID).Scan(&patientID, &branchID, &status)
	if err != nil || status == "COMPLETED" || status == "CANCELLED" {
		errorJSON(w, 409, "order_closed", "Laboratory order is closed", requestID(r))
		return
	}
	attachmentID, err := saveLabAttachment(r, tx, patientID, in.Attachment, data, p.OrganizationID, p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save attachment", requestID(r))
		return
	}
	for _, item := range in.Items {
		item.Result = strings.TrimSpace(item.Result)
		if _, e := uuid.Parse(item.ItemID); e != nil || item.Result == "" {
			errorJSON(w, 422, "validation_error", "Every test requires a result", requestID(r))
			return
		}
		var testName, specimen string
		err = tx.QueryRow(r.Context(), `SELECT t.name,t.specimen FROM laboratory_order_items i JOIN laboratory_test_catalog t ON t.id=i.test_id WHERE i.id=$1 AND i.order_id=$2`, item.ItemID, id).Scan(&testName, &specimen)
		if err != nil {
			errorJSON(w, 422, "invalid_item", "Invalid laboratory item", requestID(r))
			return
		}
		kind := "blood"
		if strings.Contains(strings.ToLower(specimen), "моч") {
			kind = "urine"
		}
		resultID := uuid.NewString()
		_, err = tx.Exec(r.Context(), `INSERT INTO patient_lab_results(id,patient_id,organization_id,analysis_type,collected_on,test_name,result_value,unit,reference_range,notes,recorded_by,attachment_id) VALUES($1,$2,$3,$4,current_date,$5,$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),$10,$11)`, resultID, patientID, p.OrganizationID, kind, testName, item.Result, strings.TrimSpace(item.Unit), strings.TrimSpace(item.ReferenceRange), strings.TrimSpace(in.ResultNote), p.UserID, attachmentID)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not save patient result", requestID(r))
			return
		}
		_, err = tx.Exec(r.Context(), `UPDATE laboratory_order_items SET result_value=$1,unit=NULLIF($2,''),reference_range=NULLIF($3,''),patient_lab_result_id=$4 WHERE id=$5 AND order_id=$6`, item.Result, strings.TrimSpace(item.Unit), strings.TrimSpace(item.ReferenceRange), resultID, item.ItemID, id)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not update result", requestID(r))
			return
		}
	}
	_, err = tx.Exec(r.Context(), `UPDATE laboratory_orders SET status='COMPLETED',attachment_id=$1,result_note=NULLIF($2,''),completed_by=$3,completed_at=now(),updated_at=now() WHERE id=$4`, attachmentID, strings.TrimSpace(in.ResultNote), p.UserID, id)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not complete order", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'LAB_ORDER_COMPLETED','laboratory_order',$4,'success',jsonb_build_object('patientId',$5::text,'attachmentId',$6::text))`, p.OrganizationID, branchID, p.UserID, id, patientID, attachmentID)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit results", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]any{"success": true, "attachmentId": attachmentID})
}
