package router

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

// HTTPTimeouts configures HTTP server timeouts in Fiber.
type HTTPTimeouts struct {
	Read  time.Duration
	Write time.Duration
	Idle  time.Duration
}

// Config returns Fiber config with sonic as JSON encoder/decoder.
// Sonic provides faster JSON serialization than encoding/json.
func Config(timeouts ...HTTPTimeouts) fiber.Config {
	cfg := fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	}

	if len(timeouts) == 0 {
		return cfg
	}

	t := timeouts[0]
	cfg.ReadTimeout = t.Read
	cfg.WriteTimeout = t.Write
	cfg.IdleTimeout = t.Idle
	return cfg
}

// Registrable is a function that registers routes on a router.
type Registrable func(fiber.Router)

// Group wraps a fiber router and adds a Register method.
type Group struct {
	fiber.Router
}

// NewGroup returns a Group that can register handlers via Register.
// Optional middlewares are applied to the router before returning.
func NewGroup(r fiber.Router, middlewares ...fiber.Handler) *Group {
	for _, mw := range middlewares {
		r.Use(mw)
	}
	return &Group{Router: r}
}

// Register calls the given Registrable with this group's router.
func (g *Group) Register(reg Registrable) {
	reg(g.Router)
}
