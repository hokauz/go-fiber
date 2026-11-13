package mid

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	adapterslog "go-fiber/adapter/log/slog"
	"go-fiber/pkg/logx"
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
)

func TestLoggerWith_EmitsSingleCanonicalAccessLog(t *testing.T) {
	var buf bytes.Buffer
	base := newTestLogger(t, &buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(LogContext(base))
	app.Use(LoggerWith(base))
	app.Get("/ok", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	lines := decodeJSONLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("log lines = %d, want 1", len(lines))
	}

	line := lines[0]
	if line["msg"] != "http.access" {
		t.Fatalf("msg = %v, want http.access", line["msg"])
	}
	if line["request_id"] == "" {
		t.Fatal("expected request_id in access log")
	}
	if line["status"] != float64(fiber.StatusOK) {
		t.Fatalf("status = %v, want %d", line["status"], fiber.StatusOK)
	}
	if _, ok := line["duration_ms"].(float64); !ok {
		t.Fatalf("duration_ms type = %T, want float64", line["duration_ms"])
	}
}

func TestLoggerWith_SkipsHealthEndpoints(t *testing.T) {
	var buf bytes.Buffer
	base := newTestLogger(t, &buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(LogContext(base))
	app.Use(LoggerWith(base))
	app.Get("/api/livez", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/livez", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	lines := decodeJSONLines(t, &buf)
	if len(lines) != 0 {
		t.Fatalf("log lines = %d, want 0", len(lines))
	}
}

func TestRecovererWith_LogsStructuredPanic(t *testing.T) {
	var buf bytes.Buffer
	base := newTestLogger(t, &buf)

	app := fiber.New()
	app.Use(RecovererWith(base))
	app.Use(RequestID())
	app.Use(LogContext(base))
	app.Use(LoggerWith(base))
	app.Get("/panic", func(c fiber.Ctx) error {
		panic("boom")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusInternalServerError)
	}

	lines := decodeJSONLines(t, &buf)
	if len(lines) == 0 {
		t.Fatal("expected at least one log line")
	}

	var panicLine map[string]any
	for _, line := range lines {
		if line["msg"] == "panic recovered" {
			panicLine = line
			break
		}
	}
	if panicLine == nil {
		t.Fatalf("panic log not found in %d lines", len(lines))
	}
	if panicLine["request_id"] == "" {
		t.Fatal("expected request_id in panic log")
	}
	if panicLine["path"] != "/panic" {
		t.Fatalf("path = %v, want /panic", panicLine["path"])
	}
	if panicLine["method"] != "GET" {
		t.Fatalf("method = %v, want GET", panicLine["method"])
	}
	if panicLine["stack"] == "" {
		t.Fatal("expected stack in panic log")
	}
}

func TestLogContext_ProvidesLoggerFromContext(t *testing.T) {
	var buf bytes.Buffer
	base := newTestLogger(t, &buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(LogContext(base))
	app.Get("/ctx", func(c fiber.Ctx) error {
		logx.FromContext(c.Context()).Info("handler.log", logx.String("rid_ctx", reqctx.RequestID(c)))
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/ctx", nil)
	req.Header.Set(fiber.HeaderXRequestID, "req-123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}

	lines := decodeJSONLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("log lines = %d, want 1", len(lines))
	}
	if lines[0]["request_id"] != "req-123" {
		t.Fatalf("request_id = %v, want req-123", lines[0]["request_id"])
	}
	if lines[0]["rid_ctx"] != "req-123" {
		t.Fatalf("rid_ctx = %v, want req-123", lines[0]["rid_ctx"])
	}
}

func newTestLogger(t *testing.T, buf *bytes.Buffer) logx.Logger {
	t.Helper()

	base, err := adapterslog.New(adapterslog.Config{
		Format: "json",
		Level:  "debug",
		Writer: buf,
	})
	if err != nil {
		t.Fatalf("logger init failed: %v", err)
	}

	logx.SetDefault(base)
	t.Cleanup(func() {
		logx.SetDefault(logx.NewNoop())
	})

	return base
}

func decodeJSONLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	raw := bytes.TrimSpace(buf.Bytes())
	if len(raw) == 0 {
		return nil
	}

	rows := bytes.Split(raw, []byte("\n"))
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if len(bytes.TrimSpace(row)) == 0 {
			continue
		}
		var line map[string]any
		if err := json.Unmarshal(row, &line); err != nil {
			t.Fatalf("invalid json line: %v", err)
		}
		out = append(out, line)
	}
	return out
}
