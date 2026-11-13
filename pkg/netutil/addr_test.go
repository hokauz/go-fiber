package netutil

import "testing"

func TestNormalizeListenAddr(t *testing.T) {
	tests := []struct {
		raw          string
		fallbackPort string
		want         string
	}{
		{"", "8080", ":8080"},
		{"  ", "8080", ":8080"},
		{":3000", "8080", ":3000"},
		{"3000", "8080", ":3000"},
		{"0.0.0.0:9090", "8080", "0.0.0.0:9090"},
		{"127.0.0.1:8080", "9090", "127.0.0.1:8080"},
		{" 4000 ", "8080", ":4000"},
	}
	for _, tt := range tests {
		got := NormalizeListenAddr(tt.raw, tt.fallbackPort)
		if got != tt.want {
			t.Errorf("NormalizeListenAddr(%q, %q) = %q, want %q", tt.raw, tt.fallbackPort, got, tt.want)
		}
	}
}
