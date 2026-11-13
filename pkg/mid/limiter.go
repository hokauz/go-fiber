package mid

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// Limiter returns a rate limiter middleware.
// Default: 20 requests per 10 seconds per IP.
func Limiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Expiration: 10 * time.Second,
		Max:        20,
	})
}
