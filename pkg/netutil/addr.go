package netutil

import "strings"

// NormalizeListenAddr ensures raw resolves to a valid host:port listen address.
// When raw is empty the fallbackPort is used; a bare port number gets a ":" prefix.
func NormalizeListenAddr(raw, fallbackPort string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ":" + fallbackPort
	}
	if strings.HasPrefix(trimmed, ":") {
		return trimmed
	}
	if strings.Contains(trimmed, ":") {
		return trimmed
	}
	return ":" + trimmed
}
