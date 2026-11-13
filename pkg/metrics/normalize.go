package metrics

import "strings"

const unknownPath = "unknown"

// NormalizePath normalizes metric path labels and avoids empty values.
func NormalizePath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return unknownPath
	}
	return normalized
}
