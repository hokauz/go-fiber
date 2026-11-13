//go:build e2e

package steps

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e/internal/helpers"
)

// AndInvitationIsValid asserts the check response has valid:true.
func AndInvitationIsValid(t *testing.T, resp *http.Response) {
	t.Helper()
	m := helpers.MustParseJSON(t, resp)
	if valid, _ := m["valid"].(bool); !valid {
		t.Fatalf("valid: got false, want true; response: %v", m)
	}
}

// AndInvitationIsInvalid asserts the check response has valid:false.
func AndInvitationIsInvalid(t *testing.T, resp *http.Response) {
	t.Helper()
	m := helpers.MustParseJSON(t, resp)
	if valid, _ := m["valid"].(bool); valid {
		t.Fatalf("valid: got true, want false; response: %v", m)
	}
}

// AndInvitationStatusIs asserts the check response has the expected status string.
func AndInvitationStatusIs(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	helpers.AndJSONFieldEquals(t, resp, "status", want)
}

// AndAccessLevelIs asserts the effective-access-level response has the expected access_level.
func AndAccessLevelIs(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	helpers.AndJSONFieldEquals(t, resp, "access_level", want)
}

// AndRoleIs keeps backward-compat naming in e2e helpers.
func AndRoleIs(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	AndAccessLevelIs(t, resp, want)
}
