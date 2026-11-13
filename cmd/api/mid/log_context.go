package mid

import (
	"go-fiber/pkg/logx"
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// LogContext injects request-scoped logger into fiber context.
func LogContext(base logx.Logger) fiber.Handler {
	return mid.LogContext(base)
}
