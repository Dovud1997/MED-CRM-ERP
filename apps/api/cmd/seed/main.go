package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

const (
	organizationID  = "10000000-0000-4000-8000-000000000001"
	centralBranchID = "20000000-0000-4000-8000-000000000001"
	secondBranchID  = "20000000-0000-4000-8000-000000000002"
	ownerUserID     = "30000000-0000-4000-8000-000000000001"
	ownerRoleID     = "40000000-0000-4000-8000-000000000001"
)

func main() {
	if os.Getenv("NODE_ENV") == "production" {
		log.Fatal("seed is disabled in production")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	loginName := strings.TrimSpace(os.Getenv("SEED_OWNER_LOGIN"))
	password := os.Getenv("SEED_OWNER_PASSWORD")
	if databaseURL == "" || loginName == "" || len(password) < 12 {
		log.Fatal("DATABASE_URL, SEED_OWNER_LOGIN and SEED_OWNER_PASSWORD (12+ chars) are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, databaseURL)
	must(err)
	defer db.Close()
	tx, err := db.Begin(ctx)
	must(err)
	defer tx.Rollback(ctx)
	hash, err := hashPassword(password)
	must(err)

	_, err = tx.Exec(ctx, `INSERT INTO organizations(id,name,currency,timezone) VALUES($1,'ONA VA BOLA KLINIKASI','UZS','Asia/Tashkent') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,updated_at=now()`, organizationID)
	must(err)
	_, err = tx.Exec(ctx, `INSERT INTO branches(id,organization_id,name,timezone,address) VALUES($1,$3,'Markaziy klinika','Asia/Tashkent','Toshkent'),($2,$3,'Ikkinchi filial','Asia/Tashkent','Toshkent') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,updated_at=now()`, centralBranchID, secondBranchID, organizationID)
	must(err)
	_, err = tx.Exec(ctx, `INSERT INTO roles(id,organization_id,code,name,is_system) VALUES($1,$2,'OWNER','Владелец',true) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name`, ownerRoleID, organizationID)
	must(err)
	_, err = tx.Exec(ctx, `INSERT INTO users(id,organization_id,login,password_hash,is_active,must_change_password) VALUES($1,$2,$3,$4,true,false) ON CONFLICT(id) DO UPDATE SET login=EXCLUDED.login,password_hash=EXCLUDED.password_hash,is_active=true,updated_at=now()`, ownerUserID, organizationID, strings.ToLower(loginName), hash)
	must(err)
	_, err = tx.Exec(ctx, `INSERT INTO employees(organization_id,user_id,branch_id,first_name,last_name,position) VALUES($1,$2,$3,'Шохрух','Хаджаев','Владелец клиники') ON CONFLICT(user_id) DO UPDATE SET branch_id=EXCLUDED.branch_id,updated_at=now()`, organizationID, ownerUserID, centralBranchID)
	must(err)
	_, err = tx.Exec(ctx, `INSERT INTO user_roles(user_id,role_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, ownerUserID, ownerRoleID)
	must(err)
	_, err = tx.Exec(ctx, `INSERT INTO role_permissions(role_id,permission_id) SELECT $1,id FROM permissions WHERE code='*' ON CONFLICT DO NOTHING`, ownerRoleID)
	must(err)
	standardRoles := []struct {
		code, name  string
		permissions []string
	}{
		{"DOCTOR_SURGEON", "Врач-хирург", []string{"branches:read", "employees:read", "patients:read", "patients:write", "clinical:read", "clinical:write", "appointments:read", "appointments:write"}},
		{"DOCTOR_DENTIST", "Врач-стоматолог", []string{"branches:read", "employees:read", "patients:read", "patients:write", "clinical:read", "clinical:write", "appointments:read", "appointments:write"}},
		{"DOCTOR_PEDIATRICIAN", "Врач-педиатр", []string{"branches:read", "employees:read", "patients:read", "patients:write", "clinical:read", "clinical:write", "appointments:read", "appointments:write"}},
		{"DOCTOR_ULTRASOUND", "Врач УЗИ", []string{"branches:read", "employees:read", "patients:read", "patients:write", "clinical:read", "clinical:write", "appointments:read", "appointments:write"}},
		{"DOCTOR_GYNECOLOGIST", "Врач-гинеколог", []string{"branches:read", "employees:read", "patients:read", "patients:write", "clinical:read", "clinical:write", "appointments:read", "appointments:write"}},
		{"SPEECH_THERAPIST", "Логопед", []string{"branches:read", "employees:read", "patients:read", "patients:write", "clinical:read", "clinical:write", "appointments:read", "appointments:write"}},
		{"ACCOUNTANT", "Бухгалтер", []string{"branches:read", "finance:read", "finance:write"}},
		{"MANAGER", "Менеджер", []string{"branches:read", "employees:read", "roles:read", "patients:read", "appointments:read"}},
		{"CASHIER", "Кассир", []string{"branches:read", "patients:read", "finance:read", "finance:write"}},
		{"RECEPTIONIST", "Рецепшн", []string{"branches:read", "patients:read", "patients:write", "appointments:read", "appointments:write"}},
	}
	for _, role := range standardRoles {
		var roleID string
		err = tx.QueryRow(ctx, `INSERT INTO roles(organization_id,code,name,is_system) VALUES($1,$2,$3,true) ON CONFLICT(organization_id,code) DO UPDATE SET name=EXCLUDED.name RETURNING id`, organizationID, role.code, role.name).Scan(&roleID)
		must(err)
		_, err = tx.Exec(ctx, `INSERT INTO role_permissions(role_id,permission_id) SELECT $1,id FROM permissions WHERE code=ANY($2::text[]) ON CONFLICT DO NOTHING`, roleID, role.permissions)
		must(err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1::uuid,$2::uuid,'LOCAL_SEED','organization',($1::uuid)::text,'success',jsonb_build_object('safe','no medical data'))`, organizationID, ownerUserID)
	must(err)
	must(tx.Commit(ctx))
	fmt.Printf("Local seed completed. Organization ID: %s, owner login: %s\n", organizationID, loginName)
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}
func must(err error) {
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
