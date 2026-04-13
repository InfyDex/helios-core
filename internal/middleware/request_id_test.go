package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestID_setsHeaderWhenMissing(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/", func(c *fiber.Ctx) error {
		if c.Get("X-Request-ID") == "" {
			return fiber.NewError(fiber.StatusInternalServerError, "missing")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if resp.Header.Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID on response")
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func TestRequestID_preservesClientValue(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(RequestID())
	const clientRID = "client-rid-123"
	app.Get("/", func(c *fiber.Ctx) error {
		if c.Get("X-Request-ID") != clientRID {
			return fiber.NewError(fiber.StatusInternalServerError, "wrong")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", clientRID)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
