package acl

import (
	"context"
	"database/sql"

	internalacl "go-fiber/internal/acl"
	"go-fiber/pkg/db"

	"github.com/gofiber/fiber/v3"
)

type ownerResolver struct {
	database db.DB
}

func (r *ownerResolver) GetOwner(ctx context.Context, resourceType, resourceID string) (string, error) {
	if r == nil || r.database == nil || r.database.Raw() == nil {
		return "", sql.ErrConnDone
	}
	var ownerID string
	err := r.database.Raw().QueryRowContext(
		ctx,
		`SELECT owner_user_id FROM acl_resource_owners WHERE resource_type = $1 AND resource_id = $2`,
		resourceType,
		resourceID,
	).Scan(&ownerID)
	if err != nil {
		return "", err
	}
	return ownerID, nil
}

type deleteExecutor struct{}

func (e *deleteExecutor) DeleteResource(_ context.Context, _ string, _ string, _ string, _ string) error {
	// Concrete callback hook. Resource modules can replace this executor with real deletions.
	return nil
}

// PublicRegister proxies internal ACL public routes.
func PublicRegister(database db.DB, observer internalacl.QueryObserver) func(fiber.Router) {
	return internalacl.PublicRegister(database, observer)
}

// PrivateRegister wires ACL private routes with concrete owner and delete resolvers.
func PrivateRegister(database db.DB, observer internalacl.QueryObserver) func(fiber.Router) {
	policyResolver := internalacl.NewStaticDeletePolicyResolver(
		internalacl.DeletePolicyDirectDelete,
		map[string]string{
			"grant": internalacl.DeletePolicyOwnerApprovalNeeded,
		},
	)
	return internalacl.PrivateRegister(
		database,
		&ownerResolver{database: database},
		policyResolver,
		&deleteExecutor{},
		observer,
	)
}
