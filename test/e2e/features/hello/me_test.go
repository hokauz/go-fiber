//go:build e2e

package hello

import (
	"net/http"
	"os"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
	"go-fiber/test/e2e/internal/steps"
)

func TestHelloMe_ReturnsUserForValidToken(t *testing.T) {
	token := os.Getenv("E2E_JWT_TOKEN")
	if token == "" {
		t.Fatal("E2E_JWT_TOKEN is required")
	}
	wantSub := os.Getenv("E2E_EXPECTED_SUB")
	if wantSub == "" {
		wantSub = "user-e2e-123"
	}

	t.Run("When_get_me_with_valid_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenGetHelloMe(t, ctx, token)
		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_response_has_request_id_and_user_id", func(t *testing.T) {
			parsed := helpers.MustParseJSON(t, resp)
			if rid, _ := parsed["request_id"].(string); rid == "" {
				t.Fatal("request_id should not be empty")
			}
			if uid, _ := parsed["user_id"].(string); uid != wantSub {
				t.Fatalf("user_id = %q, want %q", uid, wantSub)
			}
		})
	})
}
