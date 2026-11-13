package metrics

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Middleware instruments HTTP requests with counter, histogram and in-flight gauge.
func Middleware(col *Collector) fiber.Handler {
	if col == nil {
		return func(c fiber.Ctx) error { return c.Next() }
	}

	return func(c fiber.Ctx) error {
		start := time.Now()
		col.IncInFlight()
		defer col.DecInFlight()

		err := c.Next()

		statusCode := c.Response().StatusCode()
		if err != nil {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				statusCode = fiberErr.Code
			} else if statusCode < fiber.StatusBadRequest {
				statusCode = fiber.StatusInternalServerError
			}
		}

		col.ObserveRequest(c.Method(), routePath(c), statusCode, time.Since(start))
		return err
	}
}

func routePath(c fiber.Ctx) string {
	if !c.Matched() {
		return unknownPath
	}
	route := c.Route()
	if route == nil {
		return unknownPath
	}
	return NormalizePath(route.Path)
}
