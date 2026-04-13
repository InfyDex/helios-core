package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/infydex/helios-core/internal/auth"
	"github.com/infydex/helios-core/internal/user"
)

type mockGoogleLogin struct {
	fn func(ctx context.Context, idToken string) (*auth.LoginResult, error)
}

func (m mockGoogleLogin) GoogleLogin(ctx context.Context, idToken string) (*auth.LoginResult, error) {
	return m.fn(ctx, idToken)
}

func TestGoogle_badJSON(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	NewAuth(app, mockGoogleLogin{fn: func(context.Context, string) (*auth.LoginResult, error) {
		t.Fatal("should not call service")
		return nil, nil
	}})

	req := httptest.NewRequest(fiber.MethodPost, "/auth/google", bytes.NewBufferString(`{`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestGoogle_missingIdToken(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	NewAuth(app, mockGoogleLogin{fn: func(context.Context, string) (*auth.LoginResult, error) {
		t.Fatal("should not call service")
		return nil, nil
	}})

	body, _ := json.Marshal(map[string]string{"idToken": "  "})
	req := httptest.NewRequest(fiber.MethodPost, "/auth/google", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestGoogle_invalidGoogleToken(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	NewAuth(app, mockGoogleLogin{fn: func(context.Context, string) (*auth.LoginResult, error) {
		return nil, auth.ErrInvalidGoogleToken
	}})

	body, _ := json.Marshal(map[string]string{"idToken": "fake"})
	req := httptest.NewRequest(fiber.MethodPost, "/auth/google", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestGoogle_duplicateUser(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	NewAuth(app, mockGoogleLogin{fn: func(context.Context, string) (*auth.LoginResult, error) {
		return nil, user.ErrDuplicateUser
	}})

	body, _ := json.Marshal(map[string]string{"idToken": "x"})
	req := httptest.NewRequest(fiber.MethodPost, "/auth/google", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestGoogle_success(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	NewAuth(app, mockGoogleLogin{fn: func(context.Context, string) (*auth.LoginResult, error) {
		return &auth.LoginResult{
			UserID: "u1", Email: "a@b.co", Name: "N", AvatarURL: "http://x", Phone: "+1",
			Token: "jwt-here",
		}, nil
	}})

	body, _ := json.Marshal(map[string]string{"idToken": "real-would-be-long"})
	req := httptest.NewRequest(fiber.MethodPost, "/auth/google", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	var out googleLoginResponse
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Token != "jwt-here" || out.User.Email != "a@b.co" || out.User.Phone != "+1" {
		t.Fatalf("body=%s", string(b))
	}
}
