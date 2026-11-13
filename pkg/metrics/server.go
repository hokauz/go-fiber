package metrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go-fiber/pkg/netutil"
)

const defaultMetricsPort = "9090"

// NewServer returns a dedicated HTTP server for Prometheus metrics.
func NewServer(addr, path string, registry *prometheus.Registry) *http.Server {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	mux := http.NewServeMux()
	mux.Handle(normalizeMetricsPath(path), promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	return &http.Server{
		Addr:              netutil.NormalizeListenAddr(addr, defaultMetricsPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func normalizeMetricsPath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return "/metrics"
	}
	if strings.HasPrefix(normalized, "/") {
		return normalized
	}
	return "/" + normalized
}

