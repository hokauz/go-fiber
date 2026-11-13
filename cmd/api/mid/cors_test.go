package mid

import (
	"net/http/httptest"
	"reflect"
	"testing"

	apienv "go-fiber/cmd/api/env"

	"github.com/gofiber/fiber/v3"
)

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name string
		cfg  *apienv.CORSEnvs
		want []string
	}{
		{
			name: "nil cfg falls back to wildcard",
			cfg:  nil,
			want: []string{"*"},
		},
		{
			name: "empty value falls back to wildcard",
			cfg:  &apienv.CORSEnvs{Origins: ""},
			want: []string{"*"},
		},
		{
			name: "only delimiters falls back to wildcard",
			cfg:  &apienv.CORSEnvs{Origins: " ; ; "},
			want: []string{"*"},
		},
		{
			name: "specific origins split by semicolon",
			cfg:  &apienv.CORSEnvs{Origins: "https://app.example.com; https://admin.example.com"},
			want: []string{"https://app.example.com", "https://admin.example.com"},
		},
		{
			name: "wildcard mixed with specific resolves to wildcard only",
			cfg:  &apienv.CORSEnvs{Origins: "*; https://app.example.com"},
			want: []string{"*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAllowedOrigins(tt.cfg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseAllowedOrigins() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCORS_WildcardWithSpecific_DoesNotEnableCredentials(t *testing.T) {
	app := fiber.New()
	app.Use(CORS(&apienv.CORSEnvs{Origins: "*;https://app.example.com"}))
	app.Get("/", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://random.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty", got)
	}
}
