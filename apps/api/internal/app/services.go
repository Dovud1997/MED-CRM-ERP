package app

import (
	"net/http"
	"strings"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type serviceRequest struct {
	Name            string `json:"name"`
	Category        string `json:"category"`
	Code            string `json:"code"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"durationMinutes"`
	IsActive        *bool  `json:"isActive"`
}

type servicePriceRequest struct {
	BranchID   string  `json:"branchId"`
	EmployeeID *string `json:"employeeId"`
	AmountUZS  int64   `json:"amountUzs"`
}

type serviceProviderInput struct {
	EmployeeID string `json:"employeeId"`
	AmountUZS  *int64 `json:"amountUzs"`
}

type saveServiceProvidersRequest struct {
	BranchID      string                 `json:"branchId"`
	BaseAmountUZS int64                  `json:"baseAmountUzs"`
	Specialties   []string               `json:"specialties"`
	Providers     []serviceProviderInput `json:"providers"`
}

func (a *App) listServices(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT s.id,s.code,s.name,s.description,s.duration_minutes,s.is_active,COALESCE(c.name,'') FROM services s LEFT JOIN service_categories c ON c.id=s.category_id WHERE s.organization_id=$1 ORDER BY c.name,s.name`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load services", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	byID := map[string]map[string]any{}
	for rows.Next() {
		var id, code, name, description, category string
		var duration int
		var active bool
		if rows.Scan(&id, &code, &name, &description, &duration, &active, &category) == nil {
			item := map[string]any{"id": id, "code": code, "name": name, "description": description, "durationMinutes": duration, "isActive": active, "category": category, "prices": []map[string]any{}, "providers": []map[string]any{}, "specialties": []map[string]any{}}
			items = append(items, item)
			byID[id] = item
		}
	}
	priceRows, err := a.db.Query(r.Context(), `SELECT sp.service_id,sp.branch_id,b.name,sp.employee_id,COALESCE(e.last_name||' '||e.first_name,''),sp.amount_uzs FROM service_prices sp JOIN branches b ON b.id=sp.branch_id LEFT JOIN employees e ON e.id=sp.employee_id WHERE sp.organization_id=$1 ORDER BY b.name,e.last_name NULLS FIRST`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load prices", requestID(r))
		return
	}
	defer priceRows.Close()
	for priceRows.Next() {
		var serviceID, branchID, branchName, employeeName string
		var employeeID *string
		var amount int64
		if priceRows.Scan(&serviceID, &branchID, &branchName, &employeeID, &employeeName, &amount) == nil {
			if item := byID[serviceID]; item != nil {
				prices := item["prices"].([]map[string]any)
				item["prices"] = append(prices, map[string]any{"branchId": branchID, "branch": branchName, "employeeId": employeeID, "employee": employeeName, "amountUzs": amount})
			}
		}
	}
	providerRows, err := a.db.Query(r.Context(), `SELECT sp.service_id,sp.branch_id,sp.employee_id,e.last_name||' '||e.first_name FROM service_providers sp JOIN employees e ON e.id=sp.employee_id WHERE sp.organization_id=$1 ORDER BY e.last_name,e.first_name`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load service providers", requestID(r))
		return
	}
	defer providerRows.Close()
	for providerRows.Next() {
		var serviceID, branchID, employeeID, employeeName string
		if providerRows.Scan(&serviceID, &branchID, &employeeID, &employeeName) == nil {
			if item := byID[serviceID]; item != nil {
				providers := item["providers"].([]map[string]any)
				item["providers"] = append(providers, map[string]any{"branchId": branchID, "employeeId": employeeID, "employee": employeeName})
			}
		}
	}
	specialtyRows, err := a.db.Query(r.Context(), `SELECT service_id,branch_id,specialty FROM service_specialties WHERE organization_id=$1 ORDER BY specialty`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load service specialties", requestID(r))
		return
	}
	defer specialtyRows.Close()
	for specialtyRows.Next() {
		var serviceID, branchID, specialty string
		if specialtyRows.Scan(&serviceID, &branchID, &specialty) == nil {
			if item := byID[serviceID]; item != nil {
				values := item["specialties"].([]map[string]any)
				item["specialties"] = append(values, map[string]any{"branchId": branchID, "specialty": specialty})
			}
		}
	}
	writeJSON(w, 200, map[string]any{"items": items, "total": len(items)})
}

func normalizeService(in *serviceRequest) bool {
	in.Name = strings.TrimSpace(in.Name)
	in.Category = strings.TrimSpace(in.Category)
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	in.Description = strings.TrimSpace(in.Description)
	return len(in.Name) >= 2 && len(in.Name) <= 160 && len(in.Category) >= 2 && len(in.Category) <= 100 && in.DurationMinutes >= 5 && in.DurationMinutes <= 1440
}

func (a *App) createService(w http.ResponseWriter, r *http.Request) {
	var in serviceRequest
	if !decode(w, r, &in) {
		return
	}
	if !normalizeService(&in) {
		errorJSON(w, 422, "validation_error", "Check service fields", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create service", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var categoryID string
	if err = tx.QueryRow(r.Context(), `INSERT INTO service_categories(organization_id,name) VALUES($1,$2) ON CONFLICT(organization_id,name) DO UPDATE SET is_active=true RETURNING id`, p.OrganizationID, in.Category).Scan(&categoryID); err != nil {
		errorJSON(w, 500, "database_error", "Could not create category", requestID(r))
		return
	}
	id := uuid.NewString()
	if _, err = tx.Exec(r.Context(), `INSERT INTO services(id,organization_id,category_id,code,name,description,duration_minutes) VALUES($1,$2,$3,NULLIF($4,''),$5,$6,$7)`, id, p.OrganizationID, categoryID, in.Code, in.Name, in.Description, in.DurationMinutes); err != nil {
		errorJSON(w, 409, "service_exists", "Service name or code already exists", requestID(r))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'SERVICE_CREATED','service',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, in.Name)
	if err != nil || tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not save service", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

func (a *App) updateService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid service id", requestID(r))
		return
	}
	var in serviceRequest
	if !decode(w, r, &in) {
		return
	}
	if !normalizeService(&in) {
		errorJSON(w, 422, "validation_error", "Check service fields", requestID(r))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not update service", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var categoryID string
	if err = tx.QueryRow(r.Context(), `INSERT INTO service_categories(organization_id,name) VALUES($1,$2) ON CONFLICT(organization_id,name) DO UPDATE SET is_active=true RETURNING id`, p.OrganizationID, in.Category).Scan(&categoryID); err != nil {
		errorJSON(w, 500, "database_error", "Could not update category", requestID(r))
		return
	}
	tag, err := tx.Exec(r.Context(), `UPDATE services SET category_id=$1,code=NULLIF($2,''),name=$3,description=$4,duration_minutes=$5,is_active=$6,updated_at=now() WHERE id=$7 AND organization_id=$8`, categoryID, in.Code, in.Name, in.Description, in.DurationMinutes, active, id, p.OrganizationID)
	if err != nil {
		errorJSON(w, 409, "service_exists", "Service name or code already exists", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Service not found", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'SERVICE_UPDATED','service',$3,'success',jsonb_build_object('name',$4::text,'active',$5::boolean))`, p.OrganizationID, p.UserID, id, in.Name, active)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not update service", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}

func (a *App) deleteService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p := auth.PrincipalFrom(r.Context())
	tag, err := a.db.Exec(r.Context(), `UPDATE services SET is_active=false,updated_at=now() WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID)
	if err != nil || tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Service not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'SERVICE_DISABLED','service',$3,'success')`, p.OrganizationID, p.UserID, id)
	writeJSON(w, 200, map[string]bool{"success": true})
}

func (a *App) upsertServicePrice(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	var in servicePriceRequest
	if !decode(w, r, &in) {
		return
	}
	if _, err := uuid.Parse(serviceID); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid service id", requestID(r))
		return
	}
	if _, err := uuid.Parse(in.BranchID); err != nil || in.AmountUZS < 0 {
		errorJSON(w, 422, "validation_error", "Check price fields", requestID(r))
		return
	}
	var employee any = nil
	if in.EmployeeID != nil && strings.TrimSpace(*in.EmployeeID) != "" {
		if _, err := uuid.Parse(*in.EmployeeID); err != nil {
			errorJSON(w, 422, "validation_error", "Invalid employee", requestID(r))
			return
		}
		employee = *in.EmployeeID
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save price", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var exists bool
	if err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM services WHERE id=$1 AND organization_id=$2) AND EXISTS(SELECT 1 FROM branches WHERE id=$3 AND organization_id=$2)`, serviceID, p.OrganizationID, in.BranchID).Scan(&exists); err != nil || !exists {
		errorJSON(w, 404, "not_found", "Service or branch not found", requestID(r))
		return
	}
	if employee != nil {
		if err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id=$1 AND e.organization_id=$2 AND e.branch_id=$3 AND u.is_active=true AND u.deleted_at IS NULL)`, employee, p.OrganizationID, in.BranchID).Scan(&exists); err != nil {
			errorJSON(w, 500, "database_error", "Could not verify employee", requestID(r))
			return
		}
		if !exists {
			errorJSON(w, 422, "employee_branch_mismatch", "Doctor must belong to selected branch", requestID(r))
			return
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO service_prices(organization_id,service_id,branch_id,employee_id,amount_uzs) VALUES($1,$2,$3,$4,$5) ON CONFLICT(service_id,branch_id,employee_id) WHERE employee_id IS NOT NULL DO UPDATE SET amount_uzs=EXCLUDED.amount_uzs,updated_at=now()`, p.OrganizationID, serviceID, in.BranchID, employee, in.AmountUZS)
	} else {
		_, err = tx.Exec(r.Context(), `INSERT INTO service_prices(organization_id,service_id,branch_id,amount_uzs) VALUES($1,$2,$3,$4) ON CONFLICT(service_id,branch_id) WHERE employee_id IS NULL DO UPDATE SET amount_uzs=EXCLUDED.amount_uzs,updated_at=now()`, p.OrganizationID, serviceID, in.BranchID, in.AmountUZS)
	}
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save price", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'SERVICE_PRICE_UPDATED','service',$4,'success',jsonb_build_object('employeeId',$5::text,'amountUzs',$6::bigint))`, p.OrganizationID, in.BranchID, p.UserID, serviceID, employee, in.AmountUZS)
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not save price", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}

func (a *App) saveServiceProviders(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	var in saveServiceProvidersRequest
	if !decode(w, r, &in) {
		return
	}
	if _, err := uuid.Parse(serviceID); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid service id", requestID(r))
		return
	}
	if _, err := uuid.Parse(in.BranchID); err != nil || in.BaseAmountUZS < 0 {
		errorJSON(w, 422, "validation_error", "Check branch and price", requestID(r))
		return
	}
	seen := map[string]bool{}
	for _, provider := range in.Providers {
		if _, err := uuid.Parse(provider.EmployeeID); err != nil || seen[provider.EmployeeID] || provider.AmountUZS != nil && *provider.AmountUZS < 0 {
			errorJSON(w, 422, "validation_error", "Check selected doctors", requestID(r))
			return
		}
		seen[provider.EmployeeID] = true
	}
	specialtySeen := map[string]bool{}
	for index, value := range in.Specialties {
		value = strings.TrimSpace(value)
		if len(value) < 2 || len(value) > 120 || specialtySeen[value] {
			errorJSON(w, 422, "validation_error", "Check selected specialties", requestID(r))
			return
		}
		specialtySeen[value] = true
		in.Specialties[index] = value
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not save providers", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var validCount int
	if len(in.Providers) > 0 {
		ids := make([]string, 0, len(in.Providers))
		for _, provider := range in.Providers {
			ids = append(ids, provider.EmployeeID)
		}
		if err = tx.QueryRow(r.Context(), `SELECT count(*) FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id=ANY($1::uuid[]) AND e.organization_id=$2 AND e.branch_id=$3 AND u.is_active=true AND u.deleted_at IS NULL`, ids, p.OrganizationID, in.BranchID).Scan(&validCount); err != nil || validCount != len(ids) {
			errorJSON(w, 422, "employee_branch_mismatch", "All doctors must belong to selected branch", requestID(r))
			return
		}
	}
	var exists bool
	if err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM services WHERE id=$1 AND organization_id=$2) AND EXISTS(SELECT 1 FROM branches WHERE id=$3 AND organization_id=$2)`, serviceID, p.OrganizationID, in.BranchID).Scan(&exists); err != nil || !exists {
		errorJSON(w, 404, "not_found", "Service or branch not found", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `DELETE FROM service_providers WHERE service_id=$1 AND branch_id=$2`, serviceID, in.BranchID); err != nil {
		errorJSON(w, 500, "database_error", "Could not update providers", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `DELETE FROM service_specialties WHERE service_id=$1 AND branch_id=$2`, serviceID, in.BranchID); err != nil {
		errorJSON(w, 500, "database_error", "Could not update specialties", requestID(r))
		return
	}
	for _, specialty := range in.Specialties {
		var valid bool
		if err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM employees e JOIN users u ON u.id=e.user_id WHERE e.organization_id=$1 AND e.branch_id=$2 AND lower(e.specialty)=lower($3) AND u.is_active=true AND u.deleted_at IS NULL)`, p.OrganizationID, in.BranchID, specialty).Scan(&valid); err != nil || !valid {
			errorJSON(w, 422, "invalid_specialty", "Specialty is not available in selected branch", requestID(r))
			return
		}
		if _, err = tx.Exec(r.Context(), `INSERT INTO service_specialties(organization_id,service_id,branch_id,specialty) VALUES($1,$2,$3,$4)`, p.OrganizationID, serviceID, in.BranchID, specialty); err != nil {
			errorJSON(w, 500, "database_error", "Could not attach specialty", requestID(r))
			return
		}
	}
	if _, err = tx.Exec(r.Context(), `DELETE FROM service_prices WHERE service_id=$1 AND branch_id=$2 AND employee_id IS NOT NULL`, serviceID, in.BranchID); err != nil {
		errorJSON(w, 500, "database_error", "Could not update prices", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO service_prices(organization_id,service_id,branch_id,amount_uzs) VALUES($1,$2,$3,$4) ON CONFLICT(service_id,branch_id) WHERE employee_id IS NULL DO UPDATE SET amount_uzs=EXCLUDED.amount_uzs,updated_at=now()`, p.OrganizationID, serviceID, in.BranchID, in.BaseAmountUZS); err != nil {
		errorJSON(w, 500, "database_error", "Could not save base price", requestID(r))
		return
	}
	for _, provider := range in.Providers {
		if _, err = tx.Exec(r.Context(), `INSERT INTO service_providers(organization_id,service_id,branch_id,employee_id) VALUES($1,$2,$3,$4)`, p.OrganizationID, serviceID, in.BranchID, provider.EmployeeID); err != nil {
			errorJSON(w, 500, "database_error", "Could not attach doctor", requestID(r))
			return
		}
		if provider.AmountUZS != nil {
			if _, err = tx.Exec(r.Context(), `INSERT INTO service_prices(organization_id,service_id,branch_id,employee_id,amount_uzs) VALUES($1,$2,$3,$4,$5)`, p.OrganizationID, serviceID, in.BranchID, provider.EmployeeID, *provider.AmountUZS); err != nil {
				errorJSON(w, 500, "database_error", "Could not save doctor price", requestID(r))
				return
			}
		}
	}
	_, _ = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'SERVICE_PROVIDERS_UPDATED','service',$4,'success',jsonb_build_object('baseAmountUzs',$5::bigint,'providers',$6::int,'specialties',$7::int))`, p.OrganizationID, in.BranchID, p.UserID, serviceID, in.BaseAmountUZS, len(in.Providers), len(in.Specialties))
	if err = tx.Commit(r.Context()); err != nil {
		errorJSON(w, 500, "database_error", "Could not commit providers", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}
