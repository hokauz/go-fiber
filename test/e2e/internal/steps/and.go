//go:build e2e

package steps

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e/internal/helpers"
)

// AndJSONFieldEquals delegates to helpers.AndJSONFieldEquals.
func AndJSONFieldEquals(t *testing.T, resp *http.Response, field, want string) {
	t.Helper()
	helpers.AndJSONFieldEquals(t, resp, field, want)
}

// AndErrorCodeIs delegates to helpers.AndErrorCodeIs.
func AndErrorCodeIs(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	helpers.AndErrorCodeIs(t, resp, want)
}
