package acl

import "context"

// DeletePolicyResolver resolves delete policy by resource type.
type DeletePolicyResolver interface {
	GetDeletePolicy(ctx context.Context, resourceType string) (string, error)
}

// StaticDeletePolicyResolver is an in-memory resolver keyed by resource type.
type StaticDeletePolicyResolver struct {
	defaultPolicy string
	byResource    map[string]string
}

// NewStaticDeletePolicyResolver builds a static delete policy resolver.
func NewStaticDeletePolicyResolver(defaultPolicy string, byResource map[string]string) *StaticDeletePolicyResolver {
	if defaultPolicy == "" {
		defaultPolicy = DeletePolicyDirectDelete
	}
	if byResource == nil {
		byResource = map[string]string{}
	}
	return &StaticDeletePolicyResolver{
		defaultPolicy: defaultPolicy,
		byResource:    byResource,
	}
}

// GetDeletePolicy returns the configured policy for a resource type.
func (r *StaticDeletePolicyResolver) GetDeletePolicy(_ context.Context, resourceType string) (string, error) {
	if p, ok := r.byResource[resourceType]; ok && p != "" {
		return p, nil
	}
	return r.defaultPolicy, nil
}
