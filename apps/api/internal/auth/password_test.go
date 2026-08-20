package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"testing"
)

func encoded(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}
func TestVerifyPassword(t *testing.T) {
	hash := encoded("correct-horse")
	if !VerifyPassword(hash, "correct-horse") {
		t.Fatal("valid password rejected")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("invalid password accepted")
	}
}
