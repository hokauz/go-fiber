package acl

import (
	pkgacl "go-fiber/pkg/acl"
	"go-fiber/pkg/db"

	"github.com/gofiber/fiber/v3"
)

func registerUnavailableACLRoutes(r fiber.Router) {
	unavailable := func(c fiber.Ctx) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal-error",
			"message": "Internal server error",
		})
	}

	// Groups
	r.Post("/v1/acl/groups", unavailable)
	r.Get("/v1/acl/groups", unavailable)
	r.Get("/v1/acl/groups/:id", unavailable)
	r.Patch("/v1/acl/groups/:id", unavailable)
	r.Delete("/v1/acl/groups/:id", unavailable)

	// Members
	r.Get("/v1/acl/groups/:id/members", unavailable)
	r.Post("/v1/acl/groups/:id/members", unavailable)
	r.Delete("/v1/acl/groups/:id/members/:member_id", unavailable)

	// Invitations
	r.Post("/v1/acl/invitations", unavailable)
	r.Get("/v1/acl/invitations", unavailable)
	r.Post("/v1/acl/invitations/accept", unavailable)
	r.Post("/v1/acl/invitations/reject", unavailable)
	r.Post("/v1/acl/invitations/:id/cancel", unavailable)

	// Access helper
	r.Get("/v1/acl/access/level", unavailable)

	// Delete requests
	r.Post("/v1/acl/delete-requests", unavailable)
	r.Get("/v1/acl/delete-requests", unavailable)
	r.Post("/v1/acl/delete-requests/:id/approve", unavailable)
	r.Post("/v1/acl/delete-requests/:id/reject", unavailable)

	// Activity logs
	r.Get("/v1/acl/activity-logs", unavailable)
}

// PublicRegister registers the public ACL endpoint (invitation check — no auth required).
func PublicRegister(database db.DB, observer QueryObserver) func(fiber.Router) {
	return func(r fiber.Router) {
		if database == nil {
			r.Get("/v1/acl/invitations/check", func(c fiber.Ctx) error {
				return c.JSON(InvitationCheckResponse{Valid: false, Status: "invalid"})
			})
			return
		}

		repo := NewRepository(database, observer)
		svc := NewService(repo, nil, nil, nil) // no ownerResolver needed for public check
		h := NewHandler(svc)
		r.Get("/v1/acl/invitations/check", h.CheckInvitationDoc)
	}
}

// PrivateRegister registers all authenticated ACL endpoints.
// ownerResolver is provided by the resource module (e.g. accounts.NewOwnerResolver(db)).
func PrivateRegister(
	database db.DB,
	ownerResolver pkgacl.OwnerResolver,
	deletePolicyResolver DeletePolicyResolver,
	deleteExecutor ResourceDeleteExecutor,
	observer QueryObserver,
) func(fiber.Router) {
	return func(r fiber.Router) {
		if database == nil {
			registerUnavailableACLRoutes(r)
			return
		}

		repo := NewRepository(database, observer)
		svc := NewService(repo, ownerResolver, deletePolicyResolver, deleteExecutor)
		h := NewHandler(svc)

		// Groups
		r.Post("/v1/acl/groups", h.CreateGroupDoc)
		r.Get("/v1/acl/groups", h.ListGroupsDoc)
		r.Get("/v1/acl/groups/:id", h.GetGroupDoc)
		r.Patch("/v1/acl/groups/:id", h.UpdateGroupDoc)
		r.Delete("/v1/acl/groups/:id", h.DeleteGroupDoc)

		// Members
		r.Get("/v1/acl/groups/:id/members", h.ListMembersDoc)
		r.Post("/v1/acl/groups/:id/members", h.AddMemberDoc)
		r.Delete("/v1/acl/groups/:id/members/:member_id", h.RemoveMemberDoc)

		// Invitations
		r.Post("/v1/acl/invitations", h.CreateInvitationDoc)
		r.Get("/v1/acl/invitations", h.ListInvitationsDoc)
		r.Post("/v1/acl/invitations/accept", h.AcceptInvitationDoc)
		r.Post("/v1/acl/invitations/reject", h.RejectInvitationDoc)
		r.Post("/v1/acl/invitations/:id/cancel", h.CancelInvitationDoc)

		// Access helper
		r.Get("/v1/acl/access/level", h.GetEffectiveAccessLevelDoc)

		// Delete requests
		r.Post("/v1/acl/delete-requests", h.CreateDeleteRequestDoc)
		r.Get("/v1/acl/delete-requests", h.ListDeleteRequestsDoc)
		r.Post("/v1/acl/delete-requests/:id/approve", h.ApproveDeleteRequestDoc)
		r.Post("/v1/acl/delete-requests/:id/reject", h.RejectDeleteRequestDoc)

		// Activity logs
		r.Get("/v1/acl/activity-logs", h.ListActivityLogsDoc)
	}
}
