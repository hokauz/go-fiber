//go:build e2e

package steps

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e/internal/helpers"
)

// ThenStatusIs delegates to helpers.ThenStatusIs.
func ThenStatusIs(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	helpers.ThenStatusIs(t, resp, want)
}
