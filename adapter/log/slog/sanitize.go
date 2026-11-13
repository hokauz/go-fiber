package slog

import (
	"reflect"
	"strings"
)

const redactedValue = "[REDACTED]"

var sensitiveFragments = []string{
	"password",
	"passwd",
	"pwd",
	"token",
	"authorization",
	"secret",
	"apikey",
	"api_key",
	"accesskey",
	"access_key",
	"clientsecret",
	"client_secret",
}

func sanitizeAttr(key string, value any) any {
	if isSensitiveKey(key) {
		return redactedValue
	}
	return sanitizeValue(value)
}

func sanitizeValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, vv := range v {
			out[k] = sanitizeAttr(k, vv)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = sanitizeValue(v[i])
		}
		return out
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return value
	}

	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return value
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			out[k] = sanitizeAttr(k, iter.Value().Interface())
		}
		return out
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out[i] = sanitizeValue(rv.Index(i).Interface())
		}
		return out
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	k := normalizeKey(key)
	if k == "" {
		return false
	}
	for _, frag := range sensitiveFragments {
		if strings.Contains(k, normalizeKey(frag)) {
			return true
		}
	}
	return false
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, " ", "")
	return key
}
