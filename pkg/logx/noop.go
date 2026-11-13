package logx

import "os"

type noopLogger struct{}

// NewNoop returns a no-op logger implementation.
func NewNoop() Logger {
	return noopLogger{}
}

func (noopLogger) Debug(string, ...Attr) {}
func (noopLogger) Info(string, ...Attr)  {}
func (noopLogger) Warn(string, ...Attr)  {}
func (noopLogger) Error(string, ...Attr) {}
func (noopLogger) Fatal(string, error)   { os.Exit(1) }
func (n noopLogger) With(...Attr) Logger { return n }
