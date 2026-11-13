package mid

import (
	"go-fiber/pkg/mid"

	"github.com/gofiber/fiber/v3"
)

// Helmet returns the helmet middleware.
func Helmet() fiber.Handler {
	return mid.Helmet()
}
