package mid

import (
	"go-fiber/pkg/logx"
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// Logger returns the request logging middleware.
func Logger() fiber.Handler {
	return mid.Logger()
}

// LoggerWith returns the request logging middleware with explicit logger.
func LoggerWith(base logx.Logger) fiber.Handler {
	return mid.LoggerWith(base)
}
