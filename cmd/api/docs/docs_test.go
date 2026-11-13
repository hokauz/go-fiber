package docs

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestScalarDocsPage_Success(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")
	spec := `{"openapi":"3.0.3","info":{"title":"Test API","version":"1.0.0"},"paths":{}}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec file: %v", err)
	}

	originalPath := openAPISpecPath
	openAPISpecPath = specPath
	t.Cleanup(func() { openAPISpecPath = originalPath })

	app := fiber.New()
	PublicRegister(false)(app.Group("/api"))

	req := httptest.NewRequest("GET", "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestScalarDocsPage_MissingSpec(t *testing.T) {
	originalPath := openAPISpecPath
	openAPISpecPath = filepath.Join(t.TempDir(), "missing-openapi.json")
	t.Cleanup(func() { openAPISpecPath = originalPath })

	app := fiber.New()
	PublicRegister(false)(app.Group("/api"))

	req := httptest.NewRequest("GET", "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestPublicRegister_SkipsInProduction(t *testing.T) {
	app := fiber.New()
	PublicRegister(true)(app.Group("/api"))

	req := httptest.NewRequest("GET", "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 when docs disabled, got %d", resp.StatusCode)
	}
}
