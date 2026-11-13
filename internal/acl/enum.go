package acl

import pkgacl "go-fiber/pkg/acl"

const (
	DeletePolicyDirectDelete        = "direct_delete"
	DeletePolicyOwnerApprovalNeeded = "owner_approval_required"
)

// ValidAccessLevel reports whether the access level is assignable to groups.
// owner is not assignable; it is resolved externally via OwnerResolver.
func ValidAccessLevel(accessLevel string) bool {
	return accessLevel == pkgacl.AccessLevelManageAccess ||
		accessLevel == pkgacl.AccessLevelEdit ||
		accessLevel == pkgacl.AccessLevelView
}

// ValidInviteStatus reports whether the status string is a known invite status.
func ValidInviteStatus(status string) bool {
	switch status {
	case "pending", "accepted", "rejected", "cancelled", "expired":
		return true
	}
	return false
}

// ValidDeleteRequestStatus reports whether the status string is a known delete-request status.
func ValidDeleteRequestStatus(status string) bool {
	switch status {
	case "pending", "approved", "rejected", "expired", "executed":
		return true
	}
	return false
}
