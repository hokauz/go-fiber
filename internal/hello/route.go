package hello

import (
	"go-fiber/pkg/db"

	"github.com/gofiber/fiber/v3"
)

// PublicRegister adds public hello endpoints.
func PublicRegister(db db.DB) func(fiber.Router) {
	return func(r fiber.Router) {
		svc := NewService(db)
		h := NewHandler(svc)
		r.Get("/v1/hello", h.HelloDoc)
		r.Get("/v1/hello/db-ping", h.DBPingDoc)
	}
}

// PrivateRegister adds private hello endpoints (requires auth).
func PrivateRegister(db db.DB) func(fiber.Router) {
	return func(r fiber.Router) {
		svc := NewService(db)
		h := NewHandler(svc)
		r.Get("/v1/hello/me", h.HelloMeDoc)
	}
}
