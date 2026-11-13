package mid

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	_ "github.com/jackc/pgx/v5/stdlib"

	apienv "go-fiber/cmd/api/env"
)

type fakeDB struct {
	raw *sql.DB
}

func (f *fakeDB) Ping(context.Context) error                           { return nil }
func (f *fakeDB) Readiness(context.Context) bool                       { return true }
func (f *fakeDB) WithTx(context.Context, func(tx *sql.Tx) error) error { return nil }
func (f *fakeDB) Raw() *sql.DB                                         { return f.raw }
func (f *fakeDB) Close() error                                         { return nil }

func TestNewMetrics_DisabledOrNilConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  *apienv.MetricsEnvs
	}{
		{name: "nil config", cfg: nil},
		{name: "disabled", cfg: &apienv.MetricsEnvs{Enabled: false}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rt := NewMetrics(tc.cfg, nil)
			if rt == nil {
				t.Fatal("runtime should not be nil")
			}
			if rt.Server() != nil {
				t.Fatal("metrics server should be nil when disabled")
			}

			app := fiber.New()
			app.Use(rt.Handler())
			app.Get("/ok", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusNoContent)
			})

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ok", nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
			}
		})
	}
}

func TestNewMetrics_Enabled_ConfiguresHandlerAndDedicatedServer(t *testing.T) {
	raw, err := sql.Open("pgx", "postgres://user:pass@127.0.0.1:5432/db?sslmode=disable")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer raw.Close()
	raw.SetMaxOpenConns(9)

	rt := NewMetrics(&apienv.MetricsEnvs{
		Enabled: true,
		Port:    "9090",
		Path:    "custom-metrics",
		Prefix:  "my_api",
	}, &fakeDB{raw: raw})
	if rt == nil {
		t.Fatal("runtime should not be nil")
	}
	if rt.Server() == nil {
		t.Fatal("metrics server should not be nil when enabled")
	}

	app := fiber.New()
	app.Use(rt.Handler())
	app.Get("/api/ping", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/ping", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	rec := httptest.NewRecorder()
	rt.Server().Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/custom-metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", rec.Code, http.StatusOK)
	}

	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read metrics body: %v", err)
	}

	text := string(body)
	if !strings.Contains(text, "my_api_http_requests_total") {
		t.Fatal("expected request total metric in scrape output")
	}
	if !strings.Contains(text, "my_api_db_connections_max") {
		t.Fatal("expected db pool metric in scrape output")
	}
	if !strings.Contains(text, "path=\"/api/ping\"") {
		t.Fatal("expected route path label for instrumented endpoint")
	}
}
