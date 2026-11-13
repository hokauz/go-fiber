package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

type fakeDB struct {
	ready bool
	sleep time.Duration
	calls int32
}

func (f *fakeDB) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeDB) Readiness(ctx context.Context) bool {
	atomic.AddInt32(&f.calls, 1)
	if f.sleep > 0 {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(f.sleep):
		}
	}
	return f.ready
}

func (f *fakeDB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return nil
}

func (f *fakeDB) Raw() *sql.DB {
	return nil
}

func (f *fakeDB) Close() error {
	return nil
}

func TestPublic_Register_LivezStartupzAlwaysOK(t *testing.T) {
	app := fiber.New()
	p := NewPublic(nil, ReadyConfig{})
	defer p.Stop()
	p.Register(app.Group("/api"))

	for _, path := range []string{"/api/livez", "/api/startupz"} {
		req := httptest.NewRequest("GET", path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %s failed: %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status for %s = %d, want %d", path, resp.StatusCode, fiber.StatusOK)
		}
	}
}

func TestPublic_Register_ReadyzStateResponses(t *testing.T) {
	t.Run("db not initialized in development", func(t *testing.T) {
		app := fiber.New()
		p := NewPublic(nil, ReadyConfig{})
		defer p.Stop()
		p.Register(app.Group("/api"))

		resp := doRequest(t, app, "/api/readyz")
		if resp.StatusCode != fiber.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
		}

		body := decodeBody(t, resp)
		if body["status"] != "unavailable" {
			t.Fatalf("status body = %v", body["status"])
		}
		if body["reason"] != "not_initialized" {
			t.Fatalf("reason body = %v", body["reason"])
		}
	})

	t.Run("db unhealthy in development", func(t *testing.T) {
		app := fiber.New()
		p := NewPublic(&fakeDB{ready: false}, ReadyConfig{})
		defer p.Stop()
		p.Register(app.Group("/api"))

		resp := doRequest(t, app, "/api/readyz")
		if resp.StatusCode != fiber.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
		}

		body := decodeBody(t, resp)
		if body["status"] != "unavailable" {
			t.Fatalf("status body = %v", body["status"])
		}
		if body["reason"] != "unhealthy" {
			t.Fatalf("reason body = %v", body["reason"])
		}
	})

	t.Run("db ready", func(t *testing.T) {
		app := fiber.New()
		p := NewPublic(&fakeDB{ready: true}, ReadyConfig{})
		defer p.Stop()
		p.Register(app.Group("/api"))

		resp := doRequest(t, app, "/api/readyz")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
		}

		body := decodeBody(t, resp)
		if body["status"] != "ok" {
			t.Fatalf("status body = %v", body["status"])
		}
	})
}

func TestPublic_Register_ReadyzProductionMinimalBody(t *testing.T) {
	app := fiber.New()
	p := NewPublic(nil, ReadyConfig{IsProduction: true})
	defer p.Stop()
	p.Register(app.Group("/api"))

	resp := doRequest(t, app, "/api/readyz")
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
	}

	body := decodeBody(t, resp)
	if body["status"] != "unavailable" {
		t.Fatalf("status body = %v", body["status"])
	}
	if _, ok := body["reason"]; ok {
		t.Fatalf("reason must not be present in production body: %v", body)
	}
}

func TestPublic_ReadinessCacheTTL(t *testing.T) {
	fdb := &fakeDB{ready: true}
	p := NewPublic(fdb, ReadyConfig{CacheTTL: 200 * time.Millisecond})
	defer p.Stop()

	first := p.readiness(context.Background())
	second := p.readiness(context.Background())
	if !first.ready || !second.ready {
		t.Fatal("expected readiness true")
	}
	if calls := atomic.LoadInt32(&fdb.calls); calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}

	time.Sleep(250 * time.Millisecond)
	third := p.readiness(context.Background())
	if !third.ready {
		t.Fatal("expected readiness true after cache expiration")
	}
	if calls := atomic.LoadInt32(&fdb.calls); calls != 2 {
		t.Fatalf("calls after expiration = %d, want 2", calls)
	}
}

func TestPublic_ReadinessCacheTTLStartsAfterProbeCompletion(t *testing.T) {
	fdb := &fakeDB{ready: true, sleep: 120 * time.Millisecond}
	p := NewPublic(fdb, ReadyConfig{CacheTTL: 50 * time.Millisecond})
	defer p.Stop()

	first := p.readiness(context.Background())
	if !first.ready {
		t.Fatalf("first readiness = %#v, want ready", first)
	}

	second := p.readiness(context.Background())
	if !second.ready {
		t.Fatalf("second readiness = %#v, want ready", second)
	}

	if calls := atomic.LoadInt32(&fdb.calls); calls != 1 {
		t.Fatalf("calls right after slow probe = %d, want 1", calls)
	}

	time.Sleep(70 * time.Millisecond)
	third := p.readiness(context.Background())
	if !third.ready {
		t.Fatalf("third readiness = %#v, want ready", third)
	}

	if calls := atomic.LoadInt32(&fdb.calls); calls != 2 {
		t.Fatalf("calls after cache expiration = %d, want 2", calls)
	}
}

func TestPublic_ReadinessSingleflightDedupesConcurrentChecks(t *testing.T) {
	fdb := &fakeDB{ready: true, sleep: 50 * time.Millisecond}
	p := NewPublic(fdb, ReadyConfig{CacheTTL: 500 * time.Millisecond})
	defer p.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			st := p.readiness(context.Background())
			if !st.ready {
				t.Errorf("expected ready status")
			}
		}()
	}
	wg.Wait()

	if calls := atomic.LoadInt32(&fdb.calls); calls != 1 {
		t.Fatalf("readiness calls = %d, want 1", calls)
	}
}

func TestPublic_ReadinessCanceledContextDoesNotPoisonSharedResult(t *testing.T) {
	fdb := &fakeDB{ready: true, sleep: 50 * time.Millisecond}
	p := NewPublic(fdb, ReadyConfig{CacheTTL: 500 * time.Millisecond})
	defer p.Stop()

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	first := p.readiness(canceledCtx)
	if !first.ready {
		t.Fatalf("first readiness = %#v, want ready", first)
	}

	second := p.readiness(context.Background())
	if !second.ready {
		t.Fatalf("second readiness = %#v, want ready", second)
	}

	if calls := atomic.LoadInt32(&fdb.calls); calls != 1 {
		t.Fatalf("readiness calls = %d, want 1", calls)
	}
}

func TestPublic_Register_CircuitBreakerBlocksAfterThreshold(t *testing.T) {
	app := fiber.New()
	p := NewPublic(&fakeDB{ready: false}, ReadyConfig{
		CBFailureThreshold:      1,
		CBTimeout:               2 * time.Second,
		CBSuccessThreshold:      1,
		CBHalfOpenMaxConcurrent: 1,
		CacheTTL:                0,
	})
	defer p.Stop()
	p.Register(app.Group("/api"))

	first := decodeBody(t, doRequest(t, app, "/api/readyz"))
	if first["reason"] != "unhealthy" {
		t.Fatalf("first reason = %v, want unhealthy", first["reason"])
	}

	second := decodeBody(t, doRequest(t, app, "/api/readyz"))
	if second["reason"] != "circuit_open" {
		t.Fatalf("second reason = %v, want circuit_open", second["reason"])
	}
}

func doRequest(t *testing.T, app *fiber.App, path string) *http.Response {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeBody(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return body
}
