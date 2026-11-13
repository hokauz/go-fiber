package mid

import "github.com/gofiber/fiber/v3"

func NotFound() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Matched() {
			return c.Next()
		}
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}
}
