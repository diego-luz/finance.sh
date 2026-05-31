package hash_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/finance-sh/finance-sh/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordAndCheck(t *testing.T) {
	const pw = "Sup3r-S3cr3t!"

	h, err := hash.Password(pw)
	require.NoError(t, err)
	require.NotEmpty(t, h)
	assert.NotEqual(t, pw, h, "hash must not equal the plaintext")

	assert.True(t, hash.Check(h, pw), "correct password must verify")
	assert.False(t, hash.Check(h, "wrong-password"), "wrong password must fail")
	assert.False(t, hash.Check("not-a-bcrypt-hash", pw), "invalid hash must fail")
}

func TestPasswordRandomSalt(t *testing.T) {
	const pw = "same-password"
	a, err := hash.Password(pw)
	require.NoError(t, err)
	b, err := hash.Password(pw)
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "bcrypt uses a random salt so hashes differ")
}

func TestRandomTokenLengthAndUniqueness(t *testing.T) {
	const n = 16
	tok, err := hash.RandomToken(n)
	require.NoError(t, err)
	// hex encoding => 2 chars per byte.
	assert.Len(t, tok, n*2)

	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		tk, err := hash.RandomToken(n)
		require.NoError(t, err)
		_, dup := seen[tk]
		assert.False(t, dup, "tokens must be unique")
		seen[tk] = struct{}{}
	}
}

func TestRandomTokenZeroBytes(t *testing.T) {
	tok, err := hash.RandomToken(0)
	require.NoError(t, err)
	assert.Equal(t, "", tok)
}

func TestSHA256Deterministic(t *testing.T) {
	a := hash.SHA256("refresh-token-value")
	b := hash.SHA256("refresh-token-value")
	assert.Equal(t, a, b, "SHA256 must be deterministic")
	assert.Len(t, a, 64, "hex sha256 is 64 chars")

	assert.NotEqual(t, hash.SHA256("a"), hash.SHA256("b"))
}

func TestSHA256KnownVector(t *testing.T) {
	// Known: sha256("abc") = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	assert.Equal(t, want, hash.SHA256("abc"))

	// Sanity-check the helper against the stdlib for a second value.
	sum := sha256.Sum256([]byte("finance-sh"))
	assert.Equal(t, hex.EncodeToString(sum[:]), hash.SHA256("finance-sh"))
}
