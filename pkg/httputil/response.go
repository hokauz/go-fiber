package httputil

import (
	"github.com/gofiber/fiber/v3"
)

// Unauthorized sends a 401 JSON response with the given error code.
func Unauthorized(c fiber.Ctx, code string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": code,
	})
}

// ServiceUnavailable sends a 503 JSON response with the given error code.
func ServiceUnavailable(c fiber.Ctx, code string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
		"error": code,
	})
}

// InternalError sends a 500 JSON response with the given error code.
func InternalError(c fiber.Ctx, code string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": code,
	})
}
