package health

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/circuitbreaker"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"golang.org/x/sync/singleflight"

	"go-fiber/pkg/db"
)

// Public groups public health endpoint behavior and resources.
type Public struct {
	database     db.DB
	isProduction bool
	cacheTTL     time.Duration

	cb *circuitbreaker.CircuitBreaker

	group singleflight.Group
	mu    sync.RWMutex
	cache readinessCache
}

// NewPublic builds the health runtime for public endpoints.
func NewPublic(database db.DB, cfg ReadyConfig) *Public {
	cfg = normalizeReadyConfig(cfg)

	p := &Public{
		database:     database,
		isProduction: cfg.IsProduction,
		cacheTTL:     cfg.CacheTTL,
	}

	p.cb = circuitbreaker.New(circuitbreaker.Config{
		FailureThreshold:      cfg.CBFailureThreshold,
		Timeout:               cfg.CBTimeout,
		SuccessThreshold:      cfg.CBSuccessThreshold,
		HalfOpenMaxConcurrent: cfg.CBHalfOpenMaxConcurrent,
		IsFailure: func(c fiber.Ctx, err error) bool {
			return err != nil || c.Response().StatusCode() >= fiber.StatusInternalServerError
		},
		OnOpen: func(c fiber.Ctx) error {
			return p.unavailableResponse(c, ReasonCircuitOpen)
		},
		OnHalfOpen: func(c fiber.Ctx) error {
			return p.unavailableResponse(c, ReasonCircuitHalf)
		},
		OnClose: func(c fiber.Ctx) error {
			return c.Next()
		},
	})

	return p
}

// Register adds liveness, readiness and startup endpoints.
func (p *Public) Register(r fiber.Router) {
	r.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	r.Get(healthcheck.ReadinessEndpoint, circuitbreaker.Middleware(p.cb), p.readyz)
	r.Get(healthcheck.StartupEndpoint, healthcheck.New())
}

// Stop releases circuit breaker resources.
func (p *Public) Stop() {
	if p == nil || p.cb == nil {
		return
	}
	p.cb.Stop()
}
