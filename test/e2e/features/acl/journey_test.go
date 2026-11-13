//go:build e2e

package acl

import (
	"net/http"
	"os"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
	"go-fiber/test/e2e/internal/steps"
)

// journeyEnv holds credentials and resource info required for journey tests.
type journeyEnv struct {
	ownerToken   string
	memberToken  string
	resourceType string
	resourceID   string
}

// loadJourneyEnv reads journey env vars and reports whether they are all set.
// Tests call t.Skip when ok is false.
func loadJourneyEnv(t *testing.T) (journeyEnv, bool) {
	t.Helper()
	env := journeyEnv{
		ownerToken:   os.Getenv("E2E_OWNER_TOKEN"),
		memberToken:  os.Getenv("E2E_MEMBER_TOKEN"),
		resourceType: os.Getenv("E2E_RESOURCE_TYPE"),
		resourceID:   os.Getenv("E2E_RESOURCE_ID"),
	}
	if env.resourceType == "" {
		env.resourceType = "account"
	}
	ok := env.ownerToken != "" && env.memberToken != "" && env.resourceID != ""
	return env, ok
}

// mustCreateGroup is a setup helper that creates a group and extracts the ID.
// Calls t.FailNow on any failure so dependent steps are skipped automatically.
func mustCreateGroup(t *testing.T, ctx *e2e.TestContext, token, resourceType, resourceID, name, accessLevel string) string {
	t.Helper()
	resp := steps.WhenCreateGroup(t, ctx, token, resourceType, resourceID, name, accessLevel)
	helpers.ThenStatusIs(t, resp, http.StatusCreated)
	m := helpers.MustParseJSON(t, resp)
	id, _ := m["id"].(string)
	if id == "" {
		t.Fatal("group id not present in create response")
	}
	return id
}

// mustCreateInvitation is a setup helper that creates an invitation and returns the raw code.
func mustCreateInvitation(t *testing.T, ctx *e2e.TestContext, token, groupID string) string {
	t.Helper()
	resp := steps.WhenCreateInvitation(t, ctx, token, groupID, 7)
	helpers.ThenStatusIs(t, resp, http.StatusCreated)
	m := helpers.MustParseJSON(t, resp)
	code, _ := m["code"].(string)
	if code == "" {
		t.Fatal("code not present in create invitation response")
	}
	return code
}

// mustGetInvitationID is a setup helper that creates an invitation and returns both the invitation ID and code.
func mustGetInvitationID(t *testing.T, ctx *e2e.TestContext, token, groupID string) (invID, code string) {
	t.Helper()
	resp := steps.WhenCreateInvitation(t, ctx, token, groupID, 7)
	helpers.ThenStatusIs(t, resp, http.StatusCreated)
	m := helpers.MustParseJSON(t, resp)
	invID, _ = m["id"].(string)
	code, _ = m["code"].(string)
	if invID == "" || code == "" {
		t.Fatal("id or code not present in create invitation response")
	}
	return
}

// Scenario: owner creates a group, issues an invite; member checks, accepts, and gains access;
// owner then removes the member and access is revoked.
//
// Requires: E2E_OWNER_TOKEN, E2E_MEMBER_TOKEN, E2E_RESOURCE_ID (and optionally E2E_RESOURCE_TYPE).
func TestACLJourney_OwnerInvitesMember(t *testing.T) {
	env, ok := loadJourneyEnv(t)
	if !ok {
		t.Skip("E2E_OWNER_TOKEN, E2E_MEMBER_TOKEN, E2E_RESOURCE_ID not set — skipping journey test")
	}
	ctx := e2e.NewTestContext(t)

	// Setup
	groupID := mustCreateGroup(t, ctx, env.ownerToken, env.resourceType, env.resourceID, "Equipe Financeira", "edit")
	invCode := mustCreateInvitation(t, ctx, env.ownerToken, groupID)

	// Step 1: any user can check the invitation before authenticating
	t.Run("When_anyone_checks_the_pending_invitation", func(t *testing.T) {
		resp := steps.WhenCheckInvitation(t, ctx, invCode)

		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_invitation_is_valid", func(t *testing.T) {
			steps.AndInvitationIsValid(t, resp)
		})
	})

	// Step 2: authenticated member accepts the invitation
	t.Run("When_member_accepts_the_invitation", func(t *testing.T) {
		resp := steps.WhenAcceptInvitation(t, ctx, env.memberToken, invCode)

		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_status_is_accepted", func(t *testing.T) {
			helpers.AndJSONFieldEquals(t, resp, "status", "accepted")
		})
	})

	// Step 3: member now has an effective access level on the resource
	t.Run("When_member_requests_effective_access_level", func(t *testing.T) {
		resp := steps.WhenGetEffectiveAccessLevel(t, ctx, env.memberToken, env.resourceType, env.resourceID)

		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_access_level_is_edit", func(t *testing.T) {
			steps.AndAccessLevelIs(t, resp, "edit")
		})
	})

	// Step 4: check invitation code again — it is now consumed
	t.Run("When_anyone_checks_the_accepted_invitation", func(t *testing.T) {
		resp := steps.WhenCheckInvitation(t, ctx, invCode)

		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_invitation_is_no_longer_valid", func(t *testing.T) {
			steps.AndInvitationIsInvalid(t, resp)
		})
	})

	// Step 5: find member ID and remove
	var memberID string
	t.Run("When_owner_lists_group_members", func(t *testing.T) {
		resp := ctx.JSONRequest(http.MethodGet, "/api/v1/acl/groups/"+groupID+"/members", nil, helpers.AuthHeaders(env.ownerToken))
		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
		t.Run("And_member_is_listed", func(t *testing.T) {
			var list []map[string]any
			helpers.MustDecodeJSON(t, resp, &list)
			if len(list) == 0 {
				t.Fatal("expected at least one member in group")
			}
			memberID, _ = list[0]["id"].(string)
		})
	})

	if memberID == "" {
		t.Fatal("could not resolve member ID — cannot continue removal step")
	}

	t.Run("When_owner_removes_the_member", func(t *testing.T) {
		resp := steps.WhenRemoveMember(t, ctx, env.ownerToken, groupID, memberID)

		t.Run("Then_no_content", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNoContent)
		})
	})

	// Step 6: after removal, member no longer has effective access level
	t.Run("When_removed_member_requests_effective_access_level", func(t *testing.T) {
		resp := steps.WhenGetEffectiveAccessLevel(t, ctx, env.memberToken, env.resourceType, env.resourceID)

		t.Run("Then_not_found", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNotFound)
		})
	})
}

// Scenario: a user with no ownership or membership receives 404 when trying to manage a resource.
// This validates the "not-found instead of forbidden" policy that avoids leaking resource existence.
func TestACLJourney_UnauthorizedUserGets404(t *testing.T) {
	env, ok := loadJourneyEnv(t)
	if !ok {
		t.Skip("E2E_OWNER_TOKEN, E2E_MEMBER_TOKEN, E2E_RESOURCE_ID not set — skipping journey test")
	}
	ctx := e2e.NewTestContext(t)

	// memberToken has no relationship to the resource yet
	t.Run("When_unauthorized_user_tries_to_create_a_group", func(t *testing.T) {
		resp := steps.WhenCreateGroup(t, ctx, env.memberToken, env.resourceType, env.resourceID, "Shadow Group", "manage_access")

		t.Run("Then_not_found", func(t *testing.T) {
			// 404 — not 403 — so resource existence is not leaked
			helpers.ThenStatusIs(t, resp, http.StatusNotFound)
		})
		t.Run("And_error_not_found", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "not-found")
		})
	})
}

// Scenario: accepting the same invitation twice results in a conflict.
// The SELECT FOR UPDATE in AcceptInvitationTx ensures only one acceptance wins.
func TestACLJourney_DuplicateAcceptConflicts(t *testing.T) {
	env, ok := loadJourneyEnv(t)
	if !ok {
		t.Skip("E2E_OWNER_TOKEN, E2E_MEMBER_TOKEN, E2E_RESOURCE_ID not set — skipping journey test")
	}
	ctx := e2e.NewTestContext(t)

	groupID := mustCreateGroup(t, ctx, env.ownerToken, env.resourceType, env.resourceID, "Duplicate Test Group", "view")
	invCode := mustCreateInvitation(t, ctx, env.ownerToken, groupID)

	// First accept — must succeed
	t.Run("When_member_accepts_invitation_first_time", func(t *testing.T) {
		resp := steps.WhenAcceptInvitation(t, ctx, env.memberToken, invCode)
		t.Run("Then_ok", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusOK)
		})
	})

	// Second accept of the same code — must conflict
	t.Run("When_member_tries_to_accept_same_invitation_again", func(t *testing.T) {
		resp := steps.WhenAcceptInvitation(t, ctx, env.memberToken, invCode)

		t.Run("Then_conflict", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusConflict)
		})
		t.Run("And_error_already_consumed", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "acl/invite-already-consumed")
		})
	})
}

// Scenario: deleting a group with pending invitations succeeds and pending invitations are cancelled.
func TestACLJourney_GroupDeleteCancelsPendingInvites(t *testing.T) {
	env, ok := loadJourneyEnv(t)
	if !ok {
		t.Skip("E2E_OWNER_TOKEN, E2E_MEMBER_TOKEN, E2E_RESOURCE_ID not set — skipping journey test")
	}
	ctx := e2e.NewTestContext(t)

	groupID := mustCreateGroup(t, ctx, env.ownerToken, env.resourceType, env.resourceID, "Pending Invite Group", "view")
	invID, _ := mustGetInvitationID(t, ctx, env.ownerToken, groupID)

	t.Run("When_owner_deletes_group_with_pending_invite", func(t *testing.T) {
		resp := steps.WhenDeleteGroup(t, ctx, env.ownerToken, groupID)

		t.Run("Then_no_content", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNoContent)
		})
	})

	// Pending invite should no longer be cancellable after group deletion.
	t.Run("When_owner_tries_to_cancel_old_pending_invite", func(t *testing.T) {
		resp := steps.WhenCancelInvitation(t, ctx, env.ownerToken, invID)
		t.Run("Then_not_found", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNotFound)
		})
	})
}

// Scenario: malformed UUIDs in ACL path params return not-found semantics.
func TestACLJourney_InvalidUUIDPathReturnsNotFound(t *testing.T) {
	env, ok := loadJourneyEnv(t)
	if !ok {
		t.Skip("E2E_OWNER_TOKEN, E2E_MEMBER_TOKEN, E2E_RESOURCE_ID not set — skipping journey test")
	}
	ctx := e2e.NewTestContext(t)

	t.Run("When_get_group_with_invalid_uuid", func(t *testing.T) {
		resp := steps.WhenGetGroup(t, ctx, env.ownerToken, "not-a-uuid")
		t.Run("Then_not_found", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNotFound)
		})
		t.Run("And_error_not_found", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "not-found")
		})
	})

	t.Run("When_delete_group_with_invalid_uuid", func(t *testing.T) {
		resp := steps.WhenDeleteGroup(t, ctx, env.ownerToken, "not-a-uuid")
		t.Run("Then_not_found", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNotFound)
		})
		t.Run("And_error_not_found", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "not-found")
		})
	})

	t.Run("When_cancel_invitation_with_invalid_uuid", func(t *testing.T) {
		resp := steps.WhenCancelInvitation(t, ctx, env.ownerToken, "not-a-uuid")
		t.Run("Then_not_found", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusNotFound)
		})
		t.Run("And_error_not_found", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "not-found")
		})
	})
}
