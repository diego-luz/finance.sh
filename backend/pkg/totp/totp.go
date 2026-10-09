// Package totp wraps github.com/pquerna/otp for time-based one-time passwords
// (RFC 6238) and generation of single-use recovery codes.
package totp

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Generate creates a new TOTP secret for the given issuer/account. It returns
// the base32 secret and the otpauth:// URL used to render a QR code.
func Generate(issuer, account string) (secret, url string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// Validate reports whether code is a currently valid TOTP for secret. It uses
// the default skew so a code from the adjacent window is still accepted.
// ValidateStep is Validate that also says which 30-second step the code
// belongs to (current one, or the one before or after), so the caller can
// refuse a step it has already accepted: a code seen over someone's shoulder,
// or replayed from a log, stayed valid for about 90 seconds.
func ValidateStep(code, secret string, now time.Time) (int64, bool) {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return 0, false
	}
	atual := now.Unix() / 30
	for _, passo := range []int64{atual - 1, atual, atual + 1} {
		esperado, err := totp.GenerateCodeCustom(secret, time.Unix(passo*30, 0), totp.ValidateOpts{
			Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1,
		})
		if err == nil && subtle.ConstantTimeCompare([]byte(esperado), []byte(code)) == 1 {
			return passo, true
		}
	}
	return 0, false
}

func Validate(code, secret string) bool {
	return totp.Validate(code, secret)
}

// RecoveryCodes returns n cryptographically-random single-use codes. Each code
// carries 64 bits of entropy (8 random bytes -> 16 hex chars) and is formatted in
// four readable groups: "xxxx-xxxx-xxxx-xxxx". The grouping is cosmetic; the full
// string (hyphens included) is what gets hashed and stored.
func RecoveryCodes(n int) ([]string, error) {
	codes := make([]string, 0, n)
	for i := 0; i < n; i++ {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		h := hex.EncodeToString(b) // 16 hex chars = 64 bits of entropy
		codes = append(codes, fmt.Sprintf("%s-%s-%s-%s", h[0:4], h[4:8], h[8:12], h[12:16]))
	}
	return codes, nil
}
