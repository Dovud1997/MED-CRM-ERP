package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func decodeBase64(v string) ([]byte, error)        { return base64.RawStdEncoding.DecodeString(v) }
func subtleEqual(a, b []byte) bool                 { return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1 }
func fmtSscanf(s, f string, a ...any) (int, error) { return fmt.Sscanf(s, f, a...) }
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}
