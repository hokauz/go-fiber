package jwks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"go-fiber/pkg/auth"
)

func TestNewProvider_RequiresJWKSURL(t *testing.T) {
	_, err := NewProvider(Config{})
	if !errors.Is(err, auth.ErrVerifierConfig) {
		t.Fatalf("expected ErrVerifierConfig, got %v", err)
	}
}

func TestNewProvider_ResolvesURL(t *testing.T) {
	url := "https://custom.example.com/.well-known/jwks.json"
	p, err := NewProvider(Config{JWKSURL: url})
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if p.jwksURL != url {
		t.Errorf("jwksURL = %q, want %q", p.jwksURL, url)
	}
}

func TestProvider_PublicKey_EmptyKid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	defer srv.Close()

	p, err := NewProvider(Config{
		JWKSURL:     srv.URL,
		CacheTTL:    time.Minute,
		HTTPTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	_, err = p.PublicKey(context.Background(), "")
	if !errors.Is(err, auth.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound for empty kid, got %v", err)
	}
}

func TestProvider_PublicKey_KidNotFound(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	jwkKey, err := jwk.FromRaw(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("create jwk: %v", err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, "kid-1"); err != nil {
		t.Fatalf("set kid: %v", err)
	}
	set := jwk.NewSet()
	set.AddKey(jwkKey)
	payload, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	p, err := NewProvider(Config{
		JWKSURL:     srv.URL,
		CacheTTL:    time.Minute,
		HTTPTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	if err := p.Warmup(context.Background()); err != nil {
		t.Fatalf("warmup: %v", err)
	}

	_, err = p.PublicKey(context.Background(), "nonexistent-kid")
	if !errors.Is(err, auth.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound for unknown kid, got %v", err)
	}
}

func TestProvider_PublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	jwkKey, err := jwk.FromRaw(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("create jwk: %v", err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, "kid-1"); err != nil {
		t.Fatalf("set kid: %v", err)
	}
	set := jwk.NewSet()
	set.AddKey(jwkKey)
	payload, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	p, err := NewProvider(Config{
		JWKSURL:     srv.URL,
		CacheTTL:    time.Minute,
		HTTPTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	if err := p.Warmup(context.Background()); err != nil {
		t.Fatalf("warmup: %v", err)
	}

	raw, err := p.PublicKey(context.Background(), "kid-1")
	if err != nil {
		t.Fatalf("public key: %v", err)
	}
	if _, ok := raw.(*rsa.PublicKey); !ok {
		t.Fatalf("expected *rsa.PublicKey, got %T", raw)
	}
}
