package handler

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/infydex/helios-core/internal/auth"
	"github.com/infydex/helios-core/internal/user"
)

const maxIDTokenRunes = 16384

// Auth exposes HTTP handlers for authentication.
type Auth struct {
	svc *auth.Service
}

// NewAuth registers routes on the given Fiber app.
func NewAuth(app fiber.Router, svc *auth.Service) *Auth {
	h := &Auth{svc: svc}
	app.Post("/auth/google", h.Google)
	return h
}

type googleLoginRequest struct {
	IDToken string `json:"idToken"`
}

type googleLoginResponse struct {
	User  userJSON `json:"user"`
	Token string   `json:"token"`
}

type userJSON struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// Google handles POST /core/v1/auth/google.
func (h *Auth) Google(c *fiber.Ctx) error {
	var req googleLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	token := strings.TrimSpace(req.IDToken)
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "idToken is required")
	}
	if utf8.RuneCountInString(token) > maxIDTokenRunes {
		return fiber.NewError(fiber.StatusBadRequest, "idToken too large")
	}

	res, err := h.svc.GoogleLogin(c.UserContext(), token)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidGoogleToken) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}
		if errors.Is(err, user.ErrDuplicateUser) {
			return fiber.NewError(fiber.StatusConflict, "user conflict")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
	}

	return c.JSON(googleLoginResponse{
		User: userJSON{
			ID:     res.UserID,
			Email:  res.Email,
			Name:   res.Name,
			Avatar: res.AvatarURL,
		},
		Token: res.Token,
	})
}
