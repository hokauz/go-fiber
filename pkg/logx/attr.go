package logx

import "time"

// Attr represents a structured log field.
type Attr struct {
	Key   string
	Value any
}

// String builds a string attribute.
func String(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

// Int builds an int attribute.
func Int(key string, value int) Attr {
	return Attr{Key: key, Value: value}
}

// Float64 builds a float64 attribute.
func Float64(key string, value float64) Attr {
	return Attr{Key: key, Value: value}
}

// Bool builds a bool attribute.
func Bool(key string, value bool) Attr {
	return Attr{Key: key, Value: value}
}

// Duration builds a duration attribute.
func Duration(key string, value time.Duration) Attr {
	return Attr{Key: key, Value: value}
}

// Any builds a generic attribute.
func Any(key string, value any) Attr {
	return Attr{Key: key, Value: value}
}

// Error builds the canonical error attribute.
func Error(err error) Attr {
	return Attr{Key: "error", Value: err}
}
