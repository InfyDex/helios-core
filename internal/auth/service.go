// Package auth orchestrates Google login and Helios JWT issuance.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/infydex/helios-core/internal/user"
	gverify "github.com/infydex/helios-core/pkg/google"
	"github.com/infydex/helios-core/pkg/jwt"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrInvalidGoogleToken means the Google ID token could not be verified.
var ErrInvalidGoogleToken = errors.New("invalid google token")

// Service orchestrates Google token verification, user persistence, and JWT issuance.
type Service struct {
	googleClientIDs []string
	jwtSecret       string
	jwtExpiry       time.Duration
	users           *user.Store
}

// NewService constructs an auth service. googleClientIDs are OAuth client IDs whose aud the ID token may use (web, Android, iOS).
func NewService(googleClientIDs []string, jwtSecret string, jwtExpiry time.Duration, users *user.Store) *Service {
	return &Service{
		googleClientIDs: googleClientIDs,
		jwtSecret:       jwtSecret,
		jwtExpiry:       jwtExpiry,
		users:           users,
	}
}

// LoginResult is returned after a successful Google login.
type LoginResult struct {
	UserID    string
	Email     string
	Name      string
	AvatarURL string
	Phone     string
	Token     string
}

// GoogleLogin verifies the Google ID token, upserts the user, and returns a Helios JWT.
func (s *Service) GoogleLogin(ctx context.Context, idToken string) (*LoginResult, error) {
	prof, err := gverify.VerifyIDToken(ctx, idToken, s.googleClientIDs)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGoogleToken, err)
	}
	u, err := s.users.GetOrCreateByGoogle(ctx, prof.Email, prof.Name, prof.Picture, prof.Phone, prof.Sub)
	if err != nil {
		return nil, err
	}
	idStr, err := uuidFromPgtype(u.ID)
	if err != nil {
		return nil, err
	}
	token, err := jwt.Sign(idStr, u.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	avatar := ""
	if u.AvatarUrl.Valid {
		avatar = u.AvatarUrl.String
	}
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	return &LoginResult{
		UserID:    idStr,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: avatar,
		Phone:     phone,
		Token:     token,
	}, nil
}

func uuidFromPgtype(u pgtype.UUID) (string, error) {
	if !u.Valid {
		return "", fmt.Errorf("user id not set")
	}
	id, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
