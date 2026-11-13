//go:build e2e

package hello

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
	"go-fiber/test/e2e/internal/steps"
)

func TestHelloPublic_Smoke(t *testing.T) {
	t.Run("When_get_public_hello", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenGetHelloPublic(t, ctx)
		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_request_id_present", func(t *testing.T) {
			parsed := helpers.MustParseJSON(t, resp)
			if rid, _ := parsed["request_id"].(string); rid == "" {
				t.Fatal("request_id should not be empty")
			}
		})
	})
}
