package logx

import (
	"context"
	"testing"
)

type testLogger struct {
	attrs []Attr
}

func (l *testLogger) Debug(string, ...Attr) {}
func (l *testLogger) Info(string, ...Attr)  {}
func (l *testLogger) Warn(string, ...Attr)  {}
func (l *testLogger) Error(string, ...Attr) {}
func (l *testLogger) Fatal(string, error)   {}

func (l *testLogger) With(attrs ...Attr) Logger {
	out := make([]Attr, 0, len(l.attrs)+len(attrs))
	out = append(out, l.attrs...)
	out = append(out, attrs...)
	return &testLogger{attrs: out}
}

func TestWithLoggerFromContext(t *testing.T) {
	t.Cleanup(func() {
		SetDefault(NewNoop())
	})

	base := &testLogger{}
	ctx := WithLogger(context.Background(), base)

	got := FromContext(ctx)
	if got != base {
		t.Fatal("expected logger from context")
	}
}

func TestDefaultSetDefaultFallbackNoop(t *testing.T) {
	SetDefault(nil)

	if Default() == nil {
		t.Fatal("expected fallback noop logger")
	}

	// should not panic with noop fallback
	Default().Info("msg")
}

func TestWithPreservesAttributes(t *testing.T) {
	base := &testLogger{}

	next := base.With(String("request_id", "req-1"), Int("status", 200))
	next = next.With(Bool("ok", true))

	got, ok := next.(*testLogger)
	if !ok {
		t.Fatalf("expected *testLogger, got %T", next)
	}
	if len(got.attrs) != 3 {
		t.Fatalf("attrs len = %d, want 3", len(got.attrs))
	}
	if got.attrs[0].Key != "request_id" || got.attrs[1].Key != "status" || got.attrs[2].Key != "ok" {
		t.Fatalf("unexpected attrs order: %#v", got.attrs)
	}
}
