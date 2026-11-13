package mid

import (
	"net/http"

	"github.com/gofiber/fiber/v3"

	apienv "go-fiber/cmd/api/env"
	"go-fiber/pkg/db"
	appmetrics "go-fiber/pkg/metrics"
)

var noopMetricsHandler = func(c fiber.Ctx) error {
	return c.Next()
}

// MetricsRuntime groups metrics middleware instrumentation and optional dedicated server.
type MetricsRuntime struct {
	handler fiber.Handler
	server  *http.Server
	col     *appmetrics.Collector
}

// NewMetrics builds metrics runtime from API env configuration.
func NewMetrics(cfg *apienv.MetricsEnvs, database db.DB) *MetricsRuntime {
	rt := &MetricsRuntime{handler: noopMetricsHandler}
	if cfg == nil || !cfg.Enabled {
		return rt
	}

	collector := appmetrics.NewCollector(cfg.Prefix, nil)
	if database != nil {
		collector.RegisterDB(database.Raw())
	}
	rt.server = appmetrics.NewServer(cfg.Port, cfg.Path, collector.Registry())
	rt.handler = appmetrics.Middleware(collector)
	rt.col = collector
	return rt
}

// Handler returns the Fiber middleware that records request metrics.
func (m *MetricsRuntime) Handler() fiber.Handler {
	if m == nil {
		return noopMetricsHandler
	}
	return m.handler
}

// Server returns the optional dedicated metrics HTTP server.
func (m *MetricsRuntime) Server() *http.Server {
	if m == nil {
		return nil
	}
	return m.server
}

// Collector returns the underlying metrics collector.
func (m *MetricsRuntime) Collector() *appmetrics.Collector {
	if m == nil {
		return nil
	}
	return m.col
}
