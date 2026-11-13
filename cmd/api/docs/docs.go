package docs

import (
	"fmt"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gofiber/fiber/v3"
)

var openAPISpecPath = "./docs/openapi.json"

// PublicRegister returns a handler that adds the Scalar documentation endpoint under /docs.
// It skips registration when isProduction is true.
func PublicRegister(isProduction bool) func(fiber.Router) {
	return func(r fiber.Router) {
		if isProduction {
			return
		}
		r.Get("/docs", scalarDocsPage)
	}
}

func scalarDocsPage(c fiber.Ctx) error {
	htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecURL: openAPISpecPath,
		CustomOptions: scalar.CustomOptions{
			PageTitle: "Go Fiber API Docs",
		},
		Layout:           scalar.LayoutModern,
		WithDefaultFonts: true,
		Theme:            "elysiajs",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to render api docs from %s: %v", openAPISpecPath, err),
		})
	}

	c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	return c.SendString(htmlContent)
}
