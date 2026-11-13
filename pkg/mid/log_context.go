package mid

import (
	"go-fiber/pkg/logx"
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
)

// LogContext injects a request-scoped logger into c.Context().
func LogContext(base logx.Logger) fiber.Handler {
	if base == nil {
		base = logx.Default()
	}

	return func(c fiber.Ctx) error {
		reqLogger := base.With(
			logx.String("request_id", reqctx.RequestID(c)),
			logx.String("method", c.Method()),
			logx.String("path", c.Path()),
			logx.String("ip", c.IP()),
			logx.String("user_agent", c.Get(fiber.HeaderUserAgent)),
		)

		ctx := logx.WithLogger(c.Context(), reqLogger)
		c.SetContext(ctx)

		return c.Next()
	}
}
