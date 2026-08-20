package app

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *App) listPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), `SELECT id,code,description FROM permissions ORDER BY code`)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load permissions", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, code, description string
		if rows.Scan(&id, &code, &description) == nil {
			items = append(items, map[string]any{"id": id, "code": code, "description": description})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (a *App) listRoles(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT r.id,r.code,r.name,r.is_system,count(DISTINCT ur.user_id),COALESCE(array_agg(DISTINCT p.code) FILTER(WHERE p.code IS NOT NULL),'{}') FROM roles r LEFT JOIN user_roles ur ON ur.role_id=r.id LEFT JOIN role_permissions rp ON rp.role_id=r.id LEFT JOIN permissions p ON p.id=rp.permission_id WHERE r.organization_id=$1 GROUP BY r.id ORDER BY r.is_system DESC,r.name`, p.OrganizationID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load roles", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, code, name string
		var system bool
		var count int
		var permissions []string
		if rows.Scan(&id, &code, &name, &system, &count, &permissions) == nil {
			items = append(items, map[string]any{"id": id, "code": code, "name": name, "isSystem": system, "employeeCount": count, "permissions": permissions})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items, "total": len(items)})
}

type createRoleRequest struct {
	Name            string   `json:"name"`
	PermissionCodes []string `json:"permissionCodes"`
}

func (a *App) createRole(w http.ResponseWriter, r *http.Request) {
	var in createRoleRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) < 2 || len(in.Name) > 100 {
		errorJSON(w, 422, "validation_error", "Role name is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create role", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	id := uuid.NewString()
	code := "CUSTOM_" + strings.ToUpper(strings.ReplaceAll(id[:8], "-", ""))
	_, err = tx.Exec(r.Context(), `INSERT INTO roles(id,organization_id,code,name,is_system) VALUES($1,$2,$3,$4,false)`, id, p.OrganizationID, code, in.Name)
	if err != nil {
		errorJSON(w, 409, "role_exists", "Role already exists", requestID(r))
		return
	}
	if len(in.PermissionCodes) > 0 {
		tag, err := tx.Exec(r.Context(), `INSERT INTO role_permissions(role_id,permission_id) SELECT $1,id FROM permissions WHERE code=ANY($2::text[]) ON CONFLICT DO NOTHING`, id, in.PermissionCodes)
		if err != nil || tag.RowsAffected() != int64(len(in.PermissionCodes)) {
			errorJSON(w, 422, "invalid_permissions", "One or more permissions are invalid", requestID(r))
			return
		}
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,ip_address,user_agent,result,changes) VALUES($1,$2,'ROLE_CREATED','role',$3,$4,$5,'success',jsonb_build_object('name',$6::text))`, p.OrganizationID, p.UserID, id, remoteIP(r), r.UserAgent(), in.Name)
	if err != nil {
		errorJSON(w, 500, "audit_error", "Could not record role", requestID(r))
		return
	}
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit role", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "code": code, "name": in.Name, "isSystem": false, "permissions": in.PermissionCodes, "employeeCount": 0})
}

func (a *App) updateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid role id", requestID(r))
		return
	}
	var in createRoleRequest
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) < 2 || len(in.Name) > 100 {
		errorJSON(w, 422, "validation_error", "Role name is invalid", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not update role", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var code string
	if err = tx.QueryRow(r.Context(), `SELECT code FROM roles WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID).Scan(&code); err != nil {
		errorJSON(w, 404, "not_found", "Role not found", requestID(r))
		return
	}
	if code == "OWNER" {
		errorJSON(w, 409, "protected_role", "Owner role cannot be changed", requestID(r))
		return
	}
	if len(in.PermissionCodes) > 0 {
		var count int
		if err = tx.QueryRow(r.Context(), `SELECT count(*) FROM permissions WHERE code=ANY($1::text[])`, in.PermissionCodes).Scan(&count); err != nil || count != len(in.PermissionCodes) {
			errorJSON(w, 422, "invalid_permissions", "One or more permissions are invalid", requestID(r))
			return
		}
	}
	if _, err = tx.Exec(r.Context(), `UPDATE roles SET name=$1 WHERE id=$2`, in.Name, id); err != nil {
		errorJSON(w, 500, "database_error", "Could not update role", requestID(r))
		return
	}
	if _, err = tx.Exec(r.Context(), `DELETE FROM role_permissions WHERE role_id=$1`, id); err != nil {
		errorJSON(w, 500, "database_error", "Could not update permissions", requestID(r))
		return
	}
	if len(in.PermissionCodes) > 0 {
		if _, err = tx.Exec(r.Context(), `INSERT INTO role_permissions(role_id,permission_id) SELECT $1,id FROM permissions WHERE code=ANY($2::text[])`, id, in.PermissionCodes); err != nil {
			errorJSON(w, 500, "database_error", "Could not update permissions", requestID(r))
			return
		}
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'ROLE_UPDATED','role',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, in.Name); err != nil {
		errorJSON(w, 500, "audit_error", "Could not record role update", requestID(r))
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		errorJSON(w, 500, "database_error", "Could not commit role update", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "code": code, "name": in.Name, "permissions": in.PermissionCodes})
}

func (a *App) deleteRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid role id", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not delete role", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var code string
	var name string
	if err = tx.QueryRow(r.Context(), `SELECT code,name FROM roles WHERE id=$1 AND organization_id=$2`, id, p.OrganizationID).Scan(&code, &name); err != nil {
		errorJSON(w, 404, "not_found", "Role not found", requestID(r))
		return
	}
	if code == "OWNER" {
		errorJSON(w, 409, "protected_role", "Owner role cannot be deleted", requestID(r))
		return
	}
	_, _ = tx.Exec(r.Context(), `DELETE FROM user_roles WHERE role_id=$1`, id)
	_, _ = tx.Exec(r.Context(), `DELETE FROM role_permissions WHERE role_id=$1`, id)
	if _, err = tx.Exec(r.Context(), `DELETE FROM roles WHERE id=$1`, id); err != nil {
		errorJSON(w, 500, "database_error", "Could not delete role", requestID(r))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'ROLE_DELETED','role',$3,'success',jsonb_build_object('name',$4::text))`, p.OrganizationID, p.UserID, id, name)
	if err != nil {
		errorJSON(w, 500, "audit_error", "Could not record deletion", requestID(r))
		return
	}
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit deletion", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]bool{"success": true})
}

type createEmployeeRequest struct {
	Login            string   `json:"login"`
	Password         string   `json:"password"`
	FirstName        string   `json:"firstName"`
	LastName         string   `json:"lastName"`
	MiddleName       string   `json:"middleName"`
	Passport         string   `json:"passport"`
	PermanentAddress string   `json:"permanentAddress"`
	PhoneLocal       string   `json:"phoneLocal"`
	PublicEmail      string   `json:"publicEmail"`
	TelegramURL      string   `json:"telegramUrl"`
	Position         string   `json:"position"`
	Specialty        string   `json:"specialty"`
	BranchID         string   `json:"branchId"`
	RoleIDs          []string `json:"roleIds"`
}

var loginPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,64}$`)
var passportPattern = regexp.MustCompile(`^[A-Z]{2}[0-9]{7}$`)
var phonePattern = regexp.MustCompile(`^[0-9]{9}$`)
var telegramPattern = regexp.MustCompile(`^https://t\.me/[A-Za-z0-9_]{5,32}$`)

func (a *App) createEmployee(w http.ResponseWriter, r *http.Request) {
	var in createEmployeeRequest
	if !decode(w, r, &in) {
		return
	}
	in.Login = strings.ToLower(strings.TrimSpace(in.Login))
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.MiddleName = strings.TrimSpace(in.MiddleName)
	in.Passport = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(in.Passport), " ", ""))
	in.PermanentAddress = strings.TrimSpace(in.PermanentAddress)
	in.PhoneLocal = strings.ReplaceAll(strings.TrimSpace(in.PhoneLocal), " ", "")
	in.PublicEmail = strings.ToLower(strings.TrimSpace(in.PublicEmail))
	in.TelegramURL = strings.TrimSpace(in.TelegramURL)
	if strings.HasPrefix(in.TelegramURL, "@") {
		in.TelegramURL = "https://t.me/" + strings.TrimPrefix(in.TelegramURL, "@")
	}
	in.Position = strings.TrimSpace(in.Position)
	in.Specialty = strings.TrimSpace(in.Specialty)
	fieldErrors := map[string]string{}
	if !loginPattern.MatchString(in.Login) {
		fieldErrors["login"] = "invalid_login"
	}
	if len(in.Password) < 12 {
		fieldErrors["password"] = "too_short"
	}
	if len(in.FirstName) < 2 {
		fieldErrors["firstName"] = "too_short"
	}
	if len(in.LastName) < 2 {
		fieldErrors["lastName"] = "too_short"
	}
	if len(in.MiddleName) < 2 {
		fieldErrors["middleName"] = "too_short"
	}
	if !passportPattern.MatchString(in.Passport) {
		fieldErrors["passport"] = "invalid_passport"
	}
	if len(in.PermanentAddress) < 5 {
		fieldErrors["permanentAddress"] = "too_short"
	}
	if !phonePattern.MatchString(in.PhoneLocal) {
		fieldErrors["phoneLocal"] = "invalid_phone"
	}
	if in.PublicEmail != "" {
		parsedEmail, emailErr := mail.ParseAddress(in.PublicEmail)
		if emailErr != nil || parsedEmail.Address != in.PublicEmail {
			fieldErrors["publicEmail"] = "invalid_email"
		}
	}
	if in.TelegramURL != "" && !telegramPattern.MatchString(in.TelegramURL) {
		fieldErrors["telegramUrl"] = "invalid_telegram"
	}
	if len(in.Position) < 2 {
		fieldErrors["position"] = "too_short"
	}
	if len(in.RoleIDs) == 0 {
		fieldErrors["roleIds"] = "required"
	}
	if len(fieldErrors) > 0 {
		writeJSON(w, 422, map[string]any{"error": map[string]any{"code": "validation_error", "message": "Employee fields are invalid", "requestId": requestID(r), "fields": fieldErrors}})
		return
	}
	if _, err := uuid.Parse(in.BranchID); err != nil {
		writeJSON(w, 422, map[string]any{"error": map[string]any{"code": "validation_error", "message": "Branch is invalid", "requestId": requestID(r), "fields": map[string]string{"branchId": "required"}}})
		return
	}
	p := auth.PrincipalFrom(r.Context())
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		errorJSON(w, 500, "password_error", "Could not secure password", requestID(r))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create employee", requestID(r))
		return
	}
	defer tx.Rollback(r.Context())
	var branchOK bool
	_ = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, in.BranchID, p.OrganizationID).Scan(&branchOK)
	if !branchOK {
		errorJSON(w, 422, "invalid_branch", "Branch is invalid", requestID(r))
		return
	}
	var roleCount int
	_ = tx.QueryRow(r.Context(), `SELECT count(*) FROM roles WHERE id=ANY($1::uuid[]) AND organization_id=$2`, in.RoleIDs, p.OrganizationID).Scan(&roleCount)
	if roleCount != len(in.RoleIDs) {
		errorJSON(w, 422, "invalid_roles", "One or more roles are invalid", requestID(r))
		return
	}
	userID := uuid.NewString()
	employeeID := uuid.NewString()
	_, err = tx.Exec(r.Context(), `INSERT INTO users(id,organization_id,login,password_hash,is_active,must_change_password) VALUES($1,$2,$3,$4,true,true)`, userID, p.OrganizationID, in.Login, hash)
	if err != nil {
		errorJSON(w, 409, "login_exists", "Employee login already exists", requestID(r))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO employees(id,organization_id,user_id,branch_id,first_name,last_name,middle_name,passport_encrypted,permanent_address_encrypted,phone_encrypted,public_email,telegram_url,position,specialty) VALUES($1,$2,$3,$4,$5,$6,$7,pgp_sym_encrypt($8,$15),pgp_sym_encrypt($9,$15),pgp_sym_encrypt($10,$15),NULLIF($11,''),NULLIF($12,''),$13,NULLIF($14,''))`, employeeID, p.OrganizationID, userID, in.BranchID, in.FirstName, in.LastName, in.MiddleName, in.Passport, in.PermanentAddress, "+998"+in.PhoneLocal, in.PublicEmail, in.TelegramURL, in.Position, in.Specialty, a.cfg.RefreshSecret)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not create employee", requestID(r))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO user_roles(user_id,role_id) SELECT $1,unnest($2::uuid[])`, userID, in.RoleIDs)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not assign roles", requestID(r))
		return
	}
	changes, _ := json.Marshal(map[string]any{"roleCount": len(in.RoleIDs), "position": in.Position})
	_, err = tx.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'EMPLOYEE_CREATED','employee',$4,'success',$5::jsonb)`, p.OrganizationID, in.BranchID, p.UserID, employeeID, string(changes))
	if err != nil {
		errorJSON(w, 500, "audit_error", "Could not record employee", requestID(r))
		return
	}
	if tx.Commit(r.Context()) != nil {
		errorJSON(w, 500, "database_error", "Could not commit employee", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": employeeID, "userId": userID, "login": in.Login, "firstName": in.FirstName, "lastName": in.LastName, "middleName": in.MiddleName, "publicEmail": in.PublicEmail, "telegramUrl": in.TelegramURL, "position": in.Position, "specialty": in.Specialty, "branchId": in.BranchID, "roleIds": in.RoleIDs, "isActive": true, "mustChangePassword": true})
}

func normalizeEmployee(in *createEmployeeRequest) map[string]string {
	in.Login = strings.ToLower(strings.TrimSpace(in.Login)); in.FirstName = strings.TrimSpace(in.FirstName); in.LastName = strings.TrimSpace(in.LastName); in.MiddleName = strings.TrimSpace(in.MiddleName)
	in.Passport = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(in.Passport), " ", "")); in.PermanentAddress = strings.TrimSpace(in.PermanentAddress); in.PhoneLocal = strings.ReplaceAll(strings.TrimSpace(in.PhoneLocal), " ", "")
	in.PublicEmail = strings.ToLower(strings.TrimSpace(in.PublicEmail)); in.TelegramURL = strings.TrimSpace(in.TelegramURL); in.Position = strings.TrimSpace(in.Position); in.Specialty = strings.TrimSpace(in.Specialty)
	if strings.HasPrefix(in.TelegramURL, "@") { in.TelegramURL = "https://t.me/" + strings.TrimPrefix(in.TelegramURL, "@") }
	errs := map[string]string{}
	if !loginPattern.MatchString(in.Login) { errs["login"]="invalid_login" }; if len(in.FirstName)<2 { errs["firstName"]="too_short" }; if len(in.LastName)<2 { errs["lastName"]="too_short" }; if len(in.MiddleName)<2 { errs["middleName"]="too_short" }
	if !passportPattern.MatchString(in.Passport) { errs["passport"]="invalid_passport" }; if len(in.PermanentAddress)<5 { errs["permanentAddress"]="too_short" }; if !phonePattern.MatchString(in.PhoneLocal) { errs["phoneLocal"]="invalid_phone" }
	if in.PublicEmail!="" { parsed, err := mail.ParseAddress(in.PublicEmail); if err!=nil || parsed.Address!=in.PublicEmail { errs["publicEmail"]="invalid_email" } }; if in.TelegramURL!="" && !telegramPattern.MatchString(in.TelegramURL) { errs["telegramUrl"]="invalid_telegram" }
	if len(in.Position)<2 { errs["position"]="too_short" }; if len(in.RoleIDs)==0 { errs["roleIds"]="required" }; if _,err:=uuid.Parse(in.BranchID); err!=nil { errs["branchId"]="required" }
	return errs
}

func (a *App) updateEmployee(w http.ResponseWriter, r *http.Request) {
	var in createEmployeeRequest; if !decode(w,r,&in) { return }; errs:=normalizeEmployee(&in); if in.Password!="" && len(in.Password)<12 { errs["password"]="too_short" }
	if len(errs)>0 { writeJSON(w,422,map[string]any{"error":map[string]any{"code":"validation_error","message":"Employee fields are invalid","requestId":requestID(r),"fields":errs}}); return }
	p:=auth.PrincipalFrom(r.Context()); id:=chi.URLParam(r,"id"); tx,err:=a.db.Begin(r.Context()); if err!=nil { errorJSON(w,500,"database_error","Could not update employee",requestID(r)); return }; defer tx.Rollback(r.Context())
	var userID string; var owner bool
	err=tx.QueryRow(r.Context(),`SELECT e.user_id,EXISTS(SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id=ur.role_id WHERE ur.user_id=e.user_id AND ro.code='OWNER') FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id=$1 AND e.organization_id=$2 AND u.deleted_at IS NULL`,id,p.OrganizationID).Scan(&userID,&owner)
	if err!=nil { errorJSON(w,404,"employee_not_found","Employee not found",requestID(r)); return }
	var branchOK bool; _=tx.QueryRow(r.Context(),`SELECT EXISTS(SELECT 1 FROM branches WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`,in.BranchID,p.OrganizationID).Scan(&branchOK); if !branchOK { errorJSON(w,422,"invalid_branch","Branch is invalid",requestID(r)); return }
	var roleCount int; _=tx.QueryRow(r.Context(),`SELECT count(*) FROM roles WHERE id=ANY($1::uuid[]) AND organization_id=$2`,in.RoleIDs,p.OrganizationID).Scan(&roleCount); if roleCount!=len(in.RoleIDs) { errorJSON(w,422,"invalid_roles","Roles are invalid",requestID(r)); return }
	if owner { var keepsOwner bool; _=tx.QueryRow(r.Context(),`SELECT EXISTS(SELECT 1 FROM roles WHERE id=ANY($1::uuid[]) AND organization_id=$2 AND code='OWNER')`,in.RoleIDs,p.OrganizationID).Scan(&keepsOwner); if !keepsOwner { errorJSON(w,409,"protected_owner","Owner role cannot be removed",requestID(r)); return } }
	if in.Password!="" { hash,e:=auth.HashPassword(in.Password); if e!=nil { errorJSON(w,500,"password_error","Could not secure password",requestID(r)); return }; _,err=tx.Exec(r.Context(),`UPDATE users SET login=$1,password_hash=$2,must_change_password=true,token_version=token_version+1,updated_at=now() WHERE id=$3`,in.Login,hash,userID) } else { _,err=tx.Exec(r.Context(),`UPDATE users SET login=$1,token_version=token_version+1,updated_at=now() WHERE id=$2`,in.Login,userID) }
	if err!=nil { errorJSON(w,409,"login_exists","Employee login already exists",requestID(r)); return }
	_,err=tx.Exec(r.Context(),`UPDATE employees SET branch_id=$1,first_name=$2,last_name=$3,middle_name=$4,passport_encrypted=pgp_sym_encrypt($5,$14),permanent_address_encrypted=pgp_sym_encrypt($6,$14),phone_encrypted=pgp_sym_encrypt($7,$14),public_email=NULLIF($8,''),telegram_url=NULLIF($9,''),position=$10,specialty=NULLIF($11,''),updated_at=now() WHERE id=$12 AND organization_id=$13`,in.BranchID,in.FirstName,in.LastName,in.MiddleName,in.Passport,in.PermanentAddress,"+998"+in.PhoneLocal,in.PublicEmail,in.TelegramURL,in.Position,in.Specialty,id,p.OrganizationID,a.cfg.RefreshSecret); if err!=nil { errorJSON(w,500,"database_error","Could not update employee",requestID(r)); return }
	_,_=tx.Exec(r.Context(),`DELETE FROM user_roles WHERE user_id=$1`,userID); _,err=tx.Exec(r.Context(),`INSERT INTO user_roles(user_id,role_id) SELECT $1,unnest($2::uuid[])`,userID,in.RoleIDs); if err!=nil { errorJSON(w,500,"database_error","Could not update roles",requestID(r)); return }
	_,err=tx.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,branch_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'EMPLOYEE_UPDATED','employee',$4,'success',jsonb_build_object('roleCount',$5::int))`,p.OrganizationID,in.BranchID,p.UserID,id,len(in.RoleIDs)); if err!=nil { errorJSON(w,500,"audit_error","Could not record update",requestID(r)); return }
	if tx.Commit(r.Context())!=nil { errorJSON(w,500,"database_error","Could not commit update",requestID(r)); return }; writeJSON(w,200,map[string]bool{"success":true})
}

type employeeStatusRequest struct { IsActive bool `json:"isActive"` }
func (a *App) updateEmployeeStatus(w http.ResponseWriter,r *http.Request) { var in employeeStatusRequest; if !decode(w,r,&in){return}; a.changeEmployeeState(w,r,&in.IsActive,false) }
func (a *App) deleteEmployee(w http.ResponseWriter,r *http.Request) { a.changeEmployeeState(w,r,nil,true) }
func (a *App) changeEmployeeState(w http.ResponseWriter,r *http.Request,active *bool,deleted bool) {
	p:=auth.PrincipalFrom(r.Context()); id:=chi.URLParam(r,"id"); tx,err:=a.db.Begin(r.Context()); if err!=nil { errorJSON(w,500,"database_error","Could not change employee",requestID(r)); return }; defer tx.Rollback(r.Context())
	var userID string; var owner bool; err=tx.QueryRow(r.Context(),`SELECT e.user_id,EXISTS(SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id=ur.role_id WHERE ur.user_id=e.user_id AND ro.code='OWNER') FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id=$1 AND e.organization_id=$2 AND u.deleted_at IS NULL`,id,p.OrganizationID).Scan(&userID,&owner); if err!=nil { errorJSON(w,404,"employee_not_found","Employee not found",requestID(r)); return }
	if owner && (deleted || (active!=nil && !*active)) { errorJSON(w,409,"protected_owner","Owner cannot be disabled or deleted",requestID(r)); return }
	action:="EMPLOYEE_STATUS_CHANGED"; if deleted { _,err=tx.Exec(r.Context(),`UPDATE users SET is_active=false,deleted_at=now(),token_version=token_version+1,updated_at=now() WHERE id=$1`,userID); action="EMPLOYEE_ARCHIVED" } else { _,err=tx.Exec(r.Context(),`UPDATE users SET is_active=$1,token_version=token_version+1,updated_at=now() WHERE id=$2`,*active,userID) }; if err!=nil { errorJSON(w,500,"database_error","Could not change employee",requestID(r)); return }
	_,err=tx.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,$3,'employee',$4,'success',jsonb_build_object('active',$5::boolean,'archived',$6::boolean))`,p.OrganizationID,p.UserID,action,id,active!=nil && *active,deleted); if err!=nil { errorJSON(w,500,"audit_error","Could not record change",requestID(r)); return }
	if tx.Commit(r.Context())!=nil { errorJSON(w,500,"database_error","Could not commit change",requestID(r)); return }; writeJSON(w,200,map[string]bool{"success":true})
}
