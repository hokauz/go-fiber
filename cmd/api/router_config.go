package main

import (
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	apienv "go-fiber/cmd/api/env"
	"go-fiber/pkg/router"
)

const (
	defaultBasePath     = "/api"
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

func apiRouterConfig(server *apienv.ServerEnvs) fiber.Config {
	timeouts := router.HTTPTimeouts{
		Read:  defaultReadTimeout,
		Write: defaultWriteTimeout,
		Idle:  defaultIdleTimeout,
	}

	if server != nil {
		timeouts = router.HTTPTimeouts{
			Read:  timeoutFromSeconds(server.ReadTimeoutSeconds, defaultReadTimeout),
			Write: timeoutFromSeconds(server.WriteTimeoutSeconds, defaultWriteTimeout),
			Idle:  timeoutFromSeconds(server.IdleTimeoutSeconds, defaultIdleTimeout),
		}
	}

	return router.Config(timeouts)
}

func apiBasePath(server *apienv.ServerEnvs) string {
	if server == nil {
		return defaultBasePath
	}
	return normalizeBasePath(server.BasePath)
}

func normalizeBasePath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultBasePath
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}

	cleaned := path.Clean(raw)
	if cleaned == "." || cleaned == "" {
		return defaultBasePath
	}
	return cleaned
}

func timeoutFromSeconds(seconds int, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
