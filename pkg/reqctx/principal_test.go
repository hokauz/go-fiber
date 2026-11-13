package reqctx

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"go-fiber/pkg/auth"
)

func TestSetPrincipalAndPrincipalFrom(t *testing.T) {
	app := fiber.New()
	app.Get("/private", func(c fiber.Ctx) error {
		SetPrincipal(c, auth.Principal{UserID: "user-1"})

		p, ok := PrincipalFrom(c)
		if !ok {
			t.Fatal("expected principal in context")
		}
		if p.UserID != "user-1" {
			t.Fatalf("user_id = %q, want %q", p.UserID, "user-1")
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestPrincipalFrom_Missing(t *testing.T) {
	app := fiber.New()
	app.Get("/private", func(c fiber.Ctx) error {
		if _, ok := PrincipalFrom(c); ok {
			t.Fatal("expected no principal")
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestPrincipalFrom_InvalidType(t *testing.T) {
	app := fiber.New()
	app.Get("/private", func(c fiber.Ctx) error {
		c.Locals(principalKey, "not-a-principal")

		if _, ok := PrincipalFrom(c); ok {
			t.Fatal("expected type assertion failure")
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestPrincipalOrErr_WithPrincipal(t *testing.T) {
	app := fiber.New()
	app.Get("/private", func(c fiber.Ctx) error {
		SetPrincipal(c, auth.Principal{UserID: "user-1"})

		p, err := PrincipalOrErr(c)
		if err != nil {
			return err
		}
		if p.UserID != "user-1" {
			t.Fatalf("user_id = %q, want %q", p.UserID, "user-1")
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestPrincipalOrErr_MissingPrincipal(t *testing.T) {
	app := fiber.New()
	app.Get("/private", func(c fiber.Ctx) error {
		_, err := PrincipalOrErr(c)
		return err
	})

	req := httptest.NewRequest("GET", "/private", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}
