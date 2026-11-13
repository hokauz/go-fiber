package acl

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	pkgacl "go-fiber/pkg/acl"
)

const defaultDeleteRequestTTL = 7 * 24 * time.Hour
const defaultInvitationTTLDays = 7

type serviceRepository interface {
	CreateGroup(ctx context.Context, g *Group) (*Group, error)
	GetGroupByID(ctx context.Context, id string) (*Group, error)
	ListGroupsByResource(ctx context.Context, resourceType, resourceID string) ([]*Group, error)
	UpdateGroup(ctx context.Context, id, name, accessLevel string) (*Group, error)
	SoftDeleteGroup(ctx context.Context, groupID string) error

	InsertMember(ctx context.Context, m *GroupMember) (*GroupMember, error)
	ListMembersByGroup(ctx context.Context, groupID string) ([]*GroupMember, error)
	GetMemberByID(ctx context.Context, memberID string) (*GroupMember, error)
	SoftRemoveMember(ctx context.Context, memberID, groupID string) error

	InsertInvitation(ctx context.Context, inv *Invitation) (*Invitation, error)
	GetInvitationByCodeHash(ctx context.Context, hash string) (*Invitation, error)
	GetInvitationByID(ctx context.Context, id string) (*Invitation, error)
	ListInvitationsByResource(ctx context.Context, resourceType, resourceID string) ([]*Invitation, error)
	AcceptInvitationTx(ctx context.Context, codeHash, actorID, memberID string) error
	RejectInvitationTx(ctx context.Context, codeHash string) error
	CancelInvitationTx(ctx context.Context, invitationID string) error

	GetEffectiveAccessLevel(ctx context.Context, userID, resourceType, resourceID string) (string, error)
	ListAccessibleResourceAccessLevels(ctx context.Context, userID, resourceType string) ([]ResourceAccessLevel, error)

	InsertDeleteRequest(ctx context.Context, dr *DeleteRequest) (*DeleteRequest, error)
	GetDeleteRequestByID(ctx context.Context, id string) (*DeleteRequest, error)
	ListDeleteRequestsByResource(ctx context.Context, resourceType, resourceID string, requestedBy, status *string, limit int) ([]*DeleteRequest, error)
	ApproveDeleteRequestTx(ctx context.Context, requestID, approvedBy string) (*DeleteRequest, error)
	RejectDeleteRequestTx(ctx context.Context, requestID, rejectedBy string) (*DeleteRequest, error)
	MarkDeleteRequestExecutedTx(ctx context.Context, requestID string) (*DeleteRequest, error)

	InsertActivityLog(ctx context.Context, log *ActivityLog) error
	ListActivityLogsByResource(ctx context.Context, resourceType, resourceID string, actorUserID, action *string, from, to *time.Time, limit int) ([]*ActivityLog, error)
}

// Service orchestrates ACL use-cases and implements pkgacl.Authorizer.
type Service struct {
	repo                 serviceRepository
	ownerResolver        pkgacl.OwnerResolver // nil when used for public-only endpoints
	deletePolicyResolver DeletePolicyResolver
	deleteExecutor       ResourceDeleteExecutor
}

// NewService returns a Service. ownerResolver may be nil for public-only usage.
func NewService(repo serviceRepository, ownerResolver pkgacl.OwnerResolver, deletePolicyResolver DeletePolicyResolver, deleteExecutor ResourceDeleteExecutor) *Service {
	if deletePolicyResolver == nil {
		deletePolicyResolver = NewStaticDeletePolicyResolver(DeletePolicyDirectDelete, nil)
	}
	return &Service{
		repo:                 repo,
		ownerResolver:        ownerResolver,
		deletePolicyResolver: deletePolicyResolver,
		deleteExecutor:       deleteExecutor,
	}
}

// --- pkgacl.Authorizer ---

func (s *Service) CanAccess(ctx context.Context, userID, resourceType, resourceID, permission string) (bool, error) {
	if s.ownerResolver == nil {
		return false, pkgacl.ErrInternalError
	}

	ownerID, err := s.ownerResolver.GetOwner(ctx, resourceType, resourceID)
	if err != nil {
		return false, nil // resource not found or resolver error → not authorized
	}
	if userID == ownerID {
		return true, nil
	}

	accessLevel, err := s.repo.GetEffectiveAccessLevel(ctx, userID, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, pkgacl.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if !pkgacl.CanAccessLevel(accessLevel, permission) {
		return false, nil
	}

	if permission == pkgacl.PermDelete {
		policy, err := s.getDeletePolicy(ctx, resourceType)
		if err != nil {
			return false, err
		}
		if policy == DeletePolicyOwnerApprovalNeeded {
			return false, nil
		}
	}

	return true, nil
}

func (s *Service) GetEffectiveAccessLevel(ctx context.Context, userID, resourceType, resourceID string) (string, error) {
	if s.ownerResolver != nil {
		ownerID, err := s.ownerResolver.GetOwner(ctx, resourceType, resourceID)
		if err == nil && ownerID == userID {
			return pkgacl.AccessLevelOwner, nil
		}
	}
	return s.repo.GetEffectiveAccessLevel(ctx, userID, resourceType, resourceID)
}

func (s *Service) ListAccessibleResourceIDs(ctx context.Context, userID, resourceType, permission string) ([]string, error) {
	pairs, err := s.repo.ListAccessibleResourceAccessLevels(ctx, userID, resourceType)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, p := range pairs {
		if !pkgacl.CanAccessLevel(p.EffectiveAccessLevel, permission) {
			continue
		}
		if permission == pkgacl.PermDelete {
			policy, err := s.getDeletePolicy(ctx, resourceType)
			if err != nil {
				return nil, err
			}
			if policy == DeletePolicyOwnerApprovalNeeded {
				continue
			}
		}
		ids = append(ids, p.ResourceID)
	}
	return ids, nil
}

// --- Groups ---

func (s *Service) CreateGroup(ctx context.Context, actorID string, req CreateGroupRequest) (*Group, error) {
	if req.ResourceType == "" {
		return nil, pkgacl.ErrResourceTypeInvalid
	}
	if req.Name == "" {
		return nil, pkgacl.ErrValidation
	}
	if !ValidAccessLevel(req.AccessLevel) {
		return nil, pkgacl.ErrAccessLevelInvalid
	}
	if err := s.requirePerm(ctx, actorID, req.ResourceType, req.ResourceID, pkgacl.PermManageMembers); err != nil {
		return nil, err
	}
	id, err := newUUID()
	if err != nil {
		return nil, pkgacl.ErrInternalError
	}
	g := &Group{
		ID:           id,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Name:         req.Name,
		AccessLevel:  req.AccessLevel,
		CreatedBy:    actorID,
	}
	created, err := s.repo.CreateGroup(ctx, g)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, &actorID, "group", created.ID, "create_group", created.ResourceType, created.ResourceID, map[string]any{
		"group_id":      created.ID,
		"access_level":  created.AccessLevel,
		"resource_type": created.ResourceType,
		"resource_id":   created.ResourceID,
	}, nil)
	return created, nil
}

func (s *Service) GetGroup(ctx context.Context, actorID, groupID string) (*Group, error) {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermRead); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) ListGroups(ctx context.Context, actorID, resourceType, resourceID string) ([]*Group, error) {
	if err := s.requirePerm(ctx, actorID, resourceType, resourceID, pkgacl.PermRead); err != nil {
		return nil, err
	}
	return s.repo.ListGroupsByResource(ctx, resourceType, resourceID)
}

func (s *Service) UpdateGroup(ctx context.Context, actorID, groupID string, req UpdateGroupRequest) (*Group, error) {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermManageMembers); err != nil {
		return nil, err
	}

	name := g.Name
	accessLevel := g.AccessLevel
	if req.Name != nil {
		if *req.Name == "" {
			return nil, pkgacl.ErrValidation
		}
		name = *req.Name
	}
	if req.AccessLevel != nil {
		if !ValidAccessLevel(*req.AccessLevel) {
			return nil, pkgacl.ErrAccessLevelInvalid
		}
		accessLevel = *req.AccessLevel
	}
	updated, err := s.repo.UpdateGroup(ctx, groupID, name, accessLevel)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, &actorID, "group", updated.ID, "update_group", updated.ResourceType, updated.ResourceID, map[string]any{
		"group_id":         updated.ID,
		"old_access_level": g.AccessLevel,
		"new_access_level": updated.AccessLevel,
		"old_name":         g.Name,
		"new_name":         updated.Name,
		"resource_type":    updated.ResourceType,
		"resource_id":      updated.ResourceID,
	}, nil)
	return updated, nil
}

func (s *Service) DeleteGroup(ctx context.Context, actorID, groupID string) error {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermManageMembers); err != nil {
		return err
	}
	if err := s.repo.SoftDeleteGroup(ctx, groupID); err != nil {
		return err
	}
	s.logActivity(ctx, &actorID, "group", groupID, "delete_group", g.ResourceType, g.ResourceID, map[string]any{
		"group_id": groupID,
	}, nil)
	return nil
}

// --- Members ---

func (s *Service) ListMembers(ctx context.Context, actorID, groupID string) ([]*GroupMember, error) {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermRead); err != nil {
		return nil, err
	}
	return s.repo.ListMembersByGroup(ctx, groupID)
}

func (s *Service) AddMember(ctx context.Context, actorID, groupID, userID string) (*GroupMember, error) {
	if userID == "" {
		return nil, pkgacl.ErrValidation
	}
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermManageMembers); err != nil {
		return nil, err
	}
	id, err := newUUID()
	if err != nil {
		return nil, pkgacl.ErrInternalError
	}
	m := &GroupMember{
		ID:         id,
		GroupID:    groupID,
		UserID:     userID,
		SourceType: "direct",
	}
	created, err := s.repo.InsertMember(ctx, m)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, &actorID, "member", created.ID, "add_member", g.ResourceType, g.ResourceID, map[string]any{
		"group_id":       groupID,
		"target_user_id": created.UserID,
	}, nil)
	return created, nil
}

func (s *Service) RemoveMember(ctx context.Context, actorID, groupID, memberID string) error {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermManageMembers); err != nil {
		return err
	}
	m, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	if m.GroupID != groupID {
		return pkgacl.ErrNotFound
	}
	if err := s.repo.SoftRemoveMember(ctx, memberID, groupID); err != nil {
		return err
	}
	s.logActivity(ctx, &actorID, "member", memberID, "remove_member", g.ResourceType, g.ResourceID, map[string]any{
		"group_id":       groupID,
		"target_user_id": m.UserID,
	}, nil)
	return nil
}

// --- Invitations ---

func (s *Service) CreateInvitation(ctx context.Context, actorID, groupID string, expiresInDays int) (*Invitation, string, error) {
	if expiresInDays <= 0 {
		expiresInDays = defaultInvitationTTLDays
	}
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, "", err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermManageMembers); err != nil {
		return nil, "", err
	}
	id, err := newUUID()
	if err != nil {
		return nil, "", pkgacl.ErrInternalError
	}
	rawCode, codeHash, err := generateInviteCode()
	if err != nil {
		return nil, "", pkgacl.ErrInternalError
	}
	inv := &Invitation{
		ID:            id,
		GroupID:       groupID,
		CreatedBy:     actorID,
		CodeHash:      codeHash,
		CodeExpiresAt: time.Now().UTC().Add(time.Duration(expiresInDays) * 24 * time.Hour),
	}
	created, err := s.repo.InsertInvitation(ctx, inv)
	if err != nil {
		return nil, "", err
	}
	s.logActivity(ctx, &actorID, "invitation", created.ID, "create_invitation", g.ResourceType, g.ResourceID, map[string]any{
		"group_id":      g.ID,
		"invitation_id": created.ID,
		"access_level":  g.AccessLevel,
	}, nil)
	return created, rawCode, nil
}

func (s *Service) ListInvitations(ctx context.Context, actorID, resourceType, resourceID string) ([]*Invitation, error) {
	if err := s.requirePerm(ctx, actorID, resourceType, resourceID, pkgacl.PermManageMembers); err != nil {
		return nil, err
	}
	return s.repo.ListInvitationsByResource(ctx, resourceType, resourceID)
}

func (s *Service) CheckInvitation(ctx context.Context, rawCode string) (*InvitationCheckResponse, error) {
	if rawCode == "" {
		return &InvitationCheckResponse{Valid: false, Status: "invalid"}, nil
	}
	hash := hashCode(rawCode)
	inv, err := s.repo.GetInvitationByCodeHash(ctx, hash)
	if errors.Is(err, pkgacl.ErrNotFound) {
		return &InvitationCheckResponse{Valid: false, Status: "invalid"}, nil
	}
	if err != nil {
		return nil, err
	}
	if inv.Status != "pending" {
		return &InvitationCheckResponse{Valid: false, Status: inv.Status}, nil
	}
	if time.Now().After(inv.CodeExpiresAt) {
		return &InvitationCheckResponse{Valid: false, Status: "expired"}, nil
	}
	g, err := s.repo.GetGroupByID(ctx, inv.GroupID)
	if err != nil {
		return &InvitationCheckResponse{Valid: false, Status: "invalid"}, nil
	}
	s.logActivity(ctx, nil, "invitation", inv.ID, "check_invitation", g.ResourceType, g.ResourceID, map[string]any{
		"group_id":      g.ID,
		"invitation_id": inv.ID,
	}, nil)
	return &InvitationCheckResponse{
		Valid:        true,
		Status:       "pending",
		RequiresAuth: true,
		ExpiresAt:    &inv.CodeExpiresAt,
		Resource: &ResourceInfo{
			Type:        g.ResourceType,
			ID:          g.ResourceID,
			AccessLevel: g.AccessLevel,
		},
	}, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, actorID, rawCode string) error {
	if rawCode == "" {
		return pkgacl.ErrValidation
	}
	memberID, err := newUUID()
	if err != nil {
		return pkgacl.ErrInternalError
	}
	return s.repo.AcceptInvitationTx(ctx, hashCode(rawCode), actorID, memberID)
}

func (s *Service) RejectInvitation(ctx context.Context, actorID, rawCode string) error {
	if rawCode == "" {
		return pkgacl.ErrValidation
	}
	hash := hashCode(rawCode)
	inv, err := s.repo.GetInvitationByCodeHash(ctx, hash)
	if err != nil {
		return err
	}
	group, err := s.repo.GetGroupByID(ctx, inv.GroupID)
	if err != nil {
		return err
	}
	if err := s.repo.RejectInvitationTx(ctx, hash); err != nil {
		return err
	}
	s.logActivity(ctx, &actorID, "invitation", inv.ID, "reject_invitation", group.ResourceType, group.ResourceID, map[string]any{
		"group_id":      group.ID,
		"invitation_id": inv.ID,
	}, nil)
	return nil
}

func (s *Service) CancelInvitation(ctx context.Context, actorID, invitationID string) error {
	inv, err := s.repo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return err
	}
	g, err := s.repo.GetGroupByID(ctx, inv.GroupID)
	if err != nil {
		return err
	}
	if err := s.requirePerm(ctx, actorID, g.ResourceType, g.ResourceID, pkgacl.PermManageMembers); err != nil {
		return err
	}
	if err := s.repo.CancelInvitationTx(ctx, invitationID); err != nil {
		return err
	}
	s.logActivity(ctx, &actorID, "invitation", invitationID, "cancel_invitation", g.ResourceType, g.ResourceID, map[string]any{
		"group_id":      g.ID,
		"invitation_id": invitationID,
	}, nil)
	return nil
}

// --- Delete requests ---

func (s *Service) CreateDeleteRequest(ctx context.Context, actorID string, req CreateDeleteRequestRequest) (*DeleteRequest, error) {
	if req.ResourceType == "" || req.ResourceID == "" {
		return nil, pkgacl.ErrValidation
	}

	canRequest, ownerID, err := s.canRequestDelete(ctx, actorID, req.ResourceType, req.ResourceID)
	if err != nil {
		return nil, err
	}
	if !canRequest {
		return nil, pkgacl.ErrNotFound
	}

	id, err := newUUID()
	if err != nil {
		return nil, pkgacl.ErrInternalError
	}
	now := time.Now().UTC()
	status := "pending"
	var expiresAt *time.Time
	if ownerID != actorID {
		exp := now.Add(defaultDeleteRequestTTL)
		expiresAt = &exp
	}

	reason := normalizeReason(req.Reason)
	created, err := s.repo.InsertDeleteRequest(ctx, &DeleteRequest{
		ID:           id,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		RequestedBy:  actorID,
		Status:       status,
		Reason:       reason,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}

	s.logActivity(ctx, &actorID, "delete_request", created.ID, "create_delete_request", created.ResourceType, created.ResourceID, map[string]any{
		"delete_request_id": created.ID,
		"target_user_id":    actorID,
		"reason":            reason,
	}, nil)

	policy, err := s.getDeletePolicy(ctx, req.ResourceType)
	if err != nil {
		return nil, err
	}
	if ownerID == actorID || policy == DeletePolicyDirectDelete {
		return s.executeDeleteRequest(ctx, actorID, created)
	}

	return created, nil
}

func (s *Service) ListDeleteRequests(ctx context.Context, actorID, resourceType, resourceID string, requestedBy, status *string, limit int) ([]*DeleteRequest, error) {
	if resourceType == "" || resourceID == "" {
		return nil, pkgacl.ErrValidation
	}
	if status != nil && !ValidDeleteRequestStatus(*status) {
		return nil, pkgacl.ErrValidation
	}
	if err := s.requirePerm(ctx, actorID, resourceType, resourceID, pkgacl.PermManageMembers); err != nil {
		return nil, err
	}
	return s.repo.ListDeleteRequestsByResource(ctx, resourceType, resourceID, requestedBy, status, limit)
}

func (s *Service) ApproveDeleteRequest(ctx context.Context, actorID, requestID string) (*DeleteRequest, error) {
	dr, err := s.repo.GetDeleteRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwner(ctx, actorID, dr.ResourceType, dr.ResourceID); err != nil {
		return nil, err
	}

	approved, err := s.repo.ApproveDeleteRequestTx(ctx, requestID, actorID)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, &actorID, "delete_request", approved.ID, "approve_delete_request", approved.ResourceType, approved.ResourceID, map[string]any{
		"delete_request_id": approved.ID,
	}, nil)

	return s.executeDeleteRequest(ctx, actorID, approved)
}

func (s *Service) RejectDeleteRequest(ctx context.Context, actorID, requestID string) (*DeleteRequest, error) {
	dr, err := s.repo.GetDeleteRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwner(ctx, actorID, dr.ResourceType, dr.ResourceID); err != nil {
		return nil, err
	}
	rejected, err := s.repo.RejectDeleteRequestTx(ctx, requestID, actorID)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, &actorID, "delete_request", rejected.ID, "reject_delete_request", rejected.ResourceType, rejected.ResourceID, map[string]any{
		"delete_request_id": rejected.ID,
	}, nil)
	return rejected, nil
}

// --- Activity logs ---

func (s *Service) ListActivityLogs(ctx context.Context, actorID, resourceType, resourceID string, actorUserID, action *string, from, to *time.Time, limit int) ([]*ActivityLog, error) {
	if resourceType == "" || resourceID == "" {
		return nil, pkgacl.ErrValidation
	}
	if err := s.requirePerm(ctx, actorID, resourceType, resourceID, pkgacl.PermManageMembers); err != nil {
		return nil, err
	}
	return s.repo.ListActivityLogsByResource(ctx, resourceType, resourceID, actorUserID, action, from, to, limit)
}

// --- internal helpers ---

// requirePerm returns ErrNotFound if the actor does not have the required permission.
// Returns ErrInternalError for unexpected failures.
func (s *Service) requirePerm(ctx context.Context, actorID, resourceType, resourceID, permission string) error {
	can, err := s.CanAccess(ctx, actorID, resourceType, resourceID, permission)
	if err != nil {
		return pkgacl.ErrInternalError
	}
	if !can {
		return pkgacl.ErrNotFound
	}
	return nil
}

func (s *Service) requireOwner(ctx context.Context, actorID, resourceType, resourceID string) error {
	if s.ownerResolver == nil {
		return pkgacl.ErrInternalError
	}
	ownerID, err := s.ownerResolver.GetOwner(ctx, resourceType, resourceID)
	if err != nil {
		return pkgacl.ErrNotFound
	}
	if ownerID != actorID {
		return pkgacl.ErrNotFound
	}
	return nil
}

func (s *Service) canRequestDelete(ctx context.Context, actorID, resourceType, resourceID string) (bool, string, error) {
	if s.ownerResolver == nil {
		return false, "", pkgacl.ErrInternalError
	}
	ownerID, err := s.ownerResolver.GetOwner(ctx, resourceType, resourceID)
	if err != nil {
		return false, "", nil
	}
	if actorID == ownerID {
		return true, ownerID, nil
	}
	accessLevel, err := s.repo.GetEffectiveAccessLevel(ctx, actorID, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, pkgacl.ErrNotFound) {
			return false, ownerID, nil
		}
		return false, ownerID, err
	}
	return pkgacl.CanAccessLevel(accessLevel, pkgacl.PermRequestDelete), ownerID, nil
}

func (s *Service) getDeletePolicy(ctx context.Context, resourceType string) (string, error) {
	if s.deletePolicyResolver == nil {
		return DeletePolicyDirectDelete, nil
	}
	policy, err := s.deletePolicyResolver.GetDeletePolicy(ctx, resourceType)
	if err != nil {
		return "", pkgacl.ErrInternalError
	}
	if policy != DeletePolicyDirectDelete && policy != DeletePolicyOwnerApprovalNeeded {
		return DeletePolicyDirectDelete, nil
	}
	return policy, nil
}

func (s *Service) executeDeleteRequest(ctx context.Context, actorID string, dr *DeleteRequest) (*DeleteRequest, error) {
	if dr == nil {
		return nil, pkgacl.ErrInternalError
	}
	if s.deleteExecutor != nil {
		if err := s.deleteExecutor.DeleteResource(ctx, actorID, dr.ResourceType, dr.ResourceID, dr.ID); err != nil {
			return nil, err
		}
	}
	executed, err := s.repo.MarkDeleteRequestExecutedTx(ctx, dr.ID)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, &actorID, "delete_request", executed.ID, "execute_delete_request", executed.ResourceType, executed.ResourceID, map[string]any{
		"delete_request_id": executed.ID,
	}, nil)
	return executed, nil
}

func normalizeReason(reason *string) *string {
	if reason == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*reason)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// logActivity inserts an activity log entry; errors are silently discarded.
func (s *Service) logActivity(ctx context.Context, actorID *string, entityType, entityID, action, resourceType, resourceID string, metadata map[string]any, correlationID *string) {
	id, err := newUUID()
	if err != nil {
		return
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		metaBytes = []byte(`{}`)
	}
	log := &ActivityLog{
		ID:            id,
		ActorUserID:   actorID,
		EntityType:    entityType,
		EntityID:      entityID,
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Metadata:      metaBytes,
		CorrelationID: correlationID,
	}
	_ = s.repo.InsertActivityLog(ctx, log)
}
