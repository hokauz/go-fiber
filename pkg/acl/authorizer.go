package acl

import "context"

// OwnerResolver returns the owner user ID for a resource.
type OwnerResolver interface {
	GetOwner(ctx context.Context, resourceType, resourceID string) (string, error)
}

// Authorizer exposes permission decisions for resource access.
type Authorizer interface {
	CanAccess(ctx context.Context, userID, resourceType, resourceID, permission string) (bool, error)
	GetEffectiveAccessLevel(ctx context.Context, userID, resourceType, resourceID string) (string, error)
	ListAccessibleResourceIDs(ctx context.Context, userID, resourceType, permission string) ([]string, error)
}
