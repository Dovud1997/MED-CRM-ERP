package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"clinicos/api/internal/auth"
	"golang.org/x/crypto/argon2"
)

type backupRequest struct {
	Password string `json:"password"`
}

func (a *App) exportBackup(w http.ResponseWriter, r *http.Request) {
	var in backupRequest
	if !decode(w, r, &in) {
		return
	}
	if len(in.Password) < 12 {
		errorJSON(w, 422, "weak_backup_password", "Backup password must contain at least 12 characters", requestID(r))
		return
	}
	p := auth.PrincipalFrom(r.Context())
	var raw []byte
	err := a.db.QueryRow(r.Context(), `SELECT jsonb_build_object('format','OVBK-1','exportedAt',now(),'organization',(SELECT to_jsonb(o) FROM organizations o WHERE o.id=$1),'branches',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM branches x WHERE x.organization_id=$1),'[]'),'users',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM users x WHERE x.organization_id=$1),'[]'),'employees',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM employees x WHERE x.organization_id=$1),'[]'),'roles',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM roles x WHERE x.organization_id=$1),'[]'),'userRoles',COALESCE((SELECT jsonb_agg(to_jsonb(ur)) FROM user_roles ur JOIN users u ON u.id=ur.user_id WHERE u.organization_id=$1),'[]'),'rolePermissions',COALESCE((SELECT jsonb_agg(to_jsonb(rp)) FROM role_permissions rp JOIN roles ro ON ro.id=rp.role_id WHERE ro.organization_id=$1),'[]'),'patients',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM patients x WHERE x.organization_id=$1),'[]'),'appointments',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM appointments x WHERE x.organization_id=$1),'[]'),'serviceCategories',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM service_categories x WHERE x.organization_id=$1),'[]'),'services',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM services x WHERE x.organization_id=$1),'[]'),'servicePrices',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM service_prices x WHERE x.organization_id=$1),'[]'),'doctorSchedules',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM doctor_schedules x WHERE x.organization_id=$1),'[]'),'clinicalProfiles',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM patient_clinical_profiles x WHERE x.organization_id=$1),'[]'),'allergies',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM patient_allergies x WHERE x.organization_id=$1),'[]'),'vaccinations',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM patient_vaccinations x WHERE x.organization_id=$1),'[]'),'labResults',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM patient_lab_results x WHERE x.organization_id=$1),'[]'),'labAttachments',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM patient_lab_attachments x WHERE x.organization_id=$1),'[]'),'clinicalHistory',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM clinical_history_entries x WHERE x.organization_id=$1),'[]'),'retentionPolicy',(SELECT to_jsonb(x) FROM retention_policies x WHERE x.organization_id=$1),'audit',COALESCE((SELECT jsonb_agg(to_jsonb(x)) FROM audit_logs x WHERE x.organization_id=$1),'[]'))::text`, p.OrganizationID).Scan(&raw)
	if err != nil {
		errorJSON(w, 500, "backup_error", "Could not prepare backup", requestID(r))
		return
	}
	salt := make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		errorJSON(w, 500, "backup_error", "Could not secure backup", requestID(r))
		return
	}
	key := argon2.IDKey([]byte(in.Password), salt, 3, 64*1024, 2, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		errorJSON(w, 500, "backup_error", "Could not secure backup", requestID(r))
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		errorJSON(w, 500, "backup_error", "Could not secure backup", requestID(r))
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		errorJSON(w, 500, "backup_error", "Could not secure backup", requestID(r))
		return
	}
	encrypted := gcm.Seal(nil, nonce, raw, []byte("OVBK-1"))
	payload := append([]byte("OVBK1"), salt...)
	payload = append(payload, nonce...)
	payload = append(payload, encrypted...)
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,result,changes) VALUES($1,$2,'BACKUP_EXPORTED','organization','success',jsonb_build_object('bytes',$3::bigint))`, p.OrganizationID, p.UserID, len(payload))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="ona-va-bola-backup-%s.ovbk"`, time.Now().Format("2006-01-02-150405")))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(200)
	_, _ = w.Write(payload)
}
