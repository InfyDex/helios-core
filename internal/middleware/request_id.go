// Package middleware provides HTTP middleware for Helios Core.
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
			// So downstream c.Get("X-Request-ID") sees the id; c.Set only affects the response.
			c.Request().Header.Set("X-Request-ID", rid)
		}
		c.Set("X-Request-ID", rid)
		return c.Next()
	}
}
