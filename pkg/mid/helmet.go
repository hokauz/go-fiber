package mid

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

// Helmet returns the helmet middleware instance.
// Sets security-related HTTP headers.
// CrossOriginResourcePolicy is set to "cross-origin" for CORS compatibility.
func Helmet() fiber.Handler {
	return helmet.New(helmet.Config{
		CrossOriginResourcePolicy: "cross-origin",
	})
}
