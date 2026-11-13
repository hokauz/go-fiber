package acl

import (
	"context"
	"errors"
	"testing"
	"time"

	pkgacl "go-fiber/pkg/acl"
)

type fakeServiceRepo struct {
	getGroupByIDFn                       func(ctx context.Context, id string) (*Group, error)
	softDeleteGroupFn                    func(ctx context.Context, groupID string) error
	getEffectiveAccessLevelFn            func(ctx context.Context, userID, resourceType, resourceID string) (string, error)
	insertDeleteRequestFn                func(ctx context.Context, dr *DeleteRequest) (*DeleteRequest, error)
	getDeleteRequestByIDFn               func(ctx context.Context, id string) (*DeleteRequest, error)
	approveDeleteRequestTxFn             func(ctx context.Context, requestID, approvedBy string) (*DeleteRequest, error)
	markDeleteRequestExecutedTxFn        func(ctx context.Context, requestID string) (*DeleteRequest, error)
	insertActivityLogFn                  func(ctx context.Context, log *ActivityLog) error
	listActivityLogsByResourceFn         func(ctx context.Context, resourceType, resourceID string, actorUserID, action *string, from, to *time.Time, limit int) ([]*ActivityLog, error)
	listDeleteRequestsByResourceFn       func(ctx context.Context, resourceType, resourceID string, requestedBy, status *string, limit int) ([]*DeleteRequest, error)
	rejectDeleteRequestTxFn              func(ctx context.Context, requestID, rejectedBy string) (*DeleteRequest, error)
	listAccessibleResourceAccessLevelsFn func(ctx context.Context, userID, resourceType string) ([]ResourceAccessLevel, error)
}

func unexpectedCall(name string) error {
	return errors.New("unexpected call: " + name)
}

func (f *fakeServiceRepo) CreateGroup(ctx context.Context, g *Group) (*Group, error) {
	return nil, unexpectedCall("CreateGroup")
}

func (f *fakeServiceRepo) GetGroupByID(ctx context.Context, id string) (*Group, error) {
	if f.getGroupByIDFn != nil {
		return f.getGroupByIDFn(ctx, id)
	}
	return nil, unexpectedCall("GetGroupByID")
}

func (f *fakeServiceRepo) ListGroupsByResource(ctx context.Context, resourceType, resourceID string) ([]*Group, error) {
	return nil, unexpectedCall("ListGroupsByResource")
}

func (f *fakeServiceRepo) UpdateGroup(ctx context.Context, id, name, accessLevel string) (*Group, error) {
	return nil, unexpectedCall("UpdateGroup")
}

func (f *fakeServiceRepo) SoftDeleteGroup(ctx context.Context, groupID string) error {
	if f.softDeleteGroupFn != nil {
		return f.softDeleteGroupFn(ctx, groupID)
	}
	return unexpectedCall("SoftDeleteGroup")
}

func (f *fakeServiceRepo) InsertMember(ctx context.Context, m *GroupMember) (*GroupMember, error) {
	return nil, unexpectedCall("InsertMember")
}

func (f *fakeServiceRepo) ListMembersByGroup(ctx context.Context, groupID string) ([]*GroupMember, error) {
	return nil, unexpectedCall("ListMembersByGroup")
}

func (f *fakeServiceRepo) GetMemberByID(ctx context.Context, memberID string) (*GroupMember, error) {
	return nil, unexpectedCall("GetMemberByID")
}

func (f *fakeServiceRepo) SoftRemoveMember(ctx context.Context, memberID, groupID string) error {
	return unexpectedCall("SoftRemoveMember")
}

func (f *fakeServiceRepo) InsertInvitation(ctx context.Context, inv *Invitation) (*Invitation, error) {
	return nil, unexpectedCall("InsertInvitation")
}

func (f *fakeServiceRepo) GetInvitationByCodeHash(ctx context.Context, hash string) (*Invitation, error) {
	return nil, unexpectedCall("GetInvitationByCodeHash")
}

func (f *fakeServiceRepo) GetInvitationByID(ctx context.Context, id string) (*Invitation, error) {
	return nil, unexpectedCall("GetInvitationByID")
}

func (f *fakeServiceRepo) ListInvitationsByResource(ctx context.Context, resourceType, resourceID string) ([]*Invitation, error) {
	return nil, unexpectedCall("ListInvitationsByResource")
}

func (f *fakeServiceRepo) AcceptInvitationTx(ctx context.Context, codeHash, actorID, memberID string) error {
	return unexpectedCall("AcceptInvitationTx")
}

func (f *fakeServiceRepo) RejectInvitationTx(ctx context.Context, codeHash string) error {
	return unexpectedCall("RejectInvitationTx")
}

func (f *fakeServiceRepo) CancelInvitationTx(ctx context.Context, invitationID string) error {
	return unexpectedCall("CancelInvitationTx")
}

func (f *fakeServiceRepo) GetEffectiveAccessLevel(ctx context.Context, userID, resourceType, resourceID string) (string, error) {
	if f.getEffectiveAccessLevelFn != nil {
		return f.getEffectiveAccessLevelFn(ctx, userID, resourceType, resourceID)
	}
	return "", unexpectedCall("GetEffectiveAccessLevel")
}

func (f *fakeServiceRepo) ListAccessibleResourceAccessLevels(ctx context.Context, userID, resourceType string) ([]ResourceAccessLevel, error) {
	if f.listAccessibleResourceAccessLevelsFn != nil {
		return f.listAccessibleResourceAccessLevelsFn(ctx, userID, resourceType)
	}
	return nil, unexpectedCall("ListAccessibleResourceAccessLevels")
}

func (f *fakeServiceRepo) InsertDeleteRequest(ctx context.Context, dr *DeleteRequest) (*DeleteRequest, error) {
	if f.insertDeleteRequestFn != nil {
		return f.insertDeleteRequestFn(ctx, dr)
	}
	return nil, unexpectedCall("InsertDeleteRequest")
}

func (f *fakeServiceRepo) GetDeleteRequestByID(ctx context.Context, id string) (*DeleteRequest, error) {
	if f.getDeleteRequestByIDFn != nil {
		return f.getDeleteRequestByIDFn(ctx, id)
	}
	return nil, unexpectedCall("GetDeleteRequestByID")
}

func (f *fakeServiceRepo) ListDeleteRequestsByResource(ctx context.Context, resourceType, resourceID string, requestedBy, status *string, limit int) ([]*DeleteRequest, error) {
	if f.listDeleteRequestsByResourceFn != nil {
		return f.listDeleteRequestsByResourceFn(ctx, resourceType, resourceID, requestedBy, status, limit)
	}
	return nil, unexpectedCall("ListDeleteRequestsByResource")
}

func (f *fakeServiceRepo) ApproveDeleteRequestTx(ctx context.Context, requestID, approvedBy string) (*DeleteRequest, error) {
	if f.approveDeleteRequestTxFn != nil {
		return f.approveDeleteRequestTxFn(ctx, requestID, approvedBy)
	}
	return nil, unexpectedCall("ApproveDeleteRequestTx")
}

func (f *fakeServiceRepo) RejectDeleteRequestTx(ctx context.Context, requestID, rejectedBy string) (*DeleteRequest, error) {
	if f.rejectDeleteRequestTxFn != nil {
		return f.rejectDeleteRequestTxFn(ctx, requestID, rejectedBy)
	}
	return nil, unexpectedCall("RejectDeleteRequestTx")
}

func (f *fakeServiceRepo) MarkDeleteRequestExecutedTx(ctx context.Context, requestID string) (*DeleteRequest, error) {
	if f.markDeleteRequestExecutedTxFn != nil {
		return f.markDeleteRequestExecutedTxFn(ctx, requestID)
	}
	return nil, unexpectedCall("MarkDeleteRequestExecutedTx")
}

func (f *fakeServiceRepo) InsertActivityLog(ctx context.Context, log *ActivityLog) error {
	if f.insertActivityLogFn != nil {
		return f.insertActivityLogFn(ctx, log)
	}
	return unexpectedCall("InsertActivityLog")
}

func (f *fakeServiceRepo) ListActivityLogsByResource(ctx context.Context, resourceType, resourceID string, actorUserID, action *string, from, to *time.Time, limit int) ([]*ActivityLog, error) {
	if f.listActivityLogsByResourceFn != nil {
		return f.listActivityLogsByResourceFn(ctx, resourceType, resourceID, actorUserID, action, from, to, limit)
	}
	return nil, unexpectedCall("ListActivityLogsByResource")
}

type fakeOwnerResolver struct {
	getOwnerFn func(ctx context.Context, resourceType, resourceID string) (string, error)
}

func (f *fakeOwnerResolver) GetOwner(ctx context.Context, resourceType, resourceID string) (string, error) {
	if f.getOwnerFn != nil {
		return f.getOwnerFn(ctx, resourceType, resourceID)
	}
	return "", errors.New("owner resolver not configured")
}

type fakePolicyResolver struct {
	getPolicyFn func(ctx context.Context, resourceType string) (string, error)
}

func (f *fakePolicyResolver) GetDeletePolicy(ctx context.Context, resourceType string) (string, error) {
	if f.getPolicyFn != nil {
		return f.getPolicyFn(ctx, resourceType)
	}
	return DeletePolicyDirectDelete, nil
}

type fakeDeleteExecutor struct {
	called bool
	actor  string
	rType  string
	rID    string
	reqID  string
	err    error
}

func (f *fakeDeleteExecutor) DeleteResource(_ context.Context, actorUserID, resourceType, resourceID, deleteRequestID string) error {
	f.called = true
	f.actor = actorUserID
	f.rType = resourceType
	f.rID = resourceID
	f.reqID = deleteRequestID
	return f.err
}

func TestCreateDeleteRequest_PendingApprovalForResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logs := []string{}
	repo := &fakeServiceRepo{
		getEffectiveAccessLevelFn: func(context.Context, string, string, string) (string, error) {
			return pkgacl.AccessLevelEdit, nil
		},
		insertDeleteRequestFn: func(_ context.Context, dr *DeleteRequest) (*DeleteRequest, error) {
			now := time.Now().UTC()
			dr.CreatedAt = now
			dr.UpdatedAt = now
			return dr, nil
		},
		insertActivityLogFn: func(_ context.Context, log *ActivityLog) error {
			logs = append(logs, log.Action)
			return nil
		},
	}

	svc := NewService(
		repo,
		&fakeOwnerResolver{getOwnerFn: func(context.Context, string, string) (string, error) {
			return "owner-1", nil
		}},
		&fakePolicyResolver{getPolicyFn: func(context.Context, string) (string, error) {
			return DeletePolicyOwnerApprovalNeeded, nil
		}},
		&fakeDeleteExecutor{},
	)

	res, err := svc.CreateDeleteRequest(ctx, "member-1", CreateDeleteRequestRequest{
		ResourceType: "grant",
		ResourceID:   "res-1",
	})
	if err != nil {
		t.Fatalf("CreateDeleteRequest() error = %v", err)
	}
	if res.Status != "pending" {
		t.Fatalf("CreateDeleteRequest() status = %q, want pending", res.Status)
	}
	if len(logs) != 1 || logs[0] != "create_delete_request" {
		t.Fatalf("CreateDeleteRequest() logs = %v, want [create_delete_request]", logs)
	}
}

func TestApproveDeleteRequest_ExecutesCallbackAndLogs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logs := []string{}
	executor := &fakeDeleteExecutor{}

	repo := &fakeServiceRepo{
		getDeleteRequestByIDFn: func(_ context.Context, id string) (*DeleteRequest, error) {
			return &DeleteRequest{ID: id, ResourceType: "grant", ResourceID: "res-2", Status: "pending"}, nil
		},
		approveDeleteRequestTxFn: func(_ context.Context, requestID, approvedBy string) (*DeleteRequest, error) {
			return &DeleteRequest{ID: requestID, ResourceType: "grant", ResourceID: "res-2", Status: "approved", RequestedBy: "member-2"}, nil
		},
		markDeleteRequestExecutedTxFn: func(_ context.Context, requestID string) (*DeleteRequest, error) {
			return &DeleteRequest{ID: requestID, ResourceType: "grant", ResourceID: "res-2", Status: "executed", RequestedBy: "member-2"}, nil
		},
		insertActivityLogFn: func(_ context.Context, log *ActivityLog) error {
			logs = append(logs, log.Action)
			return nil
		},
	}

	svc := NewService(
		repo,
		&fakeOwnerResolver{getOwnerFn: func(context.Context, string, string) (string, error) {
			return "owner-2", nil
		}},
		nil,
		executor,
	)

	res, err := svc.ApproveDeleteRequest(ctx, "owner-2", "req-2")
	if err != nil {
		t.Fatalf("ApproveDeleteRequest() error = %v", err)
	}
	if res.Status != "executed" {
		t.Fatalf("ApproveDeleteRequest() status = %q, want executed", res.Status)
	}
	if !executor.called {
		t.Fatal("ApproveDeleteRequest() expected delete executor to be called")
	}
	if executor.rType != "grant" || executor.rID != "res-2" || executor.reqID != "req-2" {
		t.Fatalf("ApproveDeleteRequest() executor args mismatch: %+v", executor)
	}
	if len(logs) != 2 || logs[0] != "approve_delete_request" || logs[1] != "execute_delete_request" {
		t.Fatalf("ApproveDeleteRequest() logs = %v, want [approve_delete_request execute_delete_request]", logs)
	}
}

func TestDeleteGroup_OwnerCanDeleteWithoutPendingInviteCheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logs := []string{}
	deleted := false

	repo := &fakeServiceRepo{
		getGroupByIDFn: func(_ context.Context, id string) (*Group, error) {
			return &Group{
				ID:           id,
				ResourceType: "grant",
				ResourceID:   "res-10",
				AccessLevel:  pkgacl.AccessLevelManageAccess,
			}, nil
		},
		softDeleteGroupFn: func(_ context.Context, groupID string) error {
			if groupID != "group-10" {
				return errors.New("unexpected group id")
			}
			deleted = true
			return nil
		},
		insertActivityLogFn: func(_ context.Context, log *ActivityLog) error {
			logs = append(logs, log.Action)
			return nil
		},
	}

	svc := NewService(
		repo,
		&fakeOwnerResolver{getOwnerFn: func(context.Context, string, string) (string, error) {
			return "owner-10", nil
		}},
		nil,
		nil,
	)

	if err := svc.DeleteGroup(ctx, "owner-10", "group-10"); err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}
	if !deleted {
		t.Fatal("DeleteGroup() expected SoftDeleteGroup to be called")
	}
	if len(logs) != 1 || logs[0] != "delete_group" {
		t.Fatalf("DeleteGroup() logs = %v, want [delete_group]", logs)
	}
}

func TestListActivityLogs_ByOwner_ReturnsEntries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()
	actorFilter := "member-3"
	actionFilter := "add_member"
	from := now.Add(-time.Hour)
	to := now.Add(time.Hour)

	repo := &fakeServiceRepo{
		listActivityLogsByResourceFn: func(_ context.Context, resourceType, resourceID string, actorUserID, action *string, f, tt *time.Time, limit int) ([]*ActivityLog, error) {
			if resourceType != "grant" || resourceID != "res-3" {
				return nil, errors.New("unexpected resource filter")
			}
			if actorUserID == nil || *actorUserID != actorFilter {
				return nil, errors.New("missing actor filter")
			}
			if action == nil || *action != actionFilter {
				return nil, errors.New("missing action filter")
			}
			if f == nil || tt == nil || f.IsZero() || tt.IsZero() {
				return nil, errors.New("missing date filters")
			}
			if limit != 50 {
				return nil, errors.New("unexpected limit")
			}
			return []*ActivityLog{{ID: "log-1", Action: "add_member", ResourceType: "grant", ResourceID: "res-3"}}, nil
		},
	}

	svc := NewService(
		repo,
		&fakeOwnerResolver{getOwnerFn: func(context.Context, string, string) (string, error) {
			return "owner-3", nil
		}},
		nil,
		nil,
	)

	entries, err := svc.ListActivityLogs(ctx, "owner-3", "grant", "res-3", &actorFilter, &actionFilter, &from, &to, 50)
	if err != nil {
		t.Fatalf("ListActivityLogs() error = %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "log-1" {
		t.Fatalf("ListActivityLogs() entries = %#v, want one log-1", entries)
	}
}

func TestListActivityLogs_NonManagerDenied(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &fakeServiceRepo{
		getEffectiveAccessLevelFn: func(context.Context, string, string, string) (string, error) {
			return pkgacl.AccessLevelView, nil
		},
		listActivityLogsByResourceFn: func(context.Context, string, string, *string, *string, *time.Time, *time.Time, int) ([]*ActivityLog, error) {
			return nil, errors.New("should not be called")
		},
	}

	svc := NewService(
		repo,
		&fakeOwnerResolver{getOwnerFn: func(context.Context, string, string) (string, error) {
			return "owner-4", nil
		}},
		nil,
		nil,
	)

	_, err := svc.ListActivityLogs(ctx, "viewer-4", "grant", "res-4", nil, nil, nil, nil, 10)
	if !errors.Is(err, pkgacl.ErrNotFound) {
		t.Fatalf("ListActivityLogs() error = %v, want %v", err, pkgacl.ErrNotFound)
	}
}

var _ serviceRepository = (*fakeServiceRepo)(nil)
var _ pkgacl.OwnerResolver = (*fakeOwnerResolver)(nil)
var _ DeletePolicyResolver = (*fakePolicyResolver)(nil)
var _ ResourceDeleteExecutor = (*fakeDeleteExecutor)(nil)
