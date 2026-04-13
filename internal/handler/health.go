// Package handler implements Helios Core HTTP routes.
package handler

import "github.com/gofiber/fiber/v2"

// Health responds OK for load balancers and orchestrators.
func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
