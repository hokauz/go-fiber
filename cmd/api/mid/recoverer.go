package mid

import (
	"go-fiber/pkg/logx"
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// Recoverer returns the recover middleware.
func Recoverer() fiber.Handler {
	return mid.Recoverer()
}

// RecovererWith returns the recover middleware with explicit logger.
func RecovererWith(base logx.Logger) fiber.Handler {
	return mid.RecovererWith(base)
}
