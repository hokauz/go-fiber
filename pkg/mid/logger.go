package mid

import (
	"math"
	"time"

	"go-fiber/pkg/logx"
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

const accessLogFormat = "${latency}"

// Logger returns the request logging middleware with default logger.
func Logger() fiber.Handler {
	return LoggerWith(logx.Default())
}

// LoggerWith returns the request logging middleware bound to a base logger.
func LoggerWith(base logx.Logger) fiber.Handler {
	if base == nil {
		base = logx.Default()
	}

	return logger.New(logger.Config{
		Format: accessLogFormat,
		LoggerFunc: func(c fiber.Ctx, data *logger.Data, _ *logger.Config) error {
			if skipHealthz(c) {
				return nil
			}

			durationMS := roundMillis(data.Stop.Sub(data.Start))

			base.Info("http.access",
				logx.String("request_id", reqctx.RequestID(c)),
				logx.String("method", c.Method()),
				logx.String("path", c.Path()),
				logx.Int("status", c.Response().StatusCode()),
				logx.Float64("duration_ms", durationMS),
				logx.String("ip", c.IP()),
				logx.String("user_agent", c.Get(fiber.HeaderUserAgent)),
				logx.Error(data.ChainErr),
			)
			return nil
		},
	})
}

func roundMillis(dur time.Duration) float64 {
	ms := float64(dur.Microseconds()) / 1000
	return math.Round(ms*1000) / 1000
}

func skipHealthz(c fiber.Ctx) bool {
	p := c.Path()
	return p == "/api/livez" || p == "/api/readyz" || p == "/api/startupz"
}

// LoggerWithConfig returns the logger middleware with custom config.
// Deprecated: prefer LoggerWith to keep canonical logx output.
func LoggerWithConfig(cfg logger.Config) fiber.Handler {
	return logger.New(cfg)
}
