package app

import (
	"clinicos/api/internal/auth"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type assetRequest struct {
	ItemID                string `json:"itemId"`
	BranchID              string `json:"branchId"`
	RoomID                string `json:"roomId"`
	InventoryNumber       string `json:"inventoryNumber"`
	SerialNumber          string `json:"serialNumber"`
	Status                string `json:"status"`
	AssignedEmployeeID    string `json:"assignedEmployeeId"`
	ResponsibleEmployeeID string `json:"responsibleEmployeeId"`
	AcquiredOn            string `json:"acquiredOn"`
	WarrantyUntil         string `json:"warrantyUntil"`
	Notes                 string `json:"notes"`
}

func (a *App) listInventoryAssets(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT a.id,a.inventory_number,a.serial_number,a.status,a.acquired_on,a.warranty_until,a.notes,i.id,i.name,b.id,b.name,rm.id,CASE WHEN rm.id IS NULL THEN NULL ELSE concat('Палата №',rm.number,CASE WHEN rm.name='' THEN '' ELSE ' — '||rm.name END) END,ae.id,concat_ws(' ',ae.last_name,ae.first_name),re.id,concat_ws(' ',re.last_name,re.first_name) FROM inventory_assets a JOIN inventory_items i ON i.id=a.item_id JOIN branches b ON b.id=a.branch_id LEFT JOIN inpatient_rooms rm ON rm.id=a.room_id LEFT JOIN employees ae ON ae.id=a.assigned_employee_id LEFT JOIN employees re ON re.id=a.responsible_employee_id WHERE a.organization_id=$1 ORDER BY b.name,rm.number NULLS LAST,a.status,a.inventory_number`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load equipment", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, number, status, itemID, item, branchID, branch string
		var serial, notes, roomID, room, assignedID, assigned, responsibleID, responsible *string
		var acquired, warranty *time.Time
		if rows.Scan(&id, &number, &serial, &status, &acquired, &warranty, &notes, &itemID, &item, &branchID, &branch, &roomID, &room, &assignedID, &assigned, &responsibleID, &responsible) == nil {
			items = append(items, map[string]any{"id": id, "inventoryNumber": number, "serialNumber": serial, "status": status, "acquiredOn": dateString(acquired), "warrantyUntil": dateString(warranty), "notes": notes, "itemId": itemID, "item": item, "branchId": branchID, "branch": branch, "roomId": roomID, "room": room, "assignedEmployeeId": assignedID, "assignedEmployee": assigned, "responsibleEmployeeId": responsibleID, "responsibleEmployee": responsible})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func dateString(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.Format("2006-01-02")
}
func (a *App) saveInventoryAsset(w http.ResponseWriter, r *http.Request) {
	var in assetRequest
	if !decode(w, r, &in) {
		return
	}
	if in.Status == "" {
		in.Status = "AVAILABLE"
	}
	allowed := map[string]bool{"AVAILABLE": true, "IN_USE": true, "MAINTENANCE": true, "RETIRED": true}
	if !allowed[in.Status] || strings.TrimSpace(in.InventoryNumber) == "" {
		errorJSON(w, 422, "validation_error", "Invalid equipment", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	if in.RoomID != "" {
		var roomValid bool
		_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM inpatient_rooms WHERE id=$1 AND organization_id=$2 AND branch_id=$3 AND is_active=true)`, in.RoomID, p.OrganizationID, in.BranchID).Scan(&roomValid)
		if !roomValid {
			errorJSON(w, 422, "invalid_room", "Room does not belong to the selected branch", requestID(r))
			return
		}
	}
	for _, employeeID := range []string{in.AssignedEmployeeID, in.ResponsibleEmployeeID} {
		if employeeID == "" {
			continue
		}
		var valid bool
		_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id=$1 AND e.organization_id=$2 AND u.is_active=true AND u.deleted_at IS NULL)`, employeeID, p.OrganizationID).Scan(&valid)
		if !valid {
			errorJSON(w, 422, "invalid_employee", "Invalid employee", requestID(r))
			return
		}
	}
	var acquired, warranty any
	if in.AcquiredOn != "" {
		d, e := time.Parse("2006-01-02", in.AcquiredOn)
		if e != nil {
			errorJSON(w, 422, "validation_error", "Invalid date", requestID(r))
			return
		}
		acquired = d
	}
	if in.WarrantyUntil != "" {
		d, e := time.Parse("2006-01-02", in.WarrantyUntil)
		if e != nil {
			errorJSON(w, 422, "validation_error", "Invalid warranty", requestID(r))
			return
		}
		warranty = d
	}
	var room, assigned, responsible any
	if in.RoomID != "" {
		room = in.RoomID
	}
	if in.AssignedEmployeeID != "" {
		assigned = in.AssignedEmployeeID
	}
	if in.ResponsibleEmployeeID != "" {
		responsible = in.ResponsibleEmployeeID
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		id = uuid.NewString()
		tag, e := a.db.Exec(r.Context(), `INSERT INTO inventory_assets(id,organization_id,branch_id,room_id,item_id,inventory_number,serial_number,status,assigned_employee_id,responsible_employee_id,acquired_on,warranty_until,notes) SELECT $1,$2,b.id,$5,i.id,$6,NULLIF($7,''),$8,$9,$10,$11,$12,NULLIF($13,'') FROM branches b JOIN inventory_items i ON i.id=$4 AND i.organization_id=$2 AND i.category='EQUIPMENT' WHERE b.id=$3 AND b.organization_id=$2`, id, p.OrganizationID, in.BranchID, in.ItemID, room, strings.TrimSpace(in.InventoryNumber), strings.TrimSpace(in.SerialNumber), in.Status, assigned, responsible, acquired, warranty, strings.TrimSpace(in.Notes))
		if e != nil || tag.RowsAffected() == 0 {
			errorJSON(w, 409, "asset_exists", "Equipment number exists or references are invalid", requestID(r))
			return
		}
	} else {
		tag, e := a.db.Exec(r.Context(), `UPDATE inventory_assets SET branch_id=$1,room_id=$2,item_id=$3,inventory_number=$4,serial_number=NULLIF($5,''),status=$6,assigned_employee_id=$7,responsible_employee_id=$8,acquired_on=$9,warranty_until=$10,notes=NULLIF($11,''),updated_at=now() WHERE id=$12 AND organization_id=$13`, in.BranchID, room, in.ItemID, strings.TrimSpace(in.InventoryNumber), strings.TrimSpace(in.SerialNumber), in.Status, assigned, responsible, acquired, warranty, strings.TrimSpace(in.Notes), id, p.OrganizationID)
		if e != nil || tag.RowsAffected() == 0 {
			errorJSON(w, 409, "asset_update_failed", "Could not update equipment", requestID(r))
			return
		}
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'INVENTORY_ASSET_SAVED','inventory_asset',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]string{"id": id})
}

type countLineRequest struct {
	LineID         string  `json:"lineId"`
	ActualQuantity float64 `json:"actualQuantity"`
}
type completeCountRequest struct {
	Lines []countLineRequest `json:"lines"`
	Notes string             `json:"notes"`
}

func (a *App) listInventoryCounts(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT c.id,c.status,c.started_at,c.completed_at,c.notes,b.id,b.name,COALESCE(jsonb_agg(jsonb_build_object('id',l.id,'batchId',bt.id,'item',i.name,'unit',i.unit,'lotNumber',bt.lot_number,'expiresOn',bt.expires_on,'expectedQuantity',l.expected_quantity,'actualQuantity',l.actual_quantity,'difference',l.difference) ORDER BY i.name,bt.lot_number) FILTER(WHERE l.id IS NOT NULL),'[]') FROM inventory_counts c JOIN branches b ON b.id=c.branch_id LEFT JOIN inventory_count_lines l ON l.count_id=c.id LEFT JOIN inventory_batches bt ON bt.id=l.batch_id LEFT JOIN inventory_items i ON i.id=bt.item_id WHERE c.organization_id=$1 GROUP BY c.id,b.id ORDER BY c.started_at DESC LIMIT 100`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load counts", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, status, branchID, branch string
		var started time.Time
		var completed *time.Time
		var notes *string
		var raw []byte
		if rows.Scan(&id, &status, &started, &completed, &notes, &branchID, &branch, &raw) == nil {
			items = append(items, map[string]any{"id": id, "status": status, "startedAt": started, "completedAt": completed, "notes": notes, "branchId": branchID, "branch": branch, "lines": json.RawMessage(raw)})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (a *App) createInventoryCount(w http.ResponseWriter, r *http.Request) {
	var in struct {
		BranchID string `json:"branchId"`
	}
	if !decode(w, r, &in) {
		return
	}
	if _, e := uuid.Parse(in.BranchID); e != nil {
		errorJSON(w, 422, "invalid_branch", "Invalid branch", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, e := a.db.Begin(r.Context())
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not start count", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var active bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM inventory_counts WHERE organization_id=$1 AND branch_id=$2 AND status='DRAFT')`, p.OrganizationID, in.BranchID).Scan(&active)
	if active {
		errorJSON(w, 409, "count_in_progress", "An inventory count is already in progress", requestID(r))
		return
	}
	id := uuid.NewString()
	tag, e := tx.Exec(r.Context(), `INSERT INTO inventory_counts(id,organization_id,branch_id,started_by) SELECT $1,$2,id,$3 FROM branches WHERE id=$4 AND organization_id=$2`, id, p.OrganizationID, p.UserID, in.BranchID)
	if e != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 422, "invalid_branch", "Invalid branch", requestID(r))
		return
	}
	_, e = tx.Exec(r.Context(), `INSERT INTO inventory_count_lines(id,count_id,batch_id,expected_quantity) SELECT gen_random_uuid(),$1,id,quantity FROM inventory_batches WHERE organization_id=$2 AND branch_id=$3 AND quantity>0`, id, p.OrganizationID, in.BranchID)
	if e != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not create count", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]string{"id": id})
}
func (a *App) completeInventoryCount(w http.ResponseWriter, r *http.Request) {
	var in completeCountRequest
	if !decode(w, r, &in) {
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tx, e := a.db.Begin(r.Context())
	if e != nil {
		errorJSON(w, 500, "database_error", "Could not complete count", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branchID, status string
	e = tx.QueryRow(r.Context(), `SELECT branch_id,status FROM inventory_counts WHERE id=$1 AND organization_id=$2 FOR UPDATE`, id, p.OrganizationID).Scan(&branchID, &status)
	if e != nil || status != "DRAFT" {
		errorJSON(w, 409, "count_closed", "Inventory count is closed", requestID(r))
		return
	}
	for _, line := range in.Lines {
		if line.ActualQuantity < 0 {
			errorJSON(w, 422, "invalid_quantity", "Actual quantity cannot be negative", requestID(r))
			return
		}
		var batchID, itemID string
		var expected float64
		e = tx.QueryRow(r.Context(), `SELECT l.batch_id,l.expected_quantity,b.item_id FROM inventory_count_lines l JOIN inventory_batches b ON b.id=l.batch_id WHERE l.id=$1 AND l.count_id=$2 FOR UPDATE`, line.LineID, id).Scan(&batchID, &expected, &itemID)
		if e != nil {
			errorJSON(w, 422, "invalid_line", "Invalid count line", requestID(r))
			return
		}
		diff := line.ActualQuantity - expected
		_, e = tx.Exec(r.Context(), `UPDATE inventory_count_lines SET actual_quantity=$1,difference=$2 WHERE id=$3`, line.ActualQuantity, diff, line.LineID)
		if e == nil && diff != 0 {
			_, e = tx.Exec(r.Context(), `UPDATE inventory_batches SET quantity=$1,updated_at=now() WHERE id=$2`, line.ActualQuantity, batchID)
			if e == nil {
				qty := diff
				if qty < 0 {
					qty = -qty
				}
				_, e = tx.Exec(r.Context(), `INSERT INTO inventory_movements(id,organization_id,branch_id,item_id,batch_id,movement_type,quantity,reference_id,reason,performed_by) VALUES($1,$2,$3,$4,$5,'ADJUSTMENT',$6,$7,$8,$9)`, uuid.NewString(), p.OrganizationID, branchID, itemID, batchID, qty, id, "Инвентаризация: расхождение", p.UserID)
			}
		}
		if e != nil {
			errorJSON(w, 500, "database_error", "Could not save count line", requestID(r))
			return
		}
	}
	_, e = tx.Exec(r.Context(), `UPDATE inventory_counts SET status='COMPLETED',completed_by=$1,completed_at=now(),notes=NULLIF($2,'') WHERE id=$3`, p.UserID, strings.TrimSpace(in.Notes), id)
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,$3,'INVENTORY_COUNT_COMPLETED','inventory_count',$4,'success')`, p.OrganizationID, branchID, p.UserID, id)
	if e != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit count", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}
