package mid

import (
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// Idempotency returns the idempotency middleware.
func Idempotency() fiber.Handler {
	return mid.Idempotency()
}
