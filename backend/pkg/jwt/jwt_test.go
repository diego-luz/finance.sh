package jwt_test

import (
	"testing"
	"time"

	"github.com/finance-sh/finance-sh/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret = "super-secret-signing-key"
	testUserID = "11111111-2222-3333-4444-555555555555"
	testEmail  = "diego@finance.sh"
)

func TestGenerateThenParse(t *testing.T) {
	tok, err := jwt.Generate(testUserID, testEmail, testSecret, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	claims, err := jwt.Parse(tok, testSecret)
	require.NoError(t, err)
	assert.Equal(t, testUserID, claims.UserID)
	assert.Equal(t, testEmail, claims.Email)
	assert.Equal(t, testUserID, claims.Subject)
	assert.Empty(t, claims.Purpose, "default access token carries no purpose")
}

func TestParseExpiredToken(t *testing.T) {
	// Negative TTL => already expired.
	tok, err := jwt.Generate(testUserID, testEmail, testSecret, -time.Minute)
	require.NoError(t, err)

	_, err = jwt.Parse(tok, testSecret)
	assert.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestParseWrongSecret(t *testing.T) {
	tok, err := jwt.Generate(testUserID, testEmail, testSecret, time.Hour)
	require.NoError(t, err)

	_, err = jwt.Parse(tok, "a-different-secret")
	assert.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestParseTamperedToken(t *testing.T) {
	tok, err := jwt.Generate(testUserID, testEmail, testSecret, time.Hour)
	require.NoError(t, err)

	// Flip the last character of the signature segment.
	b := []byte(tok)
	last := b[len(b)-1]
	if last == 'A' {
		b[len(b)-1] = 'B'
	} else {
		b[len(b)-1] = 'A'
	}
	tampered := string(b)

	_, err = jwt.Parse(tampered, testSecret)
	assert.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestParseGarbage(t *testing.T) {
	_, err := jwt.Parse("not.a.jwt", testSecret)
	assert.ErrorIs(t, err, jwt.ErrInvalidToken)

	_, err = jwt.Parse("", testSecret)
	assert.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestGenerateWithPurpose(t *testing.T) {
	tok, err := jwt.GenerateWithPurpose(testUserID, testEmail, "mfa", testSecret, 5*time.Minute)
	require.NoError(t, err)

	claims, err := jwt.Parse(tok, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "mfa", claims.Purpose)
	assert.Equal(t, testUserID, claims.UserID)
	assert.Equal(t, testEmail, claims.Email)
}
