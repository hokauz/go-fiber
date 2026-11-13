package slog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go-fiber/pkg/logx"
)

func TestNew_EmitsJSONAndText(t *testing.T) {
	var jsonBuf bytes.Buffer
	jsonLogger, err := New(Config{Format: "json", Writer: &jsonBuf})
	if err != nil {
		t.Fatalf("New(json) error: %v", err)
	}
	jsonLogger.Info("hello", logx.String("component", "api"))

	var line map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(jsonBuf.Bytes()), &line); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if line["msg"] != "hello" {
		t.Fatalf("msg = %v, want hello", line["msg"])
	}
	if line["component"] != "api" {
		t.Fatalf("component = %v, want api", line["component"])
	}

	var textBuf bytes.Buffer
	textLogger, err := New(Config{Format: "text", Writer: &textBuf})
	if err != nil {
		t.Fatalf("New(text) error: %v", err)
	}
	textLogger.Info("world", logx.String("component", "worker"))

	out := textBuf.String()
	if !strings.Contains(out, "msg=world") {
		t.Fatalf("text output missing msg: %q", out)
	}
	if !strings.Contains(out, "component=worker") {
		t.Fatalf("text output missing attr: %q", out)
	}
}

func TestNew_LevelMapping(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(Config{Format: "json", Level: "warn", Writer: &buf})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	l.Info("ignore-info")
	l.Warn("keep-warn")

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(lines))
	}

	var row map[string]any
	if err := json.Unmarshal(lines[0], &row); err != nil {
		t.Fatalf("invalid json line: %v", err)
	}
	if row["msg"] != "keep-warn" {
		t.Fatalf("msg = %v, want keep-warn", row["msg"])
	}
}

func TestNew_SanitizesSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(Config{Format: "json", Writer: &buf})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	l.Info("auth", logx.String("authorization", "Bearer abc"), logx.Any("payload", map[string]any{
		"user":     "alice",
		"password": "my-pass",
		"nested": map[string]any{
			"api_token": "token-123",
		},
	}))

	var row map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &row); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}

	if row["authorization"] != redactedValue {
		t.Fatalf("authorization = %v, want %s", row["authorization"], redactedValue)
	}

	payload, ok := row["payload"].(map[string]any)
	if !ok {
		t.Fatalf("payload type = %T, want map[string]any", row["payload"])
	}
	if payload["password"] != redactedValue {
		t.Fatalf("payload.password = %v, want %s", payload["password"], redactedValue)
	}

	nested, ok := payload["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested type = %T, want map[string]any", payload["nested"])
	}
	if nested["api_token"] != redactedValue {
		t.Fatalf("nested.api_token = %v, want %s", nested["api_token"], redactedValue)
	}
}
