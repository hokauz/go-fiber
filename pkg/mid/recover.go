package mid

import (
	"runtime/debug"

	"go-fiber/pkg/logx"
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
)

// Recoverer returns the recover middleware configured with default logger.
func Recoverer() fiber.Handler {
	return RecovererWith(logx.Default())
}

// RecovererWith returns recover middleware bound to a base logger.
func RecovererWith(base logx.Logger) fiber.Handler {
	if base == nil {
		base = logx.Default()
	}

	cfg := recoverer.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, recovered any) {
			l := logx.FromContext(c.Context())
			l.Error("panic recovered",
				logx.String("request_id", reqctx.RequestID(c)),
				logx.String("method", c.Method()),
				logx.String("path", c.Path()),
				logx.Any("panic", recovered),
				logx.String("stack", string(debug.Stack())),
			)
		},
	}

	return recoverer.New(cfg)
}
