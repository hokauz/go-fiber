package acl

import "time"

// --- DB models ---

// Group is an ACL group bound to a specific resource.
type Group struct {
	ID           string     `db:"id"`
	ResourceType string     `db:"resource_type"`
	ResourceID   string     `db:"resource_id"`
	Name         string     `db:"name"`
	AccessLevel  string     `db:"access_level"`
	CreatedBy    string     `db:"created_by"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// GroupMember is an active membership record. Being a member equals having permission.
type GroupMember struct {
	ID         string     `db:"id"`
	GroupID    string     `db:"group_id"`
	UserID     string     `db:"user_id"`
	SourceType string     `db:"source_type"` // invite | direct
	SourceID   *string    `db:"source_id"`   // invitation.id or nil if direct
	CreatedAt  time.Time  `db:"created_at"`
	RemovedAt  *time.Time `db:"removed_at"`
}

// Invitation is a shareable link that grants membership to a group.
type Invitation struct {
	ID            string     `db:"id"`
	GroupID       string     `db:"group_id"`
	CreatedBy     string     `db:"created_by"`
	CodeHash      string     `db:"code_hash"`
	CodeExpiresAt time.Time  `db:"code_expires_at"`
	Status        string     `db:"status"` // pending | accepted | rejected | cancelled | expired
	ConsumedAt    *time.Time `db:"consumed_at"`
	AcceptedBy    *string    `db:"accepted_by"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// ActivityLog is an append-only audit record.
type ActivityLog struct {
	ID            string    `db:"id"`
	ActorUserID   *string   `db:"actor_user_id"`
	EntityType    string    `db:"entity_type"`
	EntityID      string    `db:"entity_id"`
	Action        string    `db:"action"`
	ResourceType  string    `db:"resource_type"`
	ResourceID    string    `db:"resource_id"`
	Metadata      []byte    `db:"metadata"`
	CorrelationID *string   `db:"correlation_id"`
	CreatedAt     time.Time `db:"created_at"`
}

// ResourceAccessLevel pairs a resource ID with the user's effective access level on it.
type ResourceAccessLevel struct {
	ResourceID           string
	EffectiveAccessLevel string
}

// DeleteRequest represents a pending or finalized delete workflow for a resource.
type DeleteRequest struct {
	ID           string     `db:"id"`
	ResourceType string     `db:"resource_type"`
	ResourceID   string     `db:"resource_id"`
	RequestedBy  string     `db:"requested_by"`
	Status       string     `db:"status"` // pending | approved | rejected | expired | executed
	Reason       *string    `db:"reason"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	ApprovedBy   *string    `db:"approved_by"`
	ApprovedAt   *time.Time `db:"approved_at"`
	RejectedBy   *string    `db:"rejected_by"`
	RejectedAt   *time.Time `db:"rejected_at"`
	ExecutedAt   *time.Time `db:"executed_at"`
	ExpiresAt    *time.Time `db:"expires_at"`
}

// --- HTTP request / response models ---

type CreateGroupRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Name         string `json:"name"`
	AccessLevel  string `json:"access_level"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"`
	AccessLevel *string `json:"access_level"`
}

type GroupResponse struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Name         string    `json:"name"`
	AccessLevel  string    `json:"access_level"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
}

type MemberResponse struct {
	ID         string    `json:"id"`
	GroupID    string    `json:"group_id"`
	UserID     string    `json:"user_id"`
	SourceType string    `json:"source_type"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateInvitationRequest struct {
	GroupID       string `json:"group_id"`
	ExpiresInDays int    `json:"expires_in_days"`
}

type InvitationResponse struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	Status        string    `json:"status"`
	CodeExpiresAt time.Time `json:"code_expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateInvitationResponse embeds InvitationResponse and adds the raw code (returned once only).
type CreateInvitationResponse struct {
	InvitationResponse
	Code string `json:"code"`
}

type InvitationCheckResponse struct {
	Valid        bool          `json:"valid"`
	Status       string        `json:"status"`
	RequiresAuth bool          `json:"requires_auth,omitempty"`
	ExpiresAt    *time.Time    `json:"expires_at,omitempty"`
	Resource     *ResourceInfo `json:"resource,omitempty"`
}

type ResourceInfo struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	AccessLevel string `json:"access_level"`
}

type InvitationCodeRequest struct {
	Code string `json:"code"`
}

type CreateDeleteRequestRequest struct {
	ResourceType string  `json:"resource_type"`
	ResourceID   string  `json:"resource_id"`
	Reason       *string `json:"reason,omitempty"`
}

type DeleteRequestResponse struct {
	ID           string     `json:"id"`
	ResourceType string     `json:"resource_type"`
	ResourceID   string     `json:"resource_id"`
	RequestedBy  string     `json:"requested_by"`
	Status       string     `json:"status"`
	Reason       *string    `json:"reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ApprovedBy   *string    `json:"approved_by,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	RejectedBy   *string    `json:"rejected_by,omitempty"`
	RejectedAt   *time.Time `json:"rejected_at,omitempty"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

type ActivityLogResponse struct {
	ID            string    `json:"id"`
	ActorUserID   *string   `json:"actor_user_id,omitempty"`
	EntityType    string    `json:"entity_type"`
	EntityID      string    `json:"entity_id"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	Metadata      any       `json:"metadata"`
	CorrelationID *string   `json:"correlation_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
