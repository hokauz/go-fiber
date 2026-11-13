package acl

import (
	"encoding/json"
	"strconv"
	"time"

	pkgacl "go-fiber/pkg/acl"
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
)

// Handler handles ACL HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler returns a new Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) handleError(c fiber.Ctx, err error) error {
	status, code, message := MapToHTTP(err)
	return c.Status(status).JSON(fiber.Map{
		"error":   code,
		"message": message,
	})
}

func toGroupResponse(g *Group) GroupResponse {
	return GroupResponse{
		ID:           g.ID,
		ResourceType: g.ResourceType,
		ResourceID:   g.ResourceID,
		Name:         g.Name,
		AccessLevel:  g.AccessLevel,
		CreatedBy:    g.CreatedBy,
		CreatedAt:    g.CreatedAt,
	}
}

func toMemberResponse(m *GroupMember) MemberResponse {
	return MemberResponse{
		ID:         m.ID,
		GroupID:    m.GroupID,
		UserID:     m.UserID,
		SourceType: m.SourceType,
		CreatedAt:  m.CreatedAt,
	}
}

func toInvitationResponse(inv *Invitation) InvitationResponse {
	return InvitationResponse{
		ID:            inv.ID,
		GroupID:       inv.GroupID,
		Status:        inv.Status,
		CodeExpiresAt: inv.CodeExpiresAt,
		CreatedAt:     inv.CreatedAt,
	}
}

func toDeleteRequestResponse(dr *DeleteRequest) DeleteRequestResponse {
	return DeleteRequestResponse{
		ID:           dr.ID,
		ResourceType: dr.ResourceType,
		ResourceID:   dr.ResourceID,
		RequestedBy:  dr.RequestedBy,
		Status:       dr.Status,
		Reason:       dr.Reason,
		CreatedAt:    dr.CreatedAt,
		UpdatedAt:    dr.UpdatedAt,
		ApprovedBy:   dr.ApprovedBy,
		ApprovedAt:   dr.ApprovedAt,
		RejectedBy:   dr.RejectedBy,
		RejectedAt:   dr.RejectedAt,
		ExecutedAt:   dr.ExecutedAt,
		ExpiresAt:    dr.ExpiresAt,
	}
}

func toActivityLogResponse(entry *ActivityLog) ActivityLogResponse {
	metadata := map[string]any{}
	if len(entry.Metadata) > 0 {
		_ = json.Unmarshal(entry.Metadata, &metadata)
	}
	return ActivityLogResponse{
		ID:            entry.ID,
		ActorUserID:   entry.ActorUserID,
		EntityType:    entry.EntityType,
		EntityID:      entry.EntityID,
		Action:        entry.Action,
		ResourceType:  entry.ResourceType,
		ResourceID:    entry.ResourceID,
		Metadata:      metadata,
		CorrelationID: entry.CorrelationID,
		CreatedAt:     entry.CreatedAt,
	}
}

// --- Groups ---

func (h *Handler) CreateGroup(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req CreateGroupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	g, err := h.svc.CreateGroup(c.Context(), p.UserID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toGroupResponse(g))
}

func (h *Handler) ListGroups(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	if resourceType == "" || resourceID == "" {
		return h.handleError(c, errValidation("resource_type and resource_id are required"))
	}
	groups, err := h.svc.ListGroups(c.Context(), p.UserID, resourceType, resourceID)
	if err != nil {
		return h.handleError(c, err)
	}
	resp := make([]GroupResponse, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, toGroupResponse(g))
	}
	return c.JSON(resp)
}

func (h *Handler) GetGroup(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	g, err := h.svc.GetGroup(c.Context(), p.UserID, c.Params("id"))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(toGroupResponse(g))
}

func (h *Handler) UpdateGroup(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req UpdateGroupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	g, err := h.svc.UpdateGroup(c.Context(), p.UserID, c.Params("id"), req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(toGroupResponse(g))
}

func (h *Handler) DeleteGroup(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	if err := h.svc.DeleteGroup(c.Context(), p.UserID, c.Params("id")); err != nil {
		return h.handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Members ---

func (h *Handler) ListMembers(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	members, err := h.svc.ListMembers(c.Context(), p.UserID, c.Params("id"))
	if err != nil {
		return h.handleError(c, err)
	}
	resp := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, toMemberResponse(m))
	}
	return c.JSON(resp)
}

func (h *Handler) AddMember(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req AddMemberRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	m, err := h.svc.AddMember(c.Context(), p.UserID, c.Params("id"), req.UserID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toMemberResponse(m))
}

func (h *Handler) RemoveMember(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	if err := h.svc.RemoveMember(c.Context(), p.UserID, c.Params("id"), c.Params("member_id")); err != nil {
		return h.handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Invitations ---

func (h *Handler) CreateInvitation(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req CreateInvitationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	inv, rawCode, err := h.svc.CreateInvitation(c.Context(), p.UserID, req.GroupID, req.ExpiresInDays)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(CreateInvitationResponse{
		InvitationResponse: toInvitationResponse(inv),
		Code:               rawCode,
	})
}

func (h *Handler) ListInvitations(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	if resourceType == "" || resourceID == "" {
		return h.handleError(c, errValidation("resource_type and resource_id are required"))
	}
	invitations, err := h.svc.ListInvitations(c.Context(), p.UserID, resourceType, resourceID)
	if err != nil {
		return h.handleError(c, err)
	}
	resp := make([]InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		resp = append(resp, toInvitationResponse(inv))
	}
	return c.JSON(resp)
}

func (h *Handler) CheckInvitation(c fiber.Ctx) error {
	code := c.Query("code")
	result, err := h.svc.CheckInvitation(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(result)
}

func (h *Handler) AcceptInvitation(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req InvitationCodeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	if err := h.svc.AcceptInvitation(c.Context(), p.UserID, req.Code); err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(fiber.Map{"status": "accepted"})
}

func (h *Handler) RejectInvitation(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req InvitationCodeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	if err := h.svc.RejectInvitation(c.Context(), p.UserID, req.Code); err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(fiber.Map{"status": "rejected"})
}

func (h *Handler) CancelInvitation(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	if err := h.svc.CancelInvitation(c.Context(), p.UserID, c.Params("id")); err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(fiber.Map{"status": "cancelled"})
}

// --- Access ---

func (h *Handler) GetEffectiveAccessLevel(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	if resourceType == "" || resourceID == "" {
		return h.handleError(c, errValidation("resource_type and resource_id are required"))
	}
	accessLevel, err := h.svc.GetEffectiveAccessLevel(c.Context(), p.UserID, resourceType, resourceID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(fiber.Map{"access_level": accessLevel})
}

// --- Delete requests ---

func (h *Handler) CreateDeleteRequest(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	var req CreateDeleteRequestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return h.handleError(c, errValidation("invalid request body"))
	}
	dr, err := h.svc.CreateDeleteRequest(c.Context(), p.UserID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	status := fiber.StatusCreated
	if dr.Status == "pending" {
		status = fiber.StatusAccepted
	}
	return c.Status(status).JSON(toDeleteRequestResponse(dr))
}

func (h *Handler) ListDeleteRequests(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	if resourceType == "" || resourceID == "" {
		return h.handleError(c, errValidation("resource_type and resource_id are required"))
	}
	requestedBy := nullableQuery(c, "requested_by")
	status := nullableQuery(c, "status")
	limit := parseLimit(c.Query("limit"), 100)
	drs, err := h.svc.ListDeleteRequests(c.Context(), p.UserID, resourceType, resourceID, requestedBy, status, limit)
	if err != nil {
		return h.handleError(c, err)
	}
	resp := make([]DeleteRequestResponse, 0, len(drs))
	for _, dr := range drs {
		resp = append(resp, toDeleteRequestResponse(dr))
	}
	return c.JSON(resp)
}

func (h *Handler) ApproveDeleteRequest(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	dr, err := h.svc.ApproveDeleteRequest(c.Context(), p.UserID, c.Params("id"))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(toDeleteRequestResponse(dr))
}

func (h *Handler) RejectDeleteRequest(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	dr, err := h.svc.RejectDeleteRequest(c.Context(), p.UserID, c.Params("id"))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(toDeleteRequestResponse(dr))
}

// --- Activity logs ---

func (h *Handler) ListActivityLogs(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	if resourceType == "" || resourceID == "" {
		return h.handleError(c, errValidation("resource_type and resource_id are required"))
	}
	actorUserID := nullableQuery(c, "actor_user_id")
	action := nullableQuery(c, "action")
	from, err := parseTimeQuery(c.Query("from"))
	if err != nil {
		return h.handleError(c, errValidation("invalid from time; use RFC3339"))
	}
	to, err := parseTimeQuery(c.Query("to"))
	if err != nil {
		return h.handleError(c, errValidation("invalid to time; use RFC3339"))
	}
	limit := parseLimit(c.Query("limit"), 200)
	entries, err := h.svc.ListActivityLogs(c.Context(), p.UserID, resourceType, resourceID, actorUserID, action, from, to, limit)
	if err != nil {
		return h.handleError(c, err)
	}
	resp := make([]ActivityLogResponse, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, toActivityLogResponse(entry))
	}
	return c.JSON(resp)
}

func nullableQuery(c fiber.Ctx, key string) *string {
	value := c.Query(key)
	if value == "" {
		return nil
	}
	return &value
}

func parseTimeQuery(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseLimit(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return defaultValue
	}
	if n > 1000 {
		return 1000
	}
	return n
}

// errValidation returns a validation ACLError with a custom message.
func errValidation(msg string) error {
	return &pkgacl.ACLError{Code: "validation-error", Message: msg}
}
