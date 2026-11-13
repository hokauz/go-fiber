package acl

import (
	"errors"

	pkgacl "go-fiber/pkg/acl"
)

// MapToHTTP converts an ACL domain error into an HTTP status code, error code, and message.
func MapToHTTP(err error) (status int, code, message string) {
	var ae *pkgacl.ACLError
	if !errors.As(err, &ae) {
		return 500, "internal-error", "Internal server error"
	}
	switch ae.Code {
	case "not-found":
		return 404, ae.Code, ae.Message
	case "validation-error", "acl/access-level-invalid", "acl/resource-type-invalid",
		"acl/invite-expired", "acl/invite-invalid-state",
		"acl/group-has-pending-invites", "acl/group-name-duplicate",
		"acl/delete-request-invalid-state":
		return 422, ae.Code, ae.Message
	case "acl/invite-already-consumed", "acl/member-already-exists",
		"acl/delete-request-already-open":
		return 409, ae.Code, ae.Message
	default:
		return 500, ae.Code, ae.Message
	}
}
