package reqctx

import (
	"go-fiber/pkg/auth"

	"github.com/gofiber/fiber/v3"
)

type principalLocalKey struct{}

var principalKey principalLocalKey

// SetPrincipal stores the authenticated principal in request scope.
func SetPrincipal(c fiber.Ctx, p auth.Principal) {
	c.Locals(principalKey, p)
}

// PrincipalFrom returns the authenticated principal from request scope.
func PrincipalFrom(c fiber.Ctx) (auth.Principal, bool) {
	v := c.Locals(principalKey)
	if v == nil {
		return auth.Principal{}, false
	}

	p, ok := v.(auth.Principal)
	return p, ok
}

// PrincipalOrErr returns the principal from context or an error response suitable
// for returning from a Fiber handler. Use when the handler requires auth.
// Preserves PrincipalFrom for handlers that handle missing principal differently.
func PrincipalOrErr(c fiber.Ctx) (auth.Principal, error) {
	p, ok := PrincipalFrom(c)
	if !ok {
		return auth.Principal{}, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": auth.ErrAuthRequired.Code,
		})
	}
	return p, nil
}
