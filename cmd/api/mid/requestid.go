package mid

import (
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// RequestID returns the request ID middleware.
func RequestID() fiber.Handler {
	return mid.RequestID()
}
