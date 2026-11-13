package mid

import (
	"strings"

	apienv "go-fiber/cmd/api/env"
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

const wildcardOrigin = "*"

// CORS returns the CORS middleware.
func CORS(cfg *apienv.CORSEnvs) fiber.Handler {
	return mid.CORS(parseAllowedOrigins(cfg))
}

func parseAllowedOrigins(cfg *apienv.CORSEnvs) []string {
	if cfg == nil {
		return []string{wildcardOrigin}
	}

	parts := strings.Split(cfg.Origins, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		origin := strings.TrimSpace(p)
		if origin == "" {
			continue
		}
		if origin == wildcardOrigin {
			return []string{wildcardOrigin}
		}
		out = append(out, origin)
	}

	if len(out) == 0 {
		return []string{wildcardOrigin}
	}
	return out
}
