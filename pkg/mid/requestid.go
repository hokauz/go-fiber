package mid

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// RequestID returns the request ID middleware instance.
func RequestID() fiber.Handler {
	return requestid.New()
}
