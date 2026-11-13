//go:build e2e

package e2e

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// API_BASE_URL is required for e2e.
	if os.Getenv("API_BASE_URL") == "" {
		os.Setenv("API_BASE_URL", "http://127.0.0.1:8080")
	}
	os.Exit(m.Run())
}
