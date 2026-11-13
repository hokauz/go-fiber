package logx

import "context"

type contextKey struct{}

var loggerKey = contextKey{}

// WithLogger returns a derived context carrying the provided logger.
func WithLogger(ctx context.Context, l Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = Default()
	}
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext extracts a logger from context or returns Default().
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return Default()
	}
	l, ok := ctx.Value(loggerKey).(Logger)
	if !ok || l == nil {
		return Default()
	}
	return l
}
