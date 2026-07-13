package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when a token is missing, malformed, expired, or
// signed with the wrong key.
var ErrInvalidToken = errors.New("auth: invalid token")

// TokenManager issues and verifies HS256 JWTs signed with a shared secret.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager builds a TokenManager. secret comes from JWT_SECRET; ttl is
// how long an issued token stays valid.
func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

// Issue returns a signed JWT whose subject is the user id and which expires
// after the manager's ttl.
func (tm *TokenManager) Issue(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tm.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}
	return signed, nil
}

// Parse verifies the token's signature and expiry and returns its user id.
// Any failure (bad signature, wrong method, expired, malformed) is reported as
// ErrInvalidToken so callers can respond with a single 401.
func (tm *TokenManager) Parse(tokenStr string) (uint, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return tm.secret, nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return uint(id), nil
}
