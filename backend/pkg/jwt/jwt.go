package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

// Claims is the access-token payload. Purpose distinguishes token kinds: an
// empty value (or "access") is a normal access token; "mfa" is a short-lived
// token issued during a two-factor login challenge.
type Claims struct {
	UserID  string `json:"uid"`
	Email   string `json:"email"`
	Purpose string `json:"purpose,omitempty"`
	jwt.RegisteredClaims
}

// Generate signs a new access token for the given user.
func Generate(userID, email, secret string, ttl time.Duration) (string, error) {
	return GenerateWithPurpose(userID, email, "", secret, ttl)
}

// GenerateWithPurpose signs a token carrying a purpose claim (e.g. "mfa").
func GenerateWithPurpose(userID, email, purpose, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Email:   email,
		Purpose: purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// Parse validates a token and returns its claims.
func Parse(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
