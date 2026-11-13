//go:build e2e

// Package acl contains e2e tests for the ACL (access control) endpoints.
//
// Public tests (no env vars needed):
//   - invitation check with unknown/empty code
//   - auth rejection on private endpoints
//
// Journey tests require the following env vars:
//
//	E2E_OWNER_TOKEN   — JWT for the user who owns the resource
//	E2E_MEMBER_TOKEN  — JWT for the user being invited
//	E2E_RESOURCE_ID   — UUID of a resource owned by E2E_OWNER_TOKEN user
//	E2E_RESOURCE_TYPE — resource type (default: "account")
package acl
