//go:build e2e

package acl

import (
	"net/http"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
	"go-fiber/test/e2e/internal/steps"
)

// Scenario: creating a group without a token is rejected.
func TestACLCreateGroup_RequiresAuth(t *testing.T) {
	t.Run("When_create_group_without_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenCreateGroupNoAuth(t, ctx)

		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_auth_required", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/missing-header")
		})
	})
}

// Scenario: listing groups without a token is rejected.
func TestACLListGroups_RequiresAuth(t *testing.T) {
	t.Run("When_list_groups_without_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenListGroupsNoAuth(t, ctx)

		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_auth_required", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/missing-header")
		})
	})
}

// Scenario: creating an invitation without a token is rejected.
func TestACLCreateInvitation_RequiresAuth(t *testing.T) {
	t.Run("When_create_invitation_without_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenCreateInvitationNoAuth(t, ctx)

		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_auth_required", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/missing-header")
		})
	})
}

// Scenario: accepting an invitation without a token is rejected.
func TestACLAcceptInvitation_RequiresAuth(t *testing.T) {
	t.Run("When_accept_invitation_without_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenAcceptInvitationNoAuth(t, ctx)

		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_auth_required", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/missing-header")
		})
	})
}

// Scenario: rejecting an invitation without a token is rejected.
func TestACLRejectInvitation_RequiresAuth(t *testing.T) {
	t.Run("When_reject_invitation_without_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenRejectInvitationNoAuth(t, ctx)

		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_auth_required", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/missing-header")
		})
	})
}
