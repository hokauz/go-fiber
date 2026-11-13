package slog

import (
	"context"
	"fmt"
	"io"
	stdslog "log/slog"
	"os"
	"strings"

	"go-fiber/pkg/logx"
)

type logger struct {
	inner *stdslog.Logger
}

// New builds a logx logger backed by log/slog.
func New(cfg Config) (logx.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(strings.TrimSpace(cfg.Format))
	if format == "" {
		format = "json"
	}

	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	handler, err := newHandler(format, writer, &stdslog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	})
	if err != nil {
		return nil, err
	}

	base := stdslog.New(handler)
	if cfg.Service != "" {
		base = base.With(stdslog.String("service", cfg.Service))
	}
	if cfg.Environment != "" {
		base = base.With(stdslog.String("environment", cfg.Environment))
	}

	return &logger{inner: base}, nil
}

func newHandler(format string, w io.Writer, opts *stdslog.HandlerOptions) (stdslog.Handler, error) {
	switch format {
	case "json":
		return stdslog.NewJSONHandler(w, opts), nil
	case "text":
		return stdslog.NewTextHandler(w, opts), nil
	default:
		return nil, fmt.Errorf("invalid log format %q: use json or text", format)
	}
}

func parseLevel(level string) (stdslog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return stdslog.LevelInfo, nil
	case "debug":
		return stdslog.LevelDebug, nil
	case "warn", "warning":
		return stdslog.LevelWarn, nil
	case "error":
		return stdslog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q: use debug, info, warn, error", level)
	}
}

func (l *logger) Debug(msg string, attrs ...logx.Attr) {
	l.log(stdslog.LevelDebug, msg, attrs...)
}

func (l *logger) Info(msg string, attrs ...logx.Attr) {
	l.log(stdslog.LevelInfo, msg, attrs...)
}

func (l *logger) Warn(msg string, attrs ...logx.Attr) {
	l.log(stdslog.LevelWarn, msg, attrs...)
}

func (l *logger) Error(msg string, attrs ...logx.Attr) {
	l.log(stdslog.LevelError, msg, attrs...)
}

func (l *logger) Fatal(msg string, err error) {
	if err != nil {
		l.log(stdslog.LevelError, msg, logx.Error(err))
	} else {
		l.log(stdslog.LevelError, msg)
	}
	os.Exit(1)
}

func (l *logger) With(attrs ...logx.Attr) logx.Logger {
	if l == nil || l.inner == nil {
		return logx.NewNoop()
	}
	if len(attrs) == 0 {
		return l
	}
	return &logger{inner: l.inner.With(toAnyArgs(attrs)...)}
}

func (l *logger) log(level stdslog.Level, msg string, attrs ...logx.Attr) {
	if l == nil || l.inner == nil {
		return
	}
	l.inner.LogAttrs(context.Background(), level, msg, toSlogAttrs(attrs)...)
}

func toSlogAttrs(attrs []logx.Attr) []stdslog.Attr {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]stdslog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		key := strings.TrimSpace(attr.Key)
		if key == "" {
			continue
		}
		out = append(out, stdslog.Any(key, sanitizeAttr(key, attr.Value)))
	}
	return out
}

func toAnyArgs(attrs []logx.Attr) []any {
	slogAttrs := toSlogAttrs(attrs)
	if len(slogAttrs) == 0 {
		return nil
	}
	out := make([]any, 0, len(slogAttrs))
	for _, attr := range slogAttrs {
		out = append(out, attr)
	}
	return out
}
