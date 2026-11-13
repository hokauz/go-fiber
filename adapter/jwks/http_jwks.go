package jwks

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"go-fiber/pkg/auth"
)

const (
	defaultCacheTTL    = 10 * time.Minute
	defaultHTTPTimeout = 3 * time.Second
)

// Config holds settings for the HTTP JWKS provider.
type Config struct {
	// JWKSURL is the full URL to the JWKS JSON document (e.g. https://example.com/.well-known/jwks.json).
	// Required.
	JWKSURL     string
	CacheTTL    time.Duration
	HTTPTimeout time.Duration
}

// Provider fetches and caches public keys from a remote JWKS endpoint.
// It implements auth.KeyProvider and is agnostic to the identity provider.
type Provider struct {
	jwksURL string
	client  *http.Client
	ttl     time.Duration

	mu       sync.RWMutex
	cacheSet jwk.Set
	cacheAt  time.Time
}

// NewProvider creates a JWKS provider that fetches keys from the given URL.
// JWKSURL must be non-empty.
func NewProvider(cfg Config) (*Provider, error) {
	url := strings.TrimSpace(cfg.JWKSURL)
	if url == "" {
		return nil, fmt.Errorf("%w: JWKSURL is required", auth.ErrVerifierConfig)
	}

	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}

	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}

	return &Provider{
		jwksURL: url,
		client: &http.Client{
			Timeout: timeout,
		},
		ttl: ttl,
	}, nil
}

// Warmup preloads the JWKS cache.
func (p *Provider) Warmup(ctx context.Context) error {
	return p.refresh(ctx)
}

// PublicKey returns the key for a given kid, refreshing cache on miss.
func (p *Provider) PublicKey(ctx context.Context, kid string) (any, error) {
	kid = strings.TrimSpace(kid)
	if kid == "" {
		return nil, fmt.Errorf("%w: empty kid", auth.ErrKeyNotFound)
	}

	if err := p.ensureFresh(ctx); err != nil {
		return nil, err
	}

	key, ok := p.lookup(kid)
	if !ok {
		if err := p.refresh(ctx); err != nil {
			return nil, err
		}
		key, ok = p.lookup(kid)
		if !ok {
			return nil, fmt.Errorf("%w: kid %q not found", auth.ErrKeyNotFound, kid)
		}
	}

	raw, err := keyToCrypto(key)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *Provider) ensureFresh(ctx context.Context) error {
	p.mu.RLock()
	set := p.cacheSet
	cacheAt := p.cacheAt
	ttl := p.ttl
	p.mu.RUnlock()

	if set != nil && time.Since(cacheAt) <= ttl {
		return nil
	}
	return p.refresh(ctx)
}

func (p *Provider) lookup(kid string) (jwk.Key, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.cacheSet == nil {
		return nil, false
	}
	return p.cacheSet.LookupKeyID(kid)
}

func (p *Provider) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("%w: create request: %v", auth.ErrKeyProvider, err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: fetch jwks: %v", auth.ErrKeyProvider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: fetch jwks status %d", auth.ErrKeyProvider, resp.StatusCode)
	}

	set, err := jwk.ParseReader(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: parse jwks: %v", auth.ErrKeyProvider, err)
	}

	p.mu.Lock()
	p.cacheSet = set
	p.cacheAt = time.Now()
	p.mu.Unlock()
	return nil
}

func keyToCrypto(key jwk.Key) (any, error) {
	var raw any
	if err := key.Raw(&raw); err != nil {
		return nil, fmt.Errorf("%w: extract raw key: %v", auth.ErrKeyProvider, err)
	}

	switch raw.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
		return raw, nil
	default:
		return nil, fmt.Errorf("%w: unsupported key type %T", auth.ErrKeyProvider, raw)
	}
}
