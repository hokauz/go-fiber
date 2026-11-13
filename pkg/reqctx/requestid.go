package reqctx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// RequestID returns the request ID from the context, or the X-Request-ID header if not set.
func RequestID(c fiber.Ctx) string {
	id := requestid.FromContext(c)
	if id == "" {
		return c.Get(fiber.HeaderXRequestID)
	}
	return id
}
