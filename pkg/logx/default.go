package logx

import "sync"

var (
	defaultLoggerMu sync.RWMutex
	defaultLogger   = NewNoop()
)

// SetDefault registers the global logger used as fallback.
func SetDefault(l Logger) {
	if l == nil {
		l = NewNoop()
	}

	defaultLoggerMu.Lock()
	defaultLogger = l
	defaultLoggerMu.Unlock()
}

// Default returns the global logger fallback.
func Default() Logger {
	defaultLoggerMu.RLock()
	l := defaultLogger
	defaultLoggerMu.RUnlock()

	if l == nil {
		return NewNoop()
	}
	return l
}
