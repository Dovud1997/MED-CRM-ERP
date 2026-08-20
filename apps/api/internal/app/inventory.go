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

type inventoryItemRequest struct {
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Category     string  `json:"category"`
	Unit         string  `json:"unit"`
	Manufacturer string  `json:"manufacturer"`
	Barcode      string  `json:"barcode"`
	MinStock     float64 `json:"minStock"`
	IsActive     *bool   `json:"isActive"`
}

func (a *App) listInventoryItems(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	branch := r.URL.Query().Get("branchId")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := a.db.Query(r.Context(), `SELECT i.id,i.sku,i.name,i.category,i.unit,i.manufacturer,i.barcode,i.min_stock,i.is_active,COALESCE(sum(b.quantity) FILTER(WHERE $2='' OR b.branch_id::text=$2),0),COALESCE(jsonb_agg(jsonb_build_object('id',b.id,'branchId',b.branch_id,'branch',br.name,'lotNumber',b.lot_number,'expiresOn',b.expires_on,'quantity',b.quantity,'purchasePriceUzs',b.purchase_price_uzs,'responsibleEmployeeId',b.responsible_employee_id,'responsibleEmployee',concat_ws(' ',re.last_name,re.first_name)) ORDER BY b.expires_on NULLS LAST) FILTER(WHERE b.id IS NOT NULL AND ($2='' OR b.branch_id::text=$2)),'[]') FROM inventory_items i LEFT JOIN inventory_batches b ON b.item_id=i.id LEFT JOIN branches br ON br.id=b.branch_id LEFT JOIN employees re ON re.id=b.responsible_employee_id WHERE i.organization_id=$1 AND ($3='' OR i.name ILIKE '%'||$3||'%' OR i.sku ILIKE '%'||$3||'%' OR COALESCE(i.barcode,'') ILIKE '%'||$3||'%') GROUP BY i.id ORDER BY i.is_active DESC,i.name`, p.OrganizationID, branch, q)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load inventory", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, sku, name, category, unit string
		var manufacturer, barcode *string
		var min, total float64
		var active bool
		var raw []byte
		if rows.Scan(&id, &sku, &name, &category, &unit, &manufacturer, &barcode, &min, &active, &total, &raw) == nil {
			items = append(items, map[string]any{"id": id, "sku": sku, "name": name, "category": category, "unit": unit, "manufacturer": manufacturer, "barcode": barcode, "minStock": min, "isActive": active, "totalStock": total, "batches": json.RawMessage(raw)})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (a *App) createInventoryItem(w http.ResponseWriter, r *http.Request) {
	var in inventoryItemRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	in.Unit = strings.TrimSpace(in.Unit)
	if in.Category == "" {
		in.Category = "MEDICINE"
	}
	if len(in.Name) < 2 || len(in.SKU) < 1 || len(in.Unit) < 1 || in.MinStock < 0 {
		errorJSON(w, 422, "validation_error", "Invalid inventory item", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	id := uuid.NewString()
	_, err := a.db.Exec(r.Context(), `INSERT INTO inventory_items(id,organization_id,sku,name,category,unit,manufacturer,barcode,min_stock) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),$9)`, id, p.OrganizationID, in.SKU, in.Name, in.Category, in.Unit, strings.TrimSpace(in.Manufacturer), strings.TrimSpace(in.Barcode), in.MinStock)
	if err != nil {
		errorJSON(w, 409, "sku_exists", "SKU already exists", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'INVENTORY_ITEM_CREATED','inventory_item',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 201, map[string]string{"id": id})
}
func (a *App) updateInventoryItem(w http.ResponseWriter, r *http.Request) {
	var in inventoryItemRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	in.Unit = strings.TrimSpace(in.Unit)
	if len(in.Name) < 2 || len(in.SKU) < 1 || len(in.Unit) < 1 || in.MinStock < 0 {
		errorJSON(w, 422, "validation_error", "Invalid inventory item", requestID(r))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	p := auth.PrincipalFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag, err := a.db.Exec(r.Context(), `UPDATE inventory_items SET sku=$1,name=$2,category=$3,unit=$4,manufacturer=NULLIF($5,''),barcode=NULLIF($6,''),min_stock=$7,is_active=$8,updated_at=now() WHERE id=$9 AND organization_id=$10`, in.SKU, in.Name, in.Category, in.Unit, strings.TrimSpace(in.Manufacturer), strings.TrimSpace(in.Barcode), in.MinStock, active, id, p.OrganizationID)
	if err != nil {
		errorJSON(w, 409, "sku_exists", "SKU already exists", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Item not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'INVENTORY_ITEM_UPDATED','inventory_item',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]bool{"success": true})
}

type inventoryMovementRequest struct {
	ItemID                string  `json:"itemId"`
	BranchID              string  `json:"branchId"`
	ToBranchID            string  `json:"toBranchId"`
	BatchID               string  `json:"batchId"`
	Type                  string  `json:"type"`
	Quantity              float64 `json:"quantity"`
	LotNumber             string  `json:"lotNumber"`
	ExpiresOn             string  `json:"expiresOn"`
	PurchasePriceUZS      int64   `json:"purchasePriceUzs"`
	Reason                string  `json:"reason"`
	ResponsibleEmployeeID string  `json:"responsibleEmployeeId"`
}

func (a *App) createInventoryMovement(w http.ResponseWriter, r *http.Request) {
	var in inventoryMovementRequest
	if !decode(w, r, &in) {
		return
	}
	allowed := map[string]bool{"RECEIPT": true, "ISSUE": true, "WRITE_OFF": true, "TRANSFER": true}
	if !allowed[in.Type] || in.Quantity <= 0 || in.Quantity > 100000000 {
		errorJSON(w, 422, "validation_error", "Invalid stock movement", requestID(r))
		return
	}
	for _, id := range []string{in.ItemID, in.BranchID} {
		if _, e := uuid.Parse(id); e != nil {
			errorJSON(w, 422, "validation_error", "Invalid reference", requestID(r))
			return
		}
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not start movement", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var refs bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM inventory_items i JOIN branches b ON b.id=$2 AND b.organization_id=$3 WHERE i.id=$1 AND i.organization_id=$3 AND i.is_active=true)`, in.ItemID, in.BranchID, p.OrganizationID).Scan(&refs)
	if !refs {
		errorJSON(w, 422, "invalid_reference", "Item or branch is invalid", requestID(r))
		return
	}
	reference := uuid.NewString()
	reason := strings.TrimSpace(in.Reason)
	if in.Type == "RECEIPT" {
		lot := strings.TrimSpace(in.LotNumber)
		if lot == "" {
			errorJSON(w, 422, "lot_required", "Lot number is required", requestID(r))
			return
		}
		var expires any = nil
		if in.ExpiresOn != "" {
			date, e := time.Parse("2006-01-02", in.ExpiresOn)
			if e != nil {
				errorJSON(w, 422, "invalid_expiry", "Invalid expiry date", requestID(r))
				return
			}
			expires = date
		}
		batchID := uuid.NewString()
		var responsible any = nil
		if strings.TrimSpace(in.ResponsibleEmployeeID) != "" {
			var employeeOK bool
			_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM employees WHERE id=$1 AND organization_id=$2 AND is_active=true)`, in.ResponsibleEmployeeID, p.OrganizationID).Scan(&employeeOK)
			if !employeeOK {
				errorJSON(w, 422, "invalid_responsible", "Invalid responsible employee", requestID(r))
				return
			}
			responsible = in.ResponsibleEmployeeID
		}
		err = tx.QueryRow(r.Context(), `INSERT INTO inventory_batches(id,organization_id,branch_id,item_id,lot_number,expires_on,quantity,purchase_price_uzs,responsible_employee_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(organization_id,branch_id,item_id,lot_number) DO UPDATE SET quantity=inventory_batches.quantity+EXCLUDED.quantity,expires_on=COALESCE(EXCLUDED.expires_on,inventory_batches.expires_on),purchase_price_uzs=EXCLUDED.purchase_price_uzs,responsible_employee_id=COALESCE(EXCLUDED.responsible_employee_id,inventory_batches.responsible_employee_id),updated_at=now() RETURNING id`, batchID, p.OrganizationID, in.BranchID, in.ItemID, lot, expires, in.Quantity, in.PurchasePriceUZS, responsible).Scan(&batchID)
		if err == nil {
			_, err = tx.Exec(r.Context(), `INSERT INTO inventory_movements(id,organization_id,branch_id,item_id,batch_id,movement_type,quantity,reference_id,reason,performed_by) VALUES($1,$2,$3,$4,$5,'RECEIPT',$6,$7,NULLIF($8,''),$9)`, uuid.NewString(), p.OrganizationID, in.BranchID, in.ItemID, batchID, in.Quantity, reference, reason, p.UserID)
		}
	} else {
		if _, e := uuid.Parse(in.BatchID); e != nil {
			errorJSON(w, 422, "batch_required", "Select a batch", requestID(r))
			return
		}
		var qty float64
		err = tx.QueryRow(r.Context(), `SELECT quantity FROM inventory_batches WHERE id=$1 AND item_id=$2 AND branch_id=$3 AND organization_id=$4 FOR UPDATE`, in.BatchID, in.ItemID, in.BranchID, p.OrganizationID).Scan(&qty)
		if err != nil || qty < in.Quantity {
			errorJSON(w, 409, "insufficient_stock", "Insufficient stock in selected batch", requestID(r))
			return
		}
		_, err = tx.Exec(r.Context(), `UPDATE inventory_batches SET quantity=quantity-$1,updated_at=now() WHERE id=$2`, in.Quantity, in.BatchID)
		if err == nil && in.Type == "TRANSFER" {
			if _, e := uuid.Parse(in.ToBranchID); e != nil || in.ToBranchID == in.BranchID {
				errorJSON(w, 422, "invalid_destination", "Invalid destination branch", requestID(r))
				return
			}
			var destination bool
			_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, in.ToBranchID, p.OrganizationID).Scan(&destination)
			if !destination {
				errorJSON(w, 422, "invalid_destination", "Invalid destination branch", requestID(r))
				return
			}
			var lot string
			var expires *time.Time
			var price int64
			_ = tx.QueryRow(r.Context(), `SELECT lot_number,expires_on,purchase_price_uzs FROM inventory_batches WHERE id=$1`, in.BatchID).Scan(&lot, &expires, &price)
			newBatch := uuid.NewString()
			err = tx.QueryRow(r.Context(), `INSERT INTO inventory_batches(id,organization_id,branch_id,item_id,lot_number,expires_on,quantity,purchase_price_uzs) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(organization_id,branch_id,item_id,lot_number) DO UPDATE SET quantity=inventory_batches.quantity+EXCLUDED.quantity,updated_at=now() RETURNING id`, newBatch, p.OrganizationID, in.ToBranchID, in.ItemID, lot, expires, in.Quantity, price).Scan(&newBatch)
			if err == nil {
				_, err = tx.Exec(r.Context(), `INSERT INTO inventory_movements(id,organization_id,branch_id,item_id,batch_id,movement_type,quantity,reference_id,reason,performed_by) VALUES($1,$2,$3,$4,$5,'TRANSFER_OUT',$6,$7,NULLIF($8,''),$9),($10,$2,$11,$4,$12,'TRANSFER_IN',$6,$7,NULLIF($8,''),$9)`, uuid.NewString(), p.OrganizationID, in.BranchID, in.ItemID, in.BatchID, in.Quantity, reference, reason, p.UserID, uuid.NewString(), in.ToBranchID, newBatch)
			}
		} else if err == nil {
			kind := in.Type
			_, err = tx.Exec(r.Context(), `INSERT INTO inventory_movements(id,organization_id,branch_id,item_id,batch_id,movement_type,quantity,reference_id,reason,performed_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),$10)`, uuid.NewString(), p.OrganizationID, in.BranchID, in.ItemID, in.BatchID, kind, in.Quantity, reference, reason, p.UserID)
		}
	}
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save movement", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'INVENTORY_MOVEMENT_CREATED','inventory_movement',$4,'success',jsonb_build_object('type',$5::text,'quantity',$6::numeric,'itemId',$7::text))`, p.OrganizationID, in.BranchID, p.UserID, reference, in.Type, in.Quantity, in.ItemID)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit movement", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]string{"id": reference})
}
func (a *App) listInventoryMovements(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT m.id,m.movement_type,m.quantity,m.reason,m.created_at,i.name,i.unit,b.name,bt.lot_number,concat_ws(' ',e.last_name,e.first_name) FROM inventory_movements m JOIN inventory_items i ON i.id=m.item_id JOIN branches b ON b.id=m.branch_id LEFT JOIN inventory_batches bt ON bt.id=m.batch_id LEFT JOIN employees e ON e.user_id=m.performed_by WHERE m.organization_id=$1 ORDER BY m.created_at DESC LIMIT 300`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load movements", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, kind, item, unit, branch string
		var qty float64
		var reason, lot, employee *string
		var created time.Time
		if rows.Scan(&id, &kind, &qty, &reason, &created, &item, &unit, &branch, &lot, &employee) == nil {
			items = append(items, map[string]any{"id": id, "type": kind, "quantity": qty, "reason": reason, "createdAt": created, "item": item, "unit": unit, "branch": branch, "lotNumber": lot, "employee": employee})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
