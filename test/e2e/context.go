//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

// TestContext holds shared state between e2e steps.
type TestContext struct {
	t       *testing.T
	client  *http.Client
	baseURL string
}

// NewTestContext creates a TestContext with HTTP client and base URL from API_BASE_URL.
func NewTestContext(t *testing.T) *TestContext {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}
	return &TestContext{
		t:       t,
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

// Get performs a GET request to path (relative to baseURL).
func (c *TestContext) Get(path string) *http.Response {
	url := c.baseURL + path
	resp, err := c.client.Get(url)
	if err != nil {
		c.t.Helper()
		c.t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

// JSONRequest performs an HTTP request with optional JSON payload and headers.
func (c *TestContext) JSONRequest(method, path string, payload any, headers map[string]string) *http.Response {
	c.t.Helper()

	url := c.baseURL + path
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		if err != nil {
			c.t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		c.t.Fatalf("new request %s %s: %v", method, url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("do request %s %s: %v", method, url, err)
	}
	return resp
}
