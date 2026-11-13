package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type dbPoolCollector struct {
	raw                       *sql.DB
	dbConnectionsOpen         *prometheus.Desc
	dbConnectionsMax          *prometheus.Desc
	dbConnectionsWaitTotal    *prometheus.Desc
	dbConnectionsWaitDuration *prometheus.Desc
}

func newDBPoolCollector(prefix string, raw *sql.DB) *dbPoolCollector {
	return &dbPoolCollector{
		raw: raw,
		dbConnectionsOpen: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "", "db_connections_open"),
			"Open DB connections grouped by pool state.",
			[]string{"state"},
			nil,
		),
		dbConnectionsMax: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "", "db_connections_max"),
			"Configured maximum number of open DB connections.",
			nil,
			nil,
		),
		dbConnectionsWaitTotal: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "", "db_connections_wait_total"),
			"Total number of waits for a DB connection.",
			nil,
			nil,
		),
		dbConnectionsWaitDuration: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "", "db_connections_wait_duration_seconds"),
			"Total time blocked waiting for a DB connection in seconds.",
			nil,
			nil,
		),
	}
}

func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	if c == nil {
		return
	}
	ch <- c.dbConnectionsOpen
	ch <- c.dbConnectionsMax
	ch <- c.dbConnectionsWaitTotal
	ch <- c.dbConnectionsWaitDuration
}

func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	if c == nil || c.raw == nil {
		return
	}

	stats := c.raw.Stats()

	ch <- prometheus.MustNewConstMetric(c.dbConnectionsOpen, prometheus.GaugeValue, float64(stats.InUse), "in_use")
	ch <- prometheus.MustNewConstMetric(c.dbConnectionsOpen, prometheus.GaugeValue, float64(stats.Idle), "idle")
	ch <- prometheus.MustNewConstMetric(c.dbConnectionsMax, prometheus.GaugeValue, float64(stats.MaxOpenConnections))
	ch <- prometheus.MustNewConstMetric(c.dbConnectionsWaitTotal, prometheus.CounterValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.dbConnectionsWaitDuration, prometheus.CounterValue, stats.WaitDuration.Seconds())
}
