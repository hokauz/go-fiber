package hello

import (
	"github.com/gofiber/fiber/v3"
)

// HelloDoc godoc
// @Summary      Hello endpoint (public)
// @Description  Returns the current request id.
// @Tags         Example
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/hello [get]
func (h *Handler) HelloDoc(c fiber.Ctx) error {
	return h.Hello(c)
}

// HelloMeDoc godoc
// @Summary      Hello endpoint (private)
// @Description  Returns the current request id and the authenticated user id.
// @Tags         Example
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  PrivateResponse
// @Router       /api/v1/hello/me [get]
func (h *Handler) HelloMeDoc(c fiber.Ctx) error {
	return h.HelloMe(c)
}

// DBPingDoc godoc
// @Summary      DB ping endpoint (public)
// @Description  Returns DB connectivity status.
// @Tags         Example
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /api/v1/hello/db-ping [get]
func (h *Handler) DBPingDoc(c fiber.Ctx) error {
	return h.DBPing(c)
}
