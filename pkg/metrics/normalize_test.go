package metrics

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "pattern path", in: "/api/v1/hello/:id", want: "/api/v1/hello/:id"},
		{name: "trim spaces", in: "  /metrics ", want: "/metrics"},
		{name: "empty", in: "", want: unknownPath},
		{name: "spaces", in: "   ", want: unknownPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.in)
			if got != tt.want {
				t.Fatalf("NormalizePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
