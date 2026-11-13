package acl

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	pkguuid "go-fiber/pkg/uuid"
)

func newUUID() (string, error) {
	return pkguuid.NewV7()
}

// generateInviteCode returns a random raw code and its SHA-256 hex hash.
func generateInviteCode() (rawCode, codeHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	rawCode = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(rawCode))
	codeHash = hex.EncodeToString(h[:])
	return
}

// hashCode returns the SHA-256 hex hash of the raw code.
func hashCode(rawCode string) string {
	h := sha256.Sum256([]byte(rawCode))
	return hex.EncodeToString(h[:])
}
