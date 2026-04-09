package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID assigns X-Request-ID if the client did not send one.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
			c.Set("X-Request-ID", rid)
		}
		return c.Next()
	}
}
