package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

func (p *Public) readyz(c fiber.Ctx) error {
	status := p.readiness(c.Context())
	if status.ready {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	}
	return p.unavailableResponse(c, status.reason)
}

func (p *Public) readiness(ctx context.Context) readinessStatus {
	ctx = sharedReadinessContext(ctx)

	now := time.Now()

	if cached, ok := p.cached(now); ok {
		return cached
	}

	v, _, _ := p.group.Do(readySingleflightKey, func() (any, error) {
		if cached, ok := p.cached(now); ok {
			return cached, nil
		}

		result := p.computeReadiness(ctx)
		p.storeCache(time.Now().Add(p.cacheTTL), result)
		return result, nil
	})

	ready, ok := v.(readinessStatus)
	if !ok {
		return readinessStatus{ready: false, reason: ReasonUnknown}
	}
	return ready
}

func sharedReadinessContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

func (p *Public) cached(now time.Time) (readinessStatus, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.cache.valid || now.After(p.cache.expiresAt) {
		return readinessStatus{}, false
	}
	return p.cache.status, true
}

func (p *Public) storeCache(expiresAt time.Time, status readinessStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = readinessCache{
		status:    status,
		expiresAt: expiresAt,
		valid:     true,
	}
}

func (p *Public) computeReadiness(ctx context.Context) readinessStatus {
	if p == nil || p.database == nil {
		return readinessStatus{ready: false, reason: ReasonNotInitialized}
	}
	if !p.database.Readiness(ctx) {
		return readinessStatus{ready: false, reason: ReasonUnhealthy}
	}
	return readinessStatus{ready: true}
}

func (p *Public) unavailableResponse(c fiber.Ctx, reason string) error {
	if p.isProduction {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
	}
	return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
		"status": "unavailable",
		"reason": reason,
	})
}
