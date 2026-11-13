package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCollector_RegistersHTTPAndRuntimeMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	col := NewCollector("go_fiber_api", reg)
	col.ObserveRequest("GET", "/api/livez", 200, time.Millisecond)

	if col.Registry() != reg {
		t.Fatal("collector should expose the same registry instance")
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	names := make(map[string]bool, len(mfs))
	for _, mf := range mfs {
		names[mf.GetName()] = true
	}

	for _, want := range []string{
		"go_fiber_api_http_requests_total",
		"go_fiber_api_http_request_duration_seconds",
		"go_fiber_api_http_requests_in_flight",
	} {
		if !names[want] {
			t.Fatalf("missing metric family: %s", want)
		}
	}

	hasGoRuntime := false
	hasProcessRuntime := false
	for name := range names {
		if strings.HasPrefix(name, "go_") {
			hasGoRuntime = true
		}
		if strings.HasPrefix(name, "process_") {
			hasProcessRuntime = true
		}
	}
	if !hasGoRuntime {
		t.Fatal("expected at least one go_* runtime metric")
	}
	if !hasProcessRuntime {
		t.Fatal("expected at least one process_* runtime metric")
	}
}

func TestCollector_ObserveRequest(t *testing.T) {
	col := NewCollector("test_api", nil)

	col.ObserveRequest("GET", "/users/:id", 200, 150*time.Millisecond)

	mfs, err := col.Registry().Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	labels := map[string]string{
		"method": "GET",
		"path":   "/users/:id",
		"status": "200",
	}

	counter := findCounterValue(mfs, "test_api_http_requests_total", labels)
	if counter != 1 {
		t.Fatalf("counter = %v, want 1", counter)
	}

	hCount := findHistogramCount(mfs, "test_api_http_request_duration_seconds", labels)
	if hCount != 1 {
		t.Fatalf("histogram count = %d, want 1", hCount)
	}
}
