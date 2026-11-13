//go:build e2e

package steps

import (
	"fmt"
	"net/http"
	"testing"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
)

// --- Groups ---

// WhenCreateGroup performs POST /api/v1/acl/groups with a bearer token.
func WhenCreateGroup(t *testing.T, ctx *e2e.TestContext, token, resourceType, resourceID, name, accessLevel string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/groups", map[string]any{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"name":          name,
		"access_level":  accessLevel,
	}, helpers.AuthHeaders(token))
}

// WhenCreateGroupNoAuth performs POST /api/v1/acl/groups without authentication.
func WhenCreateGroupNoAuth(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/groups", map[string]any{
		"resource_type": "account",
		"resource_id":   helpers.MustUUID(),
		"name":          "group",
		"access_level":  "view",
	}, nil)
}

// WhenGetGroup performs GET /api/v1/acl/groups/:id with a bearer token.
func WhenGetGroup(t *testing.T, ctx *e2e.TestContext, token, groupID string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodGet, "/api/v1/acl/groups/"+groupID, nil, helpers.AuthHeaders(token))
}

// WhenListGroups performs GET /api/v1/acl/groups?resource_type=...&resource_id=... with a bearer token.
func WhenListGroups(t *testing.T, ctx *e2e.TestContext, token, resourceType, resourceID string) *http.Response {
	t.Helper()
	path := fmt.Sprintf("/api/v1/acl/groups?resource_type=%s&resource_id=%s", resourceType, resourceID)
	return ctx.JSONRequest(http.MethodGet, path, nil, helpers.AuthHeaders(token))
}

// WhenListGroupsNoAuth performs GET /api/v1/acl/groups without authentication.
func WhenListGroupsNoAuth(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodGet, "/api/v1/acl/groups?resource_type=account&resource_id="+helpers.MustUUID(), nil, nil)
}

// WhenDeleteGroup performs DELETE /api/v1/acl/groups/:id with a bearer token.
func WhenDeleteGroup(t *testing.T, ctx *e2e.TestContext, token, groupID string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodDelete, "/api/v1/acl/groups/"+groupID, nil, helpers.AuthHeaders(token))
}

// --- Members ---

// WhenAddMember performs POST /api/v1/acl/groups/:id/members with a bearer token.
func WhenAddMember(t *testing.T, ctx *e2e.TestContext, token, groupID, userID string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/groups/"+groupID+"/members", map[string]any{
		"user_id": userID,
	}, helpers.AuthHeaders(token))
}

// WhenRemoveMember performs DELETE /api/v1/acl/groups/:id/members/:member_id with a bearer token.
func WhenRemoveMember(t *testing.T, ctx *e2e.TestContext, token, groupID, memberID string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodDelete, "/api/v1/acl/groups/"+groupID+"/members/"+memberID, nil, helpers.AuthHeaders(token))
}

// --- Invitations ---

// WhenCreateInvitation performs POST /api/v1/acl/invitations with a bearer token.
func WhenCreateInvitation(t *testing.T, ctx *e2e.TestContext, token, groupID string, expiresInDays int) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations", map[string]any{
		"group_id":        groupID,
		"expires_in_days": expiresInDays,
	}, helpers.AuthHeaders(token))
}

// WhenCreateInvitationNoAuth performs POST /api/v1/acl/invitations without authentication.
func WhenCreateInvitationNoAuth(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations", map[string]any{
		"group_id":        helpers.MustUUID(),
		"expires_in_days": 7,
	}, nil)
}

// WhenCheckInvitation performs GET /api/v1/acl/invitations/check?code=... (public, no auth).
func WhenCheckInvitation(t *testing.T, ctx *e2e.TestContext, code string) *http.Response {
	t.Helper()
	return ctx.Get("/api/v1/acl/invitations/check?code=" + code)
}

// WhenCheckInvitationNoCode performs GET /api/v1/acl/invitations/check without a code param.
func WhenCheckInvitationNoCode(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.Get("/api/v1/acl/invitations/check")
}

// WhenAcceptInvitation performs POST /api/v1/acl/invitations/accept with a bearer token.
func WhenAcceptInvitation(t *testing.T, ctx *e2e.TestContext, token, code string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations/accept", map[string]any{
		"code": code,
	}, helpers.AuthHeaders(token))
}

// WhenAcceptInvitationNoAuth performs POST /api/v1/acl/invitations/accept without authentication.
func WhenAcceptInvitationNoAuth(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations/accept", map[string]any{
		"code": "irrelevant",
	}, nil)
}

// WhenRejectInvitation performs POST /api/v1/acl/invitations/reject with a bearer token.
func WhenRejectInvitation(t *testing.T, ctx *e2e.TestContext, token, code string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations/reject", map[string]any{
		"code": code,
	}, helpers.AuthHeaders(token))
}

// WhenRejectInvitationNoAuth performs POST /api/v1/acl/invitations/reject without authentication.
func WhenRejectInvitationNoAuth(t *testing.T, ctx *e2e.TestContext) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations/reject", map[string]any{
		"code": "irrelevant",
	}, nil)
}

// WhenCancelInvitation performs POST /api/v1/acl/invitations/:id/cancel with a bearer token.
func WhenCancelInvitation(t *testing.T, ctx *e2e.TestContext, token, invitationID string) *http.Response {
	t.Helper()
	return ctx.JSONRequest(http.MethodPost, "/api/v1/acl/invitations/"+invitationID+"/cancel", nil, helpers.AuthHeaders(token))
}

// --- Access ---

// WhenGetEffectiveAccessLevel performs GET /api/v1/acl/access/level?... with a bearer token.
func WhenGetEffectiveAccessLevel(t *testing.T, ctx *e2e.TestContext, token, resourceType, resourceID string) *http.Response {
	t.Helper()
	path := fmt.Sprintf("/api/v1/acl/access/level?resource_type=%s&resource_id=%s", resourceType, resourceID)
	return ctx.JSONRequest(http.MethodGet, path, nil, helpers.AuthHeaders(token))
}

// WhenGetEffectiveRole keeps backward-compat naming in e2e helpers.
func WhenGetEffectiveRole(t *testing.T, ctx *e2e.TestContext, token, resourceType, resourceID string) *http.Response {
	t.Helper()
	return WhenGetEffectiveAccessLevel(t, ctx, token, resourceType, resourceID)
}
