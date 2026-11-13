//go:build e2e

package acl

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
	"go-fiber/test/e2e/internal/steps"
)

// Scenario: checking an invitation with an unknown code returns valid:false.
// The check endpoint is public — no authentication is required.
func TestACLCheckInvitation_UnknownCode(t *testing.T) {
	t.Run("When_anyone_checks_an_unknown_invitation_code", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenCheckInvitation(t, ctx, "this-code-does-not-exist-xKm9Pq")

		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_invitation_is_invalid", func(t *testing.T) {
			steps.AndInvitationIsInvalid(t, resp)
		})
	})
}

// Scenario: checking without providing a code also returns valid:false.
func TestACLCheckInvitation_EmptyCode(t *testing.T) {
	t.Run("When_anyone_checks_without_a_code", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenCheckInvitationNoCode(t, ctx)

		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_invitation_is_invalid", func(t *testing.T) {
			steps.AndInvitationIsInvalid(t, resp)
		})
	})
}
