//go:build e2e

package hello

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
)

func TestHealth_ReadinessAndLiveness(t *testing.T) {
	ctx := e2e.NewTestContext(t)

	for _, path := range []string{"/api/livez", "/api/startupz"} {
		resp := ctx.JSONRequest(http.MethodGet, path, nil, nil)
		helpers.ThenStatusIs(t, resp, http.StatusOK)
	}

	expectReady := expectedDBReady()
	readyResp := ctx.JSONRequest(http.MethodGet, "/api/readyz", nil, nil)
	if expectReady {
		helpers.ThenStatusIs(t, readyResp, http.StatusOK)
		return
	}
	helpers.ThenStatusIs(t, readyResp, http.StatusServiceUnavailable)
}

func TestHelloDBPing_Status(t *testing.T) {
	ctx := e2e.NewTestContext(t)
	resp := ctx.JSONRequest(http.MethodGet, "/api/v1/hello/db-ping", nil, nil)

	expectReady := expectedDBReady()
	if expectReady {
		helpers.ThenStatusIs(t, resp, http.StatusOK)

		var body map[string]string
		helpers.MustDecodeJSON(t, resp, &body)
		if body["status"] != "ok" {
			t.Fatalf("status = %q, want %q", body["status"], "ok")
		}
		return
	}

	helpers.ThenStatusIs(t, resp, http.StatusServiceUnavailable)
	var body map[string]string
	helpers.MustDecodeJSON(t, resp, &body)

	expectedStatus := os.Getenv("E2E_EXPECT_DB_PING_STATUS")
	if strings.TrimSpace(expectedStatus) == "" {
		expectedStatus = "db_disabled"
	}
	if body["status"] != expectedStatus {
		t.Fatalf("status = %q, want %q", body["status"], expectedStatus)
	}
}

func expectedDBReady() bool {
	return strings.EqualFold(os.Getenv("E2E_EXPECT_DB_READY"), "1") ||
		strings.EqualFold(os.Getenv("E2E_EXPECT_DB_READY"), "true")
}
