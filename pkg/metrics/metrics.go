package metrics

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const defaultPrefix = "go_fiber_api"

var metricPrefixSanitizer = strings.NewReplacer("-", "_", ".", "_", " ", "_", "/", "_")

const (
	unknownOperation = "UNKNOWN"
	unknownTable     = "unknown"
	statusOK         = "ok"
	statusError      = "error"
)

// Collector holds HTTP metrics and its backing registry.
type Collector struct {
	prefix                     string
	registry                   *prometheus.Registry
	httpRequestsTotal          *prometheus.CounterVec
	httpRequestDurationSeconds *prometheus.HistogramVec
	httpRequestsInFlight       prometheus.Gauge
	dbQueriesTotal             *prometheus.CounterVec
	dbQueryDurationSeconds     *prometheus.HistogramVec
	dbPoolCollector            *dbPoolCollector
	mu                         sync.Mutex
}

// NewCollector builds a metrics collector with an isolated registry when reg is nil.
func NewCollector(prefix string, reg *prometheus.Registry) *Collector {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	prefix = normalizePrefix(prefix)

	col := &Collector{
		prefix:   prefix,
		registry: reg,
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: prefix,
				Name:      "http_requests_total",
				Help:      "Total HTTP requests processed.",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: prefix,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		httpRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: prefix,
				Name:      "http_requests_in_flight",
				Help:      "Current number of in-flight HTTP requests.",
			},
		),
		dbQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: prefix,
				Name:      "db_queries_total",
				Help:      "Total database queries executed.",
			},
			[]string{"operation", "table", "status"},
		),
		dbQueryDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: prefix,
				Name:      "db_query_duration_seconds",
				Help:      "Database query latency in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
	}

	reg.MustRegister(
		col.httpRequestsTotal,
		col.httpRequestDurationSeconds,
		col.httpRequestsInFlight,
		col.dbQueriesTotal,
		col.dbQueryDurationSeconds,
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	return col
}

// Registry returns the collector registry.
func (c *Collector) Registry() *prometheus.Registry {
	if c == nil {
		return nil
	}
	return c.registry
}

func (c *Collector) ObserveRequest(method, path string, status int, duration time.Duration) {
	if c == nil {
		return
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "UNKNOWN"
	}
	path = NormalizePath(path)
	statusLabel := strconv.Itoa(status)

	c.httpRequestsTotal.WithLabelValues(method, path, statusLabel).Inc()
	c.httpRequestDurationSeconds.WithLabelValues(method, path, statusLabel).Observe(duration.Seconds())
}

// RegisterDB registers sql.DB pool metrics in the same collector registry.
// Calling more than once is safe; only the first non-nil DB is registered.
func (c *Collector) RegisterDB(raw *sql.DB) {
	if c == nil || raw == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dbPoolCollector != nil {
		return
	}

	poolCollector := newDBPoolCollector(c.prefix, raw)
	if err := c.registry.Register(poolCollector); err != nil {
		var alreadyRegisteredErr prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredErr) {
			c.dbPoolCollector = poolCollector
			return
		}
		return
	}

	c.dbPoolCollector = poolCollector
}

// ObserveQuery records DB query count and latency with controlled label sets.
func (c *Collector) ObserveQuery(operation, table string, duration time.Duration, err error) {
	if c == nil {
		return
	}

	op := normalizeDBOperation(operation)
	tbl := normalizeDBTable(table)
	status := statusOK
	if err != nil {
		status = statusError
	}

	c.dbQueriesTotal.WithLabelValues(op, tbl, status).Inc()
	c.dbQueryDurationSeconds.WithLabelValues(op, tbl).Observe(duration.Seconds())
}

func (c *Collector) IncInFlight() {
	if c == nil {
		return
	}
	c.httpRequestsInFlight.Inc()
}

func (c *Collector) DecInFlight() {
	if c == nil {
		return
	}
	c.httpRequestsInFlight.Dec()
}

func normalizePrefix(prefix string) string {
	normalized := metricPrefixSanitizer.Replace(strings.TrimSpace(prefix))
	if normalized == "" {
		return defaultPrefix
	}
	return normalized
}

func normalizeDBOperation(operation string) string {
	normalized := strings.ToUpper(strings.TrimSpace(operation))
	switch normalized {
	case "SELECT", "INSERT", "UPDATE", "DELETE":
		return normalized
	default:
		return unknownOperation
	}
}

func normalizeDBTable(table string) string {
	normalized := strings.TrimSpace(table)
	if normalized == "" {
		return unknownTable
	}
	return normalized
}
