package hash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// Password hashes a plaintext password with bcrypt.
func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// Check compares a bcrypt hash against a plaintext password.
func Check(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// RandomToken returns a cryptographically random opaque token (hex).
func RandomToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SHA256 returns the hex sha256 of s. Used to store refresh tokens without
// keeping the raw value.
func SHA256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
