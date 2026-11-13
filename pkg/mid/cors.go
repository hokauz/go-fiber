package mid

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// Allowed headers for CORS requests (default is empty; needed for Authorization, X-CSRF-Token, etc.).
var corsAllowHeaders = []string{
	"Content-Type",
	"Content-Length",
	"Accept-Encoding",
	"X-CSRF-Token",
	"Authorization",
	"Accept",
	"Origin",
	"Cache-Control",
	"X-Requested-With",
}

// CORS returns the CORS middleware instance.
// allowedOrigins: nil/empty or ["*"] uses wildcard with AllowCredentials: false.
// Otherwise uses the list with AllowCredentials: true.
func CORS(allowedOrigins []string) fiber.Handler {
	cfg := cors.Config{
		AllowHeaders:     corsAllowHeaders,
		AllowCredentials: false,
	}

	if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
		cfg.AllowOrigins = []string{"*"}
	} else {
		cfg.AllowOrigins = allowedOrigins
		cfg.AllowCredentials = true
	}

	return cors.New(cfg)
}
