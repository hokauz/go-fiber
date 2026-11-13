package hello

import (
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
)

// Handler handles hello endpoints.
type Handler struct {
	svc *Service
}

// NewHandler returns a new Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Hello returns the current request id (public endpoint).
func (h *Handler) Hello(c fiber.Ctx) error {
	return c.JSON(Response{RequestID: reqctx.RequestID(c)})
}

// HelloMe returns the current request id and the authenticated user id.
func (h *Handler) HelloMe(c fiber.Ctx) error {
	p, err := reqctx.PrincipalOrErr(c)
	if err != nil {
		return err
	}
	return c.JSON(PrivateResponse{
		RequestID: reqctx.RequestID(c),
		UserID:    p.UserID,
	})
}

// DBPing checks if DB connectivity is available.
func (h *Handler) DBPing(c fiber.Ctx) error {
	status := h.svc.Ping(c.Context())
	if status == PingOK {
		return c.JSON(fiber.Map{"status": "ok"})
	}
	return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": string(status)})
}
