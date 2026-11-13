//go:build e2e

package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// AuthHeaders returns standard bearer auth headers.
func AuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// ThenStatusIs asserts the response status code.
func ThenStatusIs(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("status = %d, want %d, body=%s", resp.StatusCode, want, string(body))
	}
}

// MustDecodeJSON decodes response JSON body into out and closes body.
func MustDecodeJSON(t *testing.T, resp *http.Response, out any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("decode body: %v raw=%s", err, string(raw))
	}
}

// MustParseJSON decodes resp.Body as JSON into a map and closes the body.
// Use when multiple assertions need the same parsed data.
func MustParseJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("decode JSON: %v raw=%s", err, string(raw))
	}
	return m
}

// AndJSONFieldEquals asserts a JSON field has the expected string value.
// Reads and closes resp.Body.
func AndJSONFieldEquals(t *testing.T, resp *http.Response, field, want string) {
	t.Helper()
	m := MustParseJSON(t, resp)
	got, ok := m[field]
	if !ok {
		t.Fatalf("field %q not found in response", field)
	}
	gotStr, ok := got.(string)
	if !ok {
		t.Fatalf("field %q is not a string: %T", field, got)
	}
	if gotStr != want {
		t.Fatalf("field %q: got %q, want %q", field, gotStr, want)
	}
}

// AndErrorCodeIs asserts the error envelope has the expected code.
// go-fiber uses {"error": "auth/missing-header"} format.
func AndErrorCodeIs(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	m := MustParseJSON(t, resp)
	got, _ := m["error"].(string)
	if got != want {
		t.Fatalf("error: got %q want %q", got, want)
	}
}
