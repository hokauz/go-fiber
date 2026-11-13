package acl

const (
	AccessLevelOwner        = "owner"
	AccessLevelManageAccess = "manage_access"
	AccessLevelEdit         = "edit"
	AccessLevelView         = "view"
)

const (
	PermRead          = "read"
	PermUpdate        = "update"
	PermDelete        = "delete"
	PermManageMembers = "manage_members"
	PermRequestDelete = "request_delete"
)

// CanAccessLevel reports whether an access level grants the given permission.
func CanAccessLevel(accessLevel, permission string) bool {
	switch accessLevel {
	case AccessLevelOwner, AccessLevelManageAccess:
		return permission == PermRead || permission == PermUpdate ||
			permission == PermDelete || permission == PermManageMembers ||
			permission == PermRequestDelete
	case AccessLevelEdit:
		return permission == PermRead || permission == PermUpdate ||
			permission == PermDelete || permission == PermRequestDelete
	case AccessLevelView:
		return permission == PermRead
	default:
		return false
	}
}

// CanRole is kept as an alias for backward compatibility in internal callers.
func CanRole(role, permission string) bool {
	return CanAccessLevel(role, permission)
}
