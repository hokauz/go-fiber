package acl

import "github.com/gofiber/fiber/v3"

// --- Groups ---

// @Summary      Create group
// @Description  Creates an ACL group bound to a resource. Requires manage_members on the resource.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        body  body      CreateGroupRequest  true  "Group data"
// @Success      201   {object}  GroupResponse
// @Failure      404   {object}  fiber.Map
// @Failure      422   {object}  fiber.Map
// @Router       /api/v1/acl/groups [post]
// @Security     BearerAuth
func (h *Handler) CreateGroupDoc(c fiber.Ctx) error { return h.CreateGroup(c) }

// @Summary      List groups
// @Description  Lists ACL groups for a resource. Requires read on the resource.
// @Tags         ACL
// @Produce      json
// @Param        resource_type  query     string  true  "Resource type"
// @Param        resource_id    query     string  true  "Resource ID"
// @Success      200            {array}   GroupResponse
// @Failure      404            {object}  fiber.Map
// @Failure      422            {object}  fiber.Map
// @Router       /api/v1/acl/groups [get]
// @Security     BearerAuth
func (h *Handler) ListGroupsDoc(c fiber.Ctx) error { return h.ListGroups(c) }

// @Summary      Get group
// @Description  Returns a single ACL group. Requires read on the resource.
// @Tags         ACL
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  GroupResponse
// @Failure      404  {object}  fiber.Map
// @Router       /api/v1/acl/groups/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetGroupDoc(c fiber.Ctx) error { return h.GetGroup(c) }

// @Summary      Update group
// @Description  Updates the name or access_level of an ACL group. Requires manage_members on the resource.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        id    path      string              true  "Group ID"
// @Param        body  body      UpdateGroupRequest  true  "Fields to update"
// @Success      200   {object}  GroupResponse
// @Failure      404   {object}  fiber.Map
// @Failure      422   {object}  fiber.Map
// @Router       /api/v1/acl/groups/{id} [patch]
// @Security     BearerAuth
func (h *Handler) UpdateGroupDoc(c fiber.Ctx) error { return h.UpdateGroup(c) }

// @Summary      Delete group
// @Description  Soft-deletes a group, removes all its members, and cancels pending invitations.
// @Tags         ACL
// @Param        id  path  string  true  "Group ID"
// @Success      204
// @Failure      404  {object}  fiber.Map
// @Router       /api/v1/acl/groups/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteGroupDoc(c fiber.Ctx) error { return h.DeleteGroup(c) }

// --- Members ---

// @Summary      List members
// @Description  Lists active members of a group. Requires read on the resource.
// @Tags         ACL
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   MemberResponse
// @Failure      404  {object}  fiber.Map
// @Router       /api/v1/acl/groups/{id}/members [get]
// @Security     BearerAuth
func (h *Handler) ListMembersDoc(c fiber.Ctx) error { return h.ListMembers(c) }

// @Summary      Add member
// @Description  Directly adds a user to a group. Requires manage_members on the resource.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        id    path      string           true  "Group ID"
// @Param        body  body      AddMemberRequest true  "User to add"
// @Success      201   {object}  MemberResponse
// @Failure      404   {object}  fiber.Map
// @Failure      409   {object}  fiber.Map
// @Router       /api/v1/acl/groups/{id}/members [post]
// @Security     BearerAuth
func (h *Handler) AddMemberDoc(c fiber.Ctx) error { return h.AddMember(c) }

// @Summary      Remove member
// @Description  Soft-removes a member from a group. Requires manage_members on the resource.
// @Tags         ACL
// @Param        id         path  string  true  "Group ID"
// @Param        member_id  path  string  true  "Member ID"
// @Success      204
// @Failure      404  {object}  fiber.Map
// @Router       /api/v1/acl/groups/{id}/members/{member_id} [delete]
// @Security     BearerAuth
func (h *Handler) RemoveMemberDoc(c fiber.Ctx) error { return h.RemoveMember(c) }

// --- Invitations ---

// @Summary      Create invitation
// @Description  Creates an invitation link for a group. Returns the raw code once. Requires manage_members.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        body  body      CreateInvitationRequest  true  "Invitation data"
// @Success      201   {object}  CreateInvitationResponse
// @Failure      404   {object}  fiber.Map
// @Failure      422   {object}  fiber.Map
// @Router       /api/v1/acl/invitations [post]
// @Security     BearerAuth
func (h *Handler) CreateInvitationDoc(c fiber.Ctx) error { return h.CreateInvitation(c) }

// @Summary      List invitations
// @Description  Lists invitations for a resource. Requires manage_members on the resource.
// @Tags         ACL
// @Produce      json
// @Param        resource_type  query     string  true  "Resource type"
// @Param        resource_id    query     string  true  "Resource ID"
// @Success      200            {array}   InvitationResponse
// @Failure      404            {object}  fiber.Map
// @Router       /api/v1/acl/invitations [get]
// @Security     BearerAuth
func (h *Handler) ListInvitationsDoc(c fiber.Ctx) error { return h.ListInvitations(c) }

// @Summary      Check invitation
// @Description  Validates an invitation by raw code. Public endpoint — no authentication required.
// @Tags         ACL
// @Produce      json
// @Param        code  query     string  true  "Raw invitation code"
// @Success      200   {object}  InvitationCheckResponse
// @Router       /api/v1/acl/invitations/check [get]
func (h *Handler) CheckInvitationDoc(c fiber.Ctx) error { return h.CheckInvitation(c) }

// @Summary      Accept invitation
// @Description  Accepts an invitation and joins the group. Authenticated endpoint.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        body  body      InvitationCodeRequest  true  "Raw code"
// @Success      200   {object}  fiber.Map
// @Failure      409   {object}  fiber.Map
// @Failure      422   {object}  fiber.Map
// @Router       /api/v1/acl/invitations/accept [post]
// @Security     BearerAuth
func (h *Handler) AcceptInvitationDoc(c fiber.Ctx) error { return h.AcceptInvitation(c) }

// @Summary      Reject invitation
// @Description  Rejects a pending invitation. Authenticated endpoint.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        body  body      InvitationCodeRequest  true  "Raw code"
// @Success      200   {object}  fiber.Map
// @Failure      422   {object}  fiber.Map
// @Router       /api/v1/acl/invitations/reject [post]
// @Security     BearerAuth
func (h *Handler) RejectInvitationDoc(c fiber.Ctx) error { return h.RejectInvitation(c) }

// @Summary      Cancel invitation
// @Description  Cancels a pending invitation. Requires manage_members on the resource.
// @Tags         ACL
// @Param        id  path  string  true  "Invitation ID"
// @Success      200  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      422  {object}  fiber.Map
// @Router       /api/v1/acl/invitations/{id}/cancel [post]
// @Security     BearerAuth
func (h *Handler) CancelInvitationDoc(c fiber.Ctx) error { return h.CancelInvitation(c) }

// --- Access ---

// @Summary      Get effective access level
// @Description  Returns the caller's effective access_level for a resource, or 404 if none.
// @Tags         ACL
// @Produce      json
// @Param        resource_type  query     string  true  "Resource type"
// @Param        resource_id    query     string  true  "Resource ID"
// @Success      200            {object}  fiber.Map
// @Failure      404            {object}  fiber.Map
// @Router       /api/v1/acl/access/level [get]
// @Security     BearerAuth
func (h *Handler) GetEffectiveAccessLevelDoc(c fiber.Ctx) error { return h.GetEffectiveAccessLevel(c) }

// --- Delete requests ---

// @Summary      Create delete request
// @Description  Creates a delete request for a resource. May execute immediately depending on policy.
// @Tags         ACL
// @Accept       json
// @Produce      json
// @Param        body  body      CreateDeleteRequestRequest  true  "Delete request data"
// @Success      201   {object}  DeleteRequestResponse
// @Success      202   {object}  DeleteRequestResponse
// @Failure      404   {object}  fiber.Map
// @Failure      422   {object}  fiber.Map
// @Router       /api/v1/acl/delete-requests [post]
// @Security     BearerAuth
func (h *Handler) CreateDeleteRequestDoc(c fiber.Ctx) error { return h.CreateDeleteRequest(c) }

// @Summary      List delete requests
// @Description  Lists delete requests for a resource. Requires manage_members on the resource.
// @Tags         ACL
// @Produce      json
// @Param        resource_type  query     string  true   "Resource type"
// @Param        resource_id    query     string  true   "Resource ID"
// @Param        requested_by   query     string  false  "Filter by requester"
// @Param        status         query     string  false  "Filter by status"
// @Param        limit          query     int     false  "Max rows"
// @Success      200            {array}   DeleteRequestResponse
// @Failure      404            {object}  fiber.Map
// @Failure      422            {object}  fiber.Map
// @Router       /api/v1/acl/delete-requests [get]
// @Security     BearerAuth
func (h *Handler) ListDeleteRequestsDoc(c fiber.Ctx) error { return h.ListDeleteRequests(c) }

// @Summary      Approve delete request
// @Description  Approves a pending delete request. Owner only.
// @Tags         ACL
// @Produce      json
// @Param        id   path      string  true  "Delete request ID"
// @Success      200  {object}  DeleteRequestResponse
// @Failure      404  {object}  fiber.Map
// @Failure      422  {object}  fiber.Map
// @Router       /api/v1/acl/delete-requests/{id}/approve [post]
// @Security     BearerAuth
func (h *Handler) ApproveDeleteRequestDoc(c fiber.Ctx) error { return h.ApproveDeleteRequest(c) }

// @Summary      Reject delete request
// @Description  Rejects a pending delete request. Owner only.
// @Tags         ACL
// @Produce      json
// @Param        id   path      string  true  "Delete request ID"
// @Success      200  {object}  DeleteRequestResponse
// @Failure      404  {object}  fiber.Map
// @Failure      422  {object}  fiber.Map
// @Router       /api/v1/acl/delete-requests/{id}/reject [post]
// @Security     BearerAuth
func (h *Handler) RejectDeleteRequestDoc(c fiber.Ctx) error { return h.RejectDeleteRequest(c) }

// --- Activity logs ---

// @Summary      List activity logs
// @Description  Lists ACL audit events for a resource. Requires manage_members on the resource.
// @Tags         ACL
// @Produce      json
// @Param        resource_type  query     string  true   "Resource type"
// @Param        resource_id    query     string  true   "Resource ID"
// @Param        actor_user_id  query     string  false  "Filter by actor"
// @Param        action         query     string  false  "Filter by action"
// @Param        from           query     string  false  "RFC3339 start timestamp"
// @Param        to             query     string  false  "RFC3339 end timestamp"
// @Param        limit          query     int     false  "Max rows"
// @Success      200            {array}   ActivityLogResponse
// @Failure      404            {object}  fiber.Map
// @Failure      422            {object}  fiber.Map
// @Router       /api/v1/acl/activity-logs [get]
// @Security     BearerAuth
func (h *Handler) ListActivityLogsDoc(c fiber.Ctx) error { return h.ListActivityLogs(c) }
