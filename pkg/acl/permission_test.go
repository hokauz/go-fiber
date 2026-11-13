package acl

import "testing"

func TestCanAccessLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		level      string
		permission string
		want       bool
	}{
		{name: "owner read", level: AccessLevelOwner, permission: PermRead, want: true},
		{name: "owner manage members", level: AccessLevelOwner, permission: PermManageMembers, want: true},
		{name: "owner request delete", level: AccessLevelOwner, permission: PermRequestDelete, want: true},
		{name: "manage_access manage members", level: AccessLevelManageAccess, permission: PermManageMembers, want: true},
		{name: "manage_access delete", level: AccessLevelManageAccess, permission: PermDelete, want: true},
		{name: "edit update", level: AccessLevelEdit, permission: PermUpdate, want: true},
		{name: "edit delete", level: AccessLevelEdit, permission: PermDelete, want: true},
		{name: "edit manage members denied", level: AccessLevelEdit, permission: PermManageMembers, want: false},
		{name: "view read", level: AccessLevelView, permission: PermRead, want: true},
		{name: "view update denied", level: AccessLevelView, permission: PermUpdate, want: false},
		{name: "unknown level denied", level: "legacy", permission: PermRead, want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := CanAccessLevel(tc.level, tc.permission)
			if got != tc.want {
				t.Fatalf("CanAccessLevel(%q,%q)=%v want=%v", tc.level, tc.permission, got, tc.want)
			}
		})
	}
}
