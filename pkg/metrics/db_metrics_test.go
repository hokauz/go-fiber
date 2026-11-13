package metrics

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCollector_RegisterDB_RegistersPoolMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	col := NewCollector("db_api", reg)

	raw, err := sql.Open("pgx", "postgres://user:pass@127.0.0.1:5432/db?sslmode=disable")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer raw.Close()
	raw.SetMaxOpenConns(13)

	col.RegisterDB(raw)
	col.RegisterDB(raw)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	inUse := findGaugeValueWithLabels(mfs, "db_api_db_connections_open", map[string]string{"state": "in_use"})
	if inUse != 0 {
		t.Fatalf("in_use connections = %v, want 0", inUse)
	}
	idle := findGaugeValueWithLabels(mfs, "db_api_db_connections_open", map[string]string{"state": "idle"})
	if idle != 0 {
		t.Fatalf("idle connections = %v, want 0", idle)
	}

	max := findGaugeValue(mfs, "db_api_db_connections_max")
	if max != 13 {
		t.Fatalf("max connections = %v, want 13", max)
	}
}

func TestCollector_ObserveQuery(t *testing.T) {
	reg := prometheus.NewRegistry()
	col := NewCollector("db_api", reg)

	col.ObserveQuery("select", "acl_groups", 25*time.Millisecond, nil)
	col.ObserveQuery("drop", "", 10*time.Millisecond, errors.New("boom"))

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	okLabels := map[string]string{
		"operation": "SELECT",
		"table":     "acl_groups",
		"status":    "ok",
	}
	if got := findCounterValue(mfs, "db_api_db_queries_total", okLabels); got != 1 {
		t.Fatalf("ok queries = %v, want 1", got)
	}

	errorLabels := map[string]string{
		"operation": "UNKNOWN",
		"table":     "unknown",
		"status":    "error",
	}
	if got := findCounterValue(mfs, "db_api_db_queries_total", errorLabels); got != 1 {
		t.Fatalf("error queries = %v, want 1", got)
	}

	if count := findHistogramCount(mfs, "db_api_db_query_duration_seconds", map[string]string{
		"operation": "SELECT",
		"table":     "acl_groups",
	}); count != 1 {
		t.Fatalf("ok query histogram count = %d, want 1", count)
	}

	if count := findHistogramCount(mfs, "db_api_db_query_duration_seconds", map[string]string{
		"operation": "UNKNOWN",
		"table":     "unknown",
	}); count != 1 {
		t.Fatalf("error query histogram count = %d, want 1", count)
	}
}
