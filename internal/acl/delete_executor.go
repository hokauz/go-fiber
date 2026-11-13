package acl

import "context"

// ResourceDeleteExecutor executes resource deletion side-effects after ACL approval.
type ResourceDeleteExecutor interface {
	DeleteResource(ctx context.Context, actorUserID, resourceType, resourceID, deleteRequestID string) error
}
