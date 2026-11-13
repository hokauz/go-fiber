package logx

// Logger is the logging port used across the application.
type Logger interface {
	Debug(msg string, attrs ...Attr)
	Info(msg string, attrs ...Attr)
	Warn(msg string, attrs ...Attr)
	Error(msg string, attrs ...Attr)
	Fatal(msg string, err error)
	With(attrs ...Attr) Logger
}
