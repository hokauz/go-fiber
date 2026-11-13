package metrics

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestMiddleware_CollectsByRoutePatternAndUnknownPath(t *testing.T) {
	col := NewCollector("http_api", nil)

	app := fiber.New()
	app.Use(Middleware(col))
	app.Get("/users/:id", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	reqOK := httptest.NewRequest("GET", "/users/42", nil)
	respOK, err := app.Test(reqOK)
	if err != nil {
		t.Fatalf("request /users/42 failed: %v", err)
	}
	if respOK.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", respOK.StatusCode, fiber.StatusOK)
	}

	reqNF := httptest.NewRequest("GET", "/not-found", nil)
	respNF, err := app.Test(reqNF)
	if err != nil {
		t.Fatalf("request /not-found failed: %v", err)
	}
	if respNF.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want %d", respNF.StatusCode, fiber.StatusNotFound)
	}

	mfs, err := col.Registry().Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	okLabels := map[string]string{
		"method": "GET",
		"path":   "/users/:id",
		"status": "200",
	}
	if v := findCounterValue(mfs, "http_api_http_requests_total", okLabels); v != 1 {
		t.Fatalf("ok counter = %v, want 1", v)
	}
	if c := findHistogramCount(mfs, "http_api_http_request_duration_seconds", okLabels); c != 1 {
		t.Fatalf("ok histogram count = %d, want 1", c)
	}

	notFoundLabels := map[string]string{
		"method": "GET",
		"path":   unknownPath,
		"status": "404",
	}
	if v := findCounterValue(mfs, "http_api_http_requests_total", notFoundLabels); v != 1 {
		t.Fatalf("not found counter = %v, want 1", v)
	}

	inFlight := findGaugeValue(mfs, "http_api_http_requests_in_flight")
	if inFlight != 0 {
		t.Fatalf("in-flight gauge = %v, want 0", inFlight)
	}
}
