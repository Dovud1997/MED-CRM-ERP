package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/argon2"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: backupcheck <archive.ovbk>")
	}
	password := os.Getenv("BACKUP_PASSWORD")
	if len(password) < 12 {
		log.Fatal("BACKUP_PASSWORD must contain at least 12 characters")
	}
	payload, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("cannot read archive")
	}
	if len(payload) < 34 || string(payload[:5]) != "OVBK1" {
		log.Fatal("invalid backup format")
	}
	salt, nonce, encrypted := payload[5:21], payload[21:33], payload[33:]
	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("cannot initialize decryption")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal("cannot initialize archive authentication")
	}
	raw, err := gcm.Open(nil, nonce, encrypted, []byte("OVBK-1"))
	if err != nil || !json.Valid(raw) {
		log.Fatal("archive authentication failed: wrong password or damaged file")
	}
	fmt.Println("Backup is authentic, decryptable, and contains valid JSON.")
}
