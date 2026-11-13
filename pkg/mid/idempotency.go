package mid

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/idempotency"
)

// Idempotency returns the idempotency middleware instance.
// Caches responses keyed by X-Idempotency-Key header (UUID format).
func Idempotency() fiber.Handler {
	return idempotency.New()
}
