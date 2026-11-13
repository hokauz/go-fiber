package acl

// ACLError represents an ACL domain error with a machine-readable code.
type ACLError struct {
	Code    string
	Message string // safe message for HTTP response
	Cause   error  // internal cause (logged, never sent to client)
}

func (e *ACLError) Error() string { return e.Code }
func (e *ACLError) Unwrap() error { return e.Cause }

// Sentinel errors.
var (
	ErrNotFound                  = &ACLError{Code: "not-found", Message: "Resource not found"}
	ErrValidation                = &ACLError{Code: "validation-error", Message: "Validation error"}
	ErrAccessLevelInvalid        = &ACLError{Code: "acl/access-level-invalid", Message: "Invalid access level"}
	ErrResourceTypeInvalid       = &ACLError{Code: "acl/resource-type-invalid", Message: "Invalid resource type"}
	ErrInviteExpired             = &ACLError{Code: "acl/invite-expired", Message: "Invite has expired"}
	ErrInviteInvalidState        = &ACLError{Code: "acl/invite-invalid-state", Message: "Invite is not in a valid state"}
	ErrInviteAlreadyConsumed     = &ACLError{Code: "acl/invite-already-consumed", Message: "Invite was already consumed"}
	ErrMemberAlreadyExists       = &ACLError{Code: "acl/member-already-exists", Message: "User is already a member"}
	ErrGroupHasPendingInvites    = &ACLError{Code: "acl/group-has-pending-invites", Message: "Group has pending invites"}
	ErrGroupNameDuplicate        = &ACLError{Code: "acl/group-name-duplicate", Message: "Group name already exists for this resource"}
	ErrDeleteRequestAlreadyOpen  = &ACLError{Code: "acl/delete-request-already-open", Message: "Delete request already exists for this resource"}
	ErrDeleteRequestInvalidState = &ACLError{Code: "acl/delete-request-invalid-state", Message: "Delete request is not in a valid state"}
	ErrInternalError             = &ACLError{Code: "internal-error", Message: "Internal server error"}
)
