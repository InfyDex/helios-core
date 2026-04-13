// Package jwt signs and validates Helios access tokens (HS256).
package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims are embedded in Helios-issued access tokens (extensible for roles later).
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwtlib.RegisteredClaims
}

// Sign creates an HS256 JWT for the given user with configurable lifetime.
func Sign(userID, email, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(expiry)),
			Subject:   userID,
		},
	}
	t := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, &claims)
	return t.SignedString([]byte(secret))
}

// Validate parses and verifies a Helios HS256 access token issued by Sign.
func Validate(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwtlib.ParseWithClaims(tokenString, claims, func(t *jwtlib.Token) (any, error) {
		if t.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims.UserID == "" || claims.Email == "" {
		return nil, fmt.Errorf("jwt: missing user_id or email claim")
	}
	return claims, nil
}
