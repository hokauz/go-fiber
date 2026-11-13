//go:build e2e

package steps

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
)

// WhenGetHelloPublic performs GET /api/v1/hello (public, no auth).
func WhenGetHelloPublic(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.Get("/api/v1/hello")
}

// WhenGetHelloMe performs GET /api/v1/hello/me with bearer token.
func WhenGetHelloMe(t *testing.T, ctx *e2e.TestContext, token string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodGet, "/api/v1/hello/me", nil, helpers.AuthHeaders(token))
}

// WhenGetHelloMeNoAuth performs GET /api/v1/hello/me without auth headers.
func WhenGetHelloMeNoAuth(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodGet, "/api/v1/hello/me", nil, nil)
}
