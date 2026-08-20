package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type loginRequest struct {
	OrganizationID string `json:"organizationId"`
	Login          string `json:"login"`
	Password       string `json:"password"`
}
type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if !decode(w, r, &in) {
		return
	}
	if in.OrganizationID == "" || in.Login == "" || len(in.Password) < 8 {
		errorJSON(w, 422, "validation_error", "Required fields are invalid", requestID(r))
		return
	}
	key := "login:" + in.OrganizationID + ":" + strings.ToLower(in.Login)
	attempts, _ := a.redis.Get(r.Context(), key).Int()
	if attempts >= 5 {
		errorJSON(w, 429, "account_locked", "Too many failed attempts", requestID(r))
		return
	}
	u, err := a.auth.FindUser(r.Context(), in.OrganizationID, in.Login)
	if err != nil || !auth.VerifyPassword(u.PasswordHash, in.Password) {
		pipe := a.redis.TxPipeline()
		pipe.Incr(r.Context(), key)
		pipe.Expire(r.Context(), key, 15*60*1e9)
		_, _ = pipe.Exec(r.Context())
		errorJSON(w, 401, "invalid_credentials", "Invalid credentials", requestID(r))
		return
	}
	a.redis.Del(r.Context(), key)
	access, refresh, err := a.auth.Issue(r.Context(), u, r.UserAgent(), remoteIP(r))
	if err != nil {
		errorJSON(w, 500, "internal_error", "Could not create session", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,ip_address,user_agent,result) VALUES($1::uuid,$2::uuid,'AUTH_LOGIN','user',($2::uuid)::text,$3,$4,'success')`, u.OrganizationID, u.ID, remoteIP(r), r.UserAgent())
	writeJSON(w, 200, map[string]any{"accessToken": access, "refreshToken": refresh, "mustChangePassword": u.MustChangePassword})
}
func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshRequest
	if !decode(w, r, &in) {
		return
	}
	access, refresh, err := a.auth.Rotate(r.Context(), in.RefreshToken, r.UserAgent(), remoteIP(r))
	if err != nil {
		errorJSON(w, 401, "invalid_refresh_token", "Refresh token is invalid", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]string{"accessToken": access, "refreshToken": refresh})
}
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	var in refreshRequest
	if !decode(w, r, &in) {
		return
	}
	_ = a.auth.Revoke(r.Context(), in.RefreshToken)
	writeJSON(w, 200, map[string]bool{"success": true})
}
func (a *App) me(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	writeJSON(w, 200, map[string]any{"id": p.UserID, "organizationId": p.OrganizationID, "permissions": p.Permissions})
}
func (a *App) listBranches(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT id,name,timezone,address,is_active,created_at,updated_at FROM branches WHERE organization_id=$1 AND deleted_at IS NULL ORDER BY name`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load branches", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name, tz string
		var address *string
		var active bool
		var created, updated any
		if rows.Scan(&id, &name, &tz, &address, &active, &created, &updated) == nil {
			items = append(items, map[string]any{"id": id, "name": name, "timezone": tz, "address": address, "isActive": active, "createdAt": created, "updatedAt": updated})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items, "total": len(items)})
}

type createBranchRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Timezone string `json:"timezone"`
}

func (a *App) createBranch(w http.ResponseWriter, r *http.Request) {
	var in createBranchRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Address = strings.TrimSpace(in.Address)
	in.Timezone = strings.TrimSpace(in.Timezone)
	if len(in.Name) < 2 || len(in.Name) > 120 || len(in.Address) > 500 {
		errorJSON(w, 422, "validation_error", "Branch fields are invalid", requestID(r))
		return
	}
	if in.Timezone == "" {
		in.Timezone = "Asia/Tashkent"
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create branch", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	id := uuid.NewString()
	var created any
	err = tx.QueryRow(r.Context(), `INSERT INTO branches(id,organization_id,name,timezone,address) VALUES($1,$2,$3,$4,NULLIF($5,'')) RETURNING created_at`, id, p.OrganizationID, in.Name, in.Timezone, in.Address).Scan(&created)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			errorJSON(w, 409, "branch_exists", "A branch with this name already exists", requestID(r))
			return
		}
		errorJSON(w, 500, "database_error", "Could not create branch", requestID(r))
		return
	}
	changes, _ := json.Marshal(map[string]string{"name": in.Name})
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,ip_address,user_agent,result,changes) VALUES($1,$2,'BRANCH_CREATED','branch',$3,$4,$5,'success',$6::jsonb)`, p.OrganizationID, p.UserID, id, remoteIP(r), r.UserAgent(), string(changes))
	if err != nil {
		errorJSON(w, 500, "audit_error", "Could not record change", requestID(r))
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		errorJSON(w, 500, "database_error", "Could not commit branch", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "name": in.Name, "address": in.Address, "timezone": in.Timezone, "isActive": true, "createdAt": created})
}

func (a *App) updateBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid branch id", requestID(r))
		return
	}
	var in createBranchRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name, in.Address, in.Timezone = strings.TrimSpace(in.Name), strings.TrimSpace(in.Address), strings.TrimSpace(in.Timezone)
	if len(in.Name) < 2 || len(in.Name) > 120 || len(in.Address) > 500 {
		errorJSON(w, 422, "validation_error", "Branch fields are invalid", requestID(r))
		return
	}
	if in.Timezone == "" {
		in.Timezone = "Asia/Tashkent"
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not update branch", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	tag, err := tx.Exec(r.Context(), `UPDATE branches SET name=$1,address=NULLIF($2,''),timezone=$3,updated_at=now() WHERE id=$4 AND organization_id=$5 AND deleted_at IS NULL`, in.Name, in.Address, in.Timezone, id, p.OrganizationID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			errorJSON(w, 409, "branch_exists", "A branch with this name already exists", requestID(r))
			return
		}
		errorJSON(w, 500, "database_error", "Could not update branch", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Branch not found", requestID(r))
		return
	}
	changes, _ := json.Marshal(map[string]string{"name": in.Name, "address": in.Address})
	if _, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'BRANCH_UPDATED','branch',$3,'success',$4::jsonb)`, p.OrganizationID, p.UserID, id, string(changes)); err != nil {
		errorJSON(w, 500, "audit_error", "Could not record branch update", requestID(r))
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		errorJSON(w, 500, "database_error", "Could not commit branch update", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "name": in.Name, "address": in.Address, "timezone": in.Timezone, "isActive": true})
}

func (a *App) deleteBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid branch id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not delete branch", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var employeeCount int
	if err = tx.QueryRow(r.Context(), `SELECT count(*) FROM employees e JOIN users u ON u.id=e.user_id WHERE e.branch_id=$1 AND u.deleted_at IS NULL`, id).Scan(&employeeCount); err != nil {
		errorJSON(w, 500, "database_error", "Could not check branch", requestID(r))
		return
	}
	if employeeCount > 0 {
		errorJSON(w, 409, "branch_has_employees", "Move employees to another branch before deletion", requestID(r))
		return
	}
	tag, err := tx.Exec(r.Context(), `UPDATE branches SET deleted_at=now(),is_active=false,updated_at=now() WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL`, id, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not delete branch", requestID(r))
		return
	}
	if tag.RowsAffected() == 0 {
		errorJSON(w, 404, "not_found", "Branch not found", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'BRANCH_DELETED','branch',$3,'success')`, p.OrganizationID, p.UserID, id); err != nil {
		errorJSON(w, 500, "audit_error", "Could not record branch deletion", requestID(r))
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		errorJSON(w, 500, "database_error", "Could not commit branch deletion", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}
func (a *App) listEmployees(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT e.id,u.id,u.login,e.first_name,e.last_name,e.middle_name,e.position,e.specialty,e.public_email,e.telegram_url,e.branch_id,b.name,u.is_active,COALESCE(pgp_sym_decrypt(e.passport_encrypted,$2),''),COALESCE(pgp_sym_decrypt(e.permanent_address_encrypted,$2),''),COALESCE(pgp_sym_decrypt(e.phone_encrypted,$2),''),COALESCE((SELECT jsonb_agg(ur.role_id) FROM user_roles ur WHERE ur.user_id=u.id),'[]'::jsonb),EXISTS(SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id=ur.role_id WHERE ur.user_id=u.id AND ro.code='OWNER'),EXISTS(SELECT 1 FROM profile_photos ph WHERE ph.organization_id=e.organization_id AND ph.entity_type='employee' AND ph.entity_id=e.id) FROM users u JOIN employees e ON e.user_id=u.id LEFT JOIN branches b ON b.id=e.branch_id WHERE u.organization_id=$1 AND u.deleted_at IS NULL ORDER BY e.last_name,e.first_name LIMIT 100`, p.OrganizationID, a.cfg.RefreshSecret)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load employees", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, userID, login, first, last, position string
		var middle, specialty, email, telegram, branchID, branch *string
		var active, owner, hasPhoto bool
		var passport, address, phone string
		var roleJSON []byte
		if rows.Scan(&id, &userID, &login, &first, &last, &middle, &position, &specialty, &email, &telegram, &branchID, &branch, &active, &passport, &address, &phone, &roleJSON, &owner, &hasPhoto) == nil {
			roleIDs := []string{}
			_ = json.Unmarshal(roleJSON, &roleIDs)
			items = append(items, map[string]any{"id": id, "userId": userID, "login": login, "firstName": first, "lastName": last, "middleName": middle, "passport": passport, "permanentAddress": address, "phoneLocal": strings.TrimPrefix(phone, "+998"), "position": position, "specialty": specialty, "publicEmail": email, "telegramUrl": telegram, "branchId": branchID, "branch": branch, "roleIds": roleIDs, "isActive": active, "isOwner": owner, "hasPhoto": hasPhoto})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items, "total": len(items)})
}
func (a *App) listAudit(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT id,actor_id,action,entity_type,entity_id,result,created_at FROM audit_logs WHERE organization_id=$1 ORDER BY created_at DESC LIMIT 100`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load audit", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var actor, entityID *string
		var action, entityType, result string
		var created any
		if rows.Scan(&id, &actor, &action, &entityType, &entityID, &result, &created) == nil {
			items = append(items, map[string]any{"id": id, "actorId": actor, "action": action, "entityType": entityType, "entityId": entityID, "result": result, "createdAt": created})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func remoteIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return strings.TrimSpace(strings.Split(v, ",")[0])
	}
	return r.RemoteAddr
}
