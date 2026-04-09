package jwt

import (
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
