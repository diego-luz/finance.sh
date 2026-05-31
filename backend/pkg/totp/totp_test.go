package totp_test

import (
	"strings"
	"testing"
	"time"

	financetotp "github.com/finance-sh/finance-sh/pkg/totp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateProducesSecretAndURL(t *testing.T) {
	secret, url, err := financetotp.Generate("finance.sh", "diego@finance.sh")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	// otpauth URL must reference the issuer.
	assert.True(t, strings.HasPrefix(url, "otpauth://totp/"), "url=%s", url)
	assert.Contains(t, url, "issuer=finance.sh")
	assert.Contains(t, url, "finance.sh")
}

func TestValidateAcceptsCorrectCode(t *testing.T) {
	secret, _, err := financetotp.Generate("finance.sh", "diego@finance.sh")
	require.NoError(t, err)

	// Compute a valid code from the secret using the underlying lib.
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	assert.True(t, financetotp.Validate(code, secret), "freshly generated code must validate")
}

func TestValidateRejectsWrongCode(t *testing.T) {
	secret, _, err := financetotp.Generate("finance.sh", "diego@finance.sh")
	require.NoError(t, err)

	now := time.Now()
	valid, err := totp.GenerateCode(secret, now)
	require.NoError(t, err)

	// Pick a 6-digit code guaranteed to differ from the valid one.
	wrong := "000000"
	if valid == wrong {
		wrong = "111111"
	}
	assert.False(t, financetotp.Validate(wrong, secret), "a wrong code must fail")

	// A non-numeric / malformed code must fail.
	assert.False(t, financetotp.Validate("abcdef", secret))
}

func TestRecoveryCodes(t *testing.T) {
	codes, err := financetotp.RecoveryCodes(8)
	require.NoError(t, err)
	require.Len(t, codes, 8)

	seen := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		// Format "xxxx-xxxx-xxxx-xxxx": four 4-char hex groups (16 hex chars =
		// 64 bits of entropy). At least 10 readable chars overall.
		parts := strings.Split(c, "-")
		require.Len(t, parts, 4)
		for _, p := range parts {
			assert.Len(t, p, 4)
		}
		assert.GreaterOrEqual(t, len(strings.ReplaceAll(c, "-", "")), 10)

		_, dup := seen[c]
		assert.False(t, dup, "recovery codes must be unique")
		seen[c] = struct{}{}
	}
}
