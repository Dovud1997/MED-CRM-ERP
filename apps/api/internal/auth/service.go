package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/argon2"
)

type Principal struct {
	UserID, OrganizationID string
	Permissions            map[string]bool
	TokenVersion           int
}
type principalKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}
func PrincipalFrom(ctx context.Context) Principal {
	p, _ := ctx.Value(principalKey{}).(Principal)
	return p
}
func (p Principal) Has(v string) bool { return p.Permissions[v] || p.Permissions["*"] }

type Service struct {
	db                    *pgxpool.Pool
	redis                 *redis.Client
	access, refresh       []byte
	accessTTL, refreshTTL time.Duration
}

func New(db *pgxpool.Pool, r *redis.Client, a, b []byte, at, rt int64) *Service {
	return &Service{db: db, redis: r, access: a, refresh: b, accessTTL: time.Duration(at), refreshTTL: time.Duration(rt)}
}

type User struct {
	ID, OrganizationID, PasswordHash string
	TokenVersion                     int
	Permissions                      []string
	MustChangePassword               bool
}

func (s *Service) FindUser(ctx context.Context, org, login string) (User, error) {
	var u User
	rows, err := s.db.Query(ctx, `SELECT u.id,u.organization_id,u.password_hash,u.token_version,u.must_change_password,COALESCE(array_agg(DISTINCT p.code) FILTER (WHERE p.code IS NOT NULL),'{}') FROM users u LEFT JOIN user_roles ur ON ur.user_id=u.id LEFT JOIN role_permissions rp ON rp.role_id=ur.role_id LEFT JOIN permissions p ON p.id=rp.permission_id WHERE u.organization_id=$1 AND lower(u.login)=lower($2) AND u.is_active AND u.deleted_at IS NULL GROUP BY u.id`, org, login)
	if err != nil {
		return u, err
	}
	defer rows.Close()
	if !rows.Next() {
		return u, errors.New("not found")
	}
	err = rows.Scan(&u.ID, &u.OrganizationID, &u.PasswordHash, &u.TokenVersion, &u.MustChangePassword, &u.Permissions)
	return u, err
}
func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, _ = fmtSscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	salt, err := decodeBase64(parts[4])
	if err != nil {
		return false
	}
	expected, err := decodeBase64(parts[5])
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtleEqual(actual, expected)
}
func (s *Service) Issue(ctx context.Context, u User, userAgent, ip string) (string, string, error) {
	perms := map[string]bool{}
	for _, p := range u.Permissions {
		perms[p] = true
	}
	now := time.Now()
	claims := jwt.MapClaims{"sub": u.ID, "org": u.OrganizationID, "ver": u.TokenVersion, "permissions": u.Permissions, "iat": now.Unix(), "exp": now.Add(s.accessTTL).Unix(), "jti": uuid.NewString()}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.access)
	if err != nil {
		return "", "", err
	}
	sessionID := uuid.NewString()
	refreshClaims := jwt.MapClaims{"sub": u.ID, "sid": sessionID, "typ": "refresh", "iat": now.Unix(), "exp": now.Add(s.refreshTTL).Unix()}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.refresh)
	if err != nil {
		return "", "", err
	}
	_, err = s.db.Exec(ctx, `INSERT INTO sessions(id,user_id,token_hash,user_agent,ip_address,expires_at) VALUES($1,$2,$3,$4,$5,$6)`, sessionID, u.ID, hash(refresh), userAgent, ip, now.Add(s.refreshTTL))
	return access, refresh, err
}
func (s *Service) VerifyAccess(ctx context.Context, raw string) (Principal, error) {
	var p Principal
	if raw == "" {
		return p, errors.New("missing")
	}
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("algorithm")
		}
		return s.access, nil
	})
	if err != nil || !token.Valid {
		return p, errors.New("invalid")
	}
	c := token.Claims.(jwt.MapClaims)
	p.UserID, _ = c["sub"].(string)
	p.OrganizationID, _ = c["org"].(string)
	p.TokenVersion = int(c["ver"].(float64))
	p.Permissions = map[string]bool{}
	if values, ok := c["permissions"].([]any); ok {
		for _, v := range values {
			if x, ok := v.(string); ok {
				p.Permissions[x] = true
			}
		}
	}
	var active bool
	err = s.db.QueryRow(ctx, `SELECT is_active AND token_version=$2 FROM users WHERE id=$1 AND deleted_at IS NULL`, p.UserID, p.TokenVersion).Scan(&active)
	if err != nil || !active {
		return Principal{}, errors.New("revoked")
	}
	return p, nil
}
func (s *Service) Rotate(ctx context.Context, raw, userAgent, ip string) (string, string, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) { return s.refresh, nil })
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid")
	}
	c := token.Claims.(jwt.MapClaims)
	sid, _ := c["sid"].(string)
	uid, _ := c["sub"].(string)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback(ctx)
	var u User
	err = tx.QueryRow(ctx, `UPDATE sessions SET revoked_at=now() WHERE id=$1 AND user_id=$2 AND token_hash=$3 AND revoked_at IS NULL AND expires_at>now() RETURNING user_id`, sid, uid, hash(raw)).Scan(&u.ID)
	if err != nil {
		return "", "", errors.New("invalid")
	}
	err = tx.QueryRow(ctx, `SELECT organization_id,password_hash,token_version,must_change_password FROM users WHERE id=$1 AND is_active`, uid).Scan(&u.OrganizationID, &u.PasswordHash, &u.TokenVersion, &u.MustChangePassword)
	if err != nil {
		return "", "", err
	}
	rows, err := tx.Query(ctx, `SELECT DISTINCT p.code FROM user_roles ur JOIN role_permissions rp ON rp.role_id=ur.role_id JOIN permissions p ON p.id=rp.permission_id WHERE ur.user_id=$1`, uid)
	if err != nil {
		return "", "", err
	}
	for rows.Next() {
		var p string
		_ = rows.Scan(&p)
		u.Permissions = append(u.Permissions, p)
	}
	rows.Close()
	if err = tx.Commit(ctx); err != nil {
		return "", "", err
	}
	return s.Issue(ctx, u, userAgent, ip)
}
func (s *Service) Revoke(ctx context.Context, raw string) error {
	_, err := s.db.Exec(ctx, `UPDATE sessions SET revoked_at=now() WHERE token_hash=$1 AND revoked_at IS NULL`, hash(raw))
	return err
}
func hash(v string) string { x := sha256.Sum256([]byte(v)); return hex.EncodeToString(x[:]) }
