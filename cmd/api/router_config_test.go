package main

import (
	"testing"
	"time"

	apienv "go-fiber/cmd/api/env"
)

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty uses default", in: "", want: "/api"},
		{name: "trim space", in: "  /v1/api  ", want: "/v1/api"},
		{name: "adds leading slash", in: "v1/api", want: "/v1/api"},
		{name: "cleans path", in: "/v1//api/", want: "/v1/api"},
		{name: "root path", in: "/", want: "/"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeBasePath(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeBasePath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestAPIRouterConfig_DefaultTimeouts(t *testing.T) {
	cfg := apiRouterConfig(nil)

	if cfg.ReadTimeout != 5*time.Second {
		t.Fatalf("ReadTimeout = %s, want 5s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Fatalf("WriteTimeout = %s, want 15s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want 60s", cfg.IdleTimeout)
	}
}

func TestAPIRouterConfig_EnvTimeouts(t *testing.T) {
	cfg := apiRouterConfig(&apienv.ServerEnvs{
		ReadTimeoutSeconds:  7,
		WriteTimeoutSeconds: 17,
		IdleTimeoutSeconds:  70,
	})

	if cfg.ReadTimeout != 7*time.Second {
		t.Fatalf("ReadTimeout = %s, want 7s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 17*time.Second {
		t.Fatalf("WriteTimeout = %s, want 17s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 70*time.Second {
		t.Fatalf("IdleTimeout = %s, want 70s", cfg.IdleTimeout)
	}
}

func TestAPIBasePath_DefaultWhenEmpty(t *testing.T) {
	got := apiBasePath(&apienv.ServerEnvs{BasePath: ""})
	if got != "/api" {
		t.Fatalf("apiBasePath(empty) = %q, want /api", got)
	}
}
