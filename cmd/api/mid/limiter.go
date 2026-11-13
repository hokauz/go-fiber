package mid

import (
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// Limiter returns the rate limiter middleware.
func Limiter() fiber.Handler {
	return mid.Limiter()
}
