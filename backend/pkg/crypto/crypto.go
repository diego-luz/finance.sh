// Package crypto provides field-level encryption at rest using AES-256-GCM.
//
// It exposes a process-wide cipher initialised once from the configured
// ENCRYPTION_KEY (base64-encoded 32 bytes) and an EncryptedString GORM custom
// type that transparently encrypts on write (driver.Valuer) and decrypts on
// read (sql.Scanner).
//
// Query-vs-encryption tradeoff: only fields that are NEVER used in WHERE/ORDER
// clauses or shown in lists/search are encrypted at the column level. Fields
// like email (login equality lookup), transaction.description, account.name,
// category.name and goal.name are searched, sorted or displayed and therefore
// stay in plaintext — encrypting them would break those queries. Confidentiality
// for those columns is provided by volume-level + backup encryption handled by
// the infrastructure layer; this package adds defence-in-depth for the most
// sensitive free-text/secret fields (transaction notes, 2FA secrets).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// DevDefaultKey is a FIXED, well-known 32-byte key (base64) used only as a
// development fallback. Using it in production is unsafe — Init logs a WARNING
// when it is detected. Generated 2026-05-28 via `openssl rand -base64 32` during
// the Cofre→finance.sh rebrand (previous ASCII-encoded dev key was rotated).
const DevDefaultKey = "a23pl9SKFOJF39DSpLJaSL3NYHpSIORb+6MMpw20GbI=" // 32 random bytes (AES-256 dev key)

var (
	gcm     cipher.AEAD
	gcmOnce sync.Once
	gcmErr  error

	// ErrNotInitialised is returned when encryption is used before Init succeeds.
	ErrNotInitialised = errors.New("crypto: cipher not initialised")
)

// Init configures the package-level AES-256-GCM cipher from a base64-encoded
// 32-byte key. It is safe to call once at startup. When the key equals the dev
// default a WARNING is logged. It returns an error if the key is malformed or
// not exactly 32 bytes.
func Init(base64Key string) error {
	gcmOnce.Do(func() {
		gcmErr = initCipher(base64Key)
	})
	return gcmErr
}

func initCipher(base64Key string) error {
	if base64Key == "" {
		base64Key = DevDefaultKey
	}
	if base64Key == DevDefaultKey {
		slog.Warn("crypto: using the built-in DEVELOPMENT encryption key — set ENCRYPTION_KEY to a unique base64-encoded 32-byte value in production")
	}

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return fmt.Errorf("crypto: ENCRYPTION_KEY is not valid base64: %w", err)
	}
	if len(key) != 32 {
		return fmt.Errorf("crypto: ENCRYPTION_KEY must decode to 32 bytes for AES-256, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("crypto: failed to create AES cipher: %w", err)
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("crypto: failed to create GCM: %w", err)
	}
	gcm = g
	return nil
}

// Encrypt seals plaintext with AES-256-GCM and returns base64(nonce||ciphertext).
// An empty input returns an empty string (so empty stays empty in the DB).
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if gcm == nil {
		return "", ErrNotInitialised
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// ErrDecrypt is returned when a value that is well-formed ciphertext (valid
// base64 of at least nonce+tag length) fails GCM authentication. This signals a
// real problem (wrong key, corruption, tampering) rather than a legacy plaintext
// row, so callers must surface it instead of returning corrupt data.
var ErrDecrypt = errors.New("crypto: decryption failed (authentication error)")

// Decrypt reverses Encrypt. An empty input returns an empty string.
//
// Legacy plaintext tolerance: a value that clearly is NOT ciphertext — not valid
// base64, or shorter than nonce+tag — is returned as-is (e.g. rows written before
// encryption was enabled). But when the input IS well-formed ciphertext yet GCM
// authentication fails, an error is returned rather than the raw input, so a key
// misconfiguration or corruption never silently masks/corrupts secrets.
func Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	if gcm == nil {
		return "", ErrNotInitialised
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Not base64: treat as legacy plaintext.
		return encoded, nil
	}
	// Minimum length for a real ciphertext is nonce + GCM tag (Overhead).
	minLen := gcm.NonceSize() + gcm.Overhead()
	if len(raw) < minLen {
		// Too short to be ciphertext we produced: treat as legacy plaintext.
		return encoded, nil
	}
	ns := gcm.NonceSize()
	nonce, ciphertext := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Well-formed ciphertext that fails authentication: wrong key or
		// corruption. Surface the error instead of returning corrupt data.
		return "", ErrDecrypt
	}
	return string(plain), nil
}

// EncryptedString is a GORM/SQL custom type that stores its value encrypted at
// rest. It behaves like a string in Go code; encryption happens at the DB
// boundary via the driver.Valuer and sql.Scanner interfaces. Empty stays empty.
type EncryptedString string

// Value implements driver.Valuer: encrypts before persisting.
func (e EncryptedString) Value() (driver.Value, error) {
	if e == "" {
		return "", nil
	}
	enc, err := Encrypt(string(e))
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Scan implements sql.Scanner: decrypts when loading from the DB.
func (e *EncryptedString) Scan(src interface{}) error {
	if src == nil {
		*e = ""
		return nil
	}
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("crypto: cannot scan %T into EncryptedString", src)
	}
	dec, err := Decrypt(raw)
	if err != nil {
		return err
	}
	*e = EncryptedString(dec)
	return nil
}

// GormDataType makes GORM map the column to text.
func (EncryptedString) GormDataType() string { return "text" }

// String returns the plaintext value.
func (e EncryptedString) String() string { return string(e) }
