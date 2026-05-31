package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testKey is a known, valid base64-encoded 32-byte AES-256 key.
const testKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=" // "0123456789abcdef0123456789abcdef"

// setupCipher forces the package-level GCM cipher to a deterministic test key.
// It bypasses the sync.Once guard so each test starts from a known state.
func setupCipher(t *testing.T) {
	t.Helper()
	key, err := base64.StdEncoding.DecodeString(testKey)
	require.NoError(t, err)
	require.Len(t, key, 32)
	block, err := aes.NewCipher(key)
	require.NoError(t, err)
	g, err := cipher.NewGCM(block)
	require.NoError(t, err)
	gcm = g
	gcmOnce = sync.Once{}
	gcmErr = nil
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	setupCipher(t)

	cases := []string{
		"hello world",
		"acentuação: ção, ã, é",
		"a",
		"a much longer secret note with spaces and 12345 numbers !@#$%",
	}
	for _, plain := range cases {
		ct, err := Encrypt(plain)
		require.NoError(t, err)
		require.NotEqual(t, plain, ct, "ciphertext must differ from plaintext")

		got, err := Decrypt(ct)
		require.NoError(t, err)
		assert.Equal(t, plain, got)
	}
}

func TestEncryptRandomNonce(t *testing.T) {
	setupCipher(t)

	a, err := Encrypt("same input")
	require.NoError(t, err)
	b, err := Encrypt("same input")
	require.NoError(t, err)

	assert.NotEqual(t, a, b, "each Encrypt call must produce a fresh nonce -> different ciphertext")

	// Both still decrypt back to the original.
	da, err := Decrypt(a)
	require.NoError(t, err)
	db, err := Decrypt(b)
	require.NoError(t, err)
	assert.Equal(t, "same input", da)
	assert.Equal(t, "same input", db)
}

func TestEncryptEmptyStaysEmpty(t *testing.T) {
	setupCipher(t)

	ct, err := Encrypt("")
	require.NoError(t, err)
	assert.Equal(t, "", ct)

	pt, err := Decrypt("")
	require.NoError(t, err)
	assert.Equal(t, "", pt)
}

func TestDecryptGarbageIsResilient(t *testing.T) {
	setupCipher(t)

	// Non-base64 input is treated as legacy plaintext and returned as-is.
	got, err := Decrypt("not base64 @@@")
	require.NoError(t, err)
	assert.Equal(t, "not base64 @@@", got)

	// Valid base64 but too short to hold a nonce: returned as-is.
	short := base64.StdEncoding.EncodeToString([]byte{0x01, 0x02})
	got, err = Decrypt(short)
	require.NoError(t, err)
	assert.Equal(t, short, got)

	// Valid base64, long enough to be well-formed ciphertext (>= nonce+tag), but
	// fails GCM authentication: this is NOT treated as legacy plaintext. It must
	// surface an error so a wrong key / corruption never silently returns garbage.
	junk := base64.StdEncoding.EncodeToString(make([]byte, 64))
	_, err = Decrypt(junk)
	assert.ErrorIs(t, err, ErrDecrypt)
}

func TestEncryptNotInitialised(t *testing.T) {
	// Force the uninitialised state.
	gcm = nil
	defer setupCipher(t) // restore so later tests aren't affected by ordering

	_, err := Encrypt("data")
	assert.ErrorIs(t, err, ErrNotInitialised)

	_, err = Decrypt("ZGF0YQ==")
	assert.ErrorIs(t, err, ErrNotInitialised)
}

func TestEncryptedStringValueScanRoundTrip(t *testing.T) {
	setupCipher(t)

	original := EncryptedString("2fa-secret-XYZ")

	// Value() encrypts.
	v, err := original.Value()
	require.NoError(t, err)
	stored, ok := v.(string)
	require.True(t, ok)
	require.NotEqual(t, original.String(), stored, "stored value must be encrypted")

	// Scan() from a string decrypts.
	var loaded EncryptedString
	require.NoError(t, loaded.Scan(stored))
	assert.Equal(t, original, loaded)

	// Scan() from []byte also works.
	var loadedBytes EncryptedString
	require.NoError(t, loadedBytes.Scan([]byte(stored)))
	assert.Equal(t, original, loadedBytes)
}

func TestEncryptedStringEmptyAndNil(t *testing.T) {
	setupCipher(t)

	empty := EncryptedString("")
	v, err := empty.Value()
	require.NoError(t, err)
	assert.Equal(t, "", v)

	var e EncryptedString
	require.NoError(t, e.Scan(nil))
	assert.Equal(t, EncryptedString(""), e)
}

func TestEncryptedStringScanBadType(t *testing.T) {
	setupCipher(t)

	var e EncryptedString
	err := e.Scan(12345)
	assert.Error(t, err)
}

func TestEncryptedStringGormDataType(t *testing.T) {
	assert.Equal(t, "text", EncryptedString("x").GormDataType())
}
