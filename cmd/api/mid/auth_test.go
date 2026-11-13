package mid

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"go-fiber/pkg/auth"
	"go-fiber/pkg/reqctx"
)

type warmupProbeVerifier struct {
	delay time.Duration
	err   error
	done  chan struct{}
}

func (f *warmupProbeVerifier) Verify(context.Context, string) (*auth.Claims, error) {
	return &auth.Claims{Subject: "user-1"}, nil
}

func (f *warmupProbeVerifier) Warmup(ctx context.Context) error {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if f.done != nil {
		close(f.done)
	}
	return f.err
}

func TestAuthWithOptions_NilEnvs(t *testing.T) {
	handler := AuthWithOptions(nil, DefaultAuthOptions())
	if handler == nil {
		t.Fatal("expected non-nil handler when envs is nil")
	}

	app := fiber.New()
	app.Get("/protected", handler, func(c fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500 when verifier not configured, got %d", resp.StatusCode)
	}
}

func TestStartWarmup_AsyncDoesNotBlock(t *testing.T) {
	done := make(chan struct{})
	verifier := &warmupProbeVerifier{
		delay: 150 * time.Millisecond,
		done:  done,
	}

	start := time.Now()
	startWarmup(verifier, AuthOptions{
		WarmupOnStart: true,
		WarmupAsync:   true,
		WarmupTimeout: time.Second,
		OnWarmupError: func(err error) {
			t.Fatalf("unexpected warmup error: %v", err)
		},
	})
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("startWarmup blocked for %v", elapsed)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("warmup did not complete")
	}
}

func TestStartWarmup_ErrorUsesCallback(t *testing.T) {
	wantErr := errors.New("warmup failed")
	verifier := &warmupProbeVerifier{err: wantErr}
	got := make(chan error, 1)

	startWarmup(verifier, AuthOptions{
		WarmupOnStart: true,
		WarmupAsync:   false,
		WarmupTimeout: time.Second,
		OnWarmupError: func(err error) {
			got <- err
		},
	})

	select {
	case err := <-got:
		if !errors.Is(err, wantErr) {
			t.Fatalf("callback error = %v, want %v", err, wantErr)
		}
	default:
		t.Fatal("expected warmup callback to be called")
	}
}

func TestRequireAuthFiber_MissingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequireAuthFiber(&warmupProbeVerifier{}))
	app.Get("/private", func(c fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
	body := decodeErrorBody(t, resp)
	if body["error"] != "auth/missing-header" {
		t.Fatalf("error = %q, want auth/missing-header", body["error"])
	}
}

func TestRequireAuthFiber_InvalidToken(t *testing.T) {
	app := fiber.New()
	app.Use(RequireAuthFiber(&fakeVerifier{err: auth.ErrInvalidToken}))
	app.Get("/private", func(c fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/private", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer invalid")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
	body := decodeErrorBody(t, resp)
	if body["error"] != "auth/invalid-token" {
		t.Fatalf("error = %q, want auth/invalid-token", body["error"])
	}
}

func TestRequireAuthFiber_SetsPrincipal(t *testing.T) {
	app := fiber.New()
	app.Use(RequireAuthFiber(&fakeVerifier{
		claims: &auth.Claims{Subject: "user-1", Email: "u@example.com", Raw: map[string]any{"role": "member"}},
	}))
	app.Get("/private", func(c fiber.Ctx) error {
		p, ok := reqctx.PrincipalFrom(c)
		if !ok || p.UserID != "user-1" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/private", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer valid")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

func TestRequireAuthFiber_NilVerifierReturnsInternalError(t *testing.T) {
	app := fiber.New()
	app.Use(RequireAuthFiber(nil))
	app.Get("/private", func(c fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusInternalServerError)
	}
	body := decodeErrorBody(t, resp)
	if body["error"] != "auth/internal-server-error" {
		t.Fatalf("error = %q, want auth/internal-server-error", body["error"])
	}
}

type fakeVerifier struct {
	claims *auth.Claims
	err    error
}

func (f *fakeVerifier) Verify(context.Context, string) (*auth.Claims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

func (f *fakeVerifier) Warmup(context.Context) error {
	return nil
}

func decodeErrorBody(t *testing.T, resp *http.Response) map[string]string {
	t.Helper()
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}
