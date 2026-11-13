package auth

import (
	"context"
	"fmt"
	"strings"
)

// OIDCVerifier verifies OIDC-compatible JWTs using injected engine + key provider.
type OIDCVerifier struct {
	engine      TokenEngine
	keys        KeyProvider
	allowedAlgs []string
	allowedSet  map[string]struct{}
	issuer      string
	audience    string
}

// NewOIDCVerifier creates a framework-agnostic JWT verifier.
func NewOIDCVerifier(cfg OIDCConfig) (Verifier, error) {
	if cfg.Engine == nil {
		return nil, fmt.Errorf("%w: missing token engine", ErrVerifierConfig)
	}
	if cfg.Keys == nil {
		return nil, fmt.Errorf("%w: missing key provider", ErrVerifierConfig)
	}

	allowed := normalizeAlgs(cfg.AllowedAlgs)
	if len(allowed) == 0 {
		return nil, fmt.Errorf("%w: empty allowed algorithms", ErrVerifierConfig)
	}

	set := make(map[string]struct{}, len(allowed))
	for _, alg := range allowed {
		set[alg] = struct{}{}
	}

	return &OIDCVerifier{
		engine:      cfg.Engine,
		keys:        cfg.Keys,
		allowedAlgs: allowed,
		allowedSet:  set,
		issuer:      strings.TrimSpace(cfg.Issuer),
		audience:    strings.TrimSpace(cfg.Audience),
	}, nil
}

// Verify validates the token header, resolves the signing key, and verifies claims.
func (v *OIDCVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("%w: empty token", ErrInvalidToken)
	}

	header, err := v.engine.ParseHeader(token)
	if err != nil {
		return nil, fmt.Errorf("%w: parse header: %v", ErrInvalidToken, err)
	}
	if header.KID == "" {
		return nil, fmt.Errorf("%w: missing kid", ErrInvalidToken)
	}
	if !v.isAllowedAlg(header.Alg) {
		return nil, fmt.Errorf("%w: unsupported alg %q", ErrInvalidToken, header.Alg)
	}

	key, err := v.keys.PublicKey(ctx, header.KID)
	if err != nil {
		return nil, classifyPublicKeyError(err)
	}

	claims, err := v.engine.Verify(token, key, VerifyConfig{
		AllowedAlgs: v.allowedAlgs,
		Issuer:      v.issuer,
		Audience:    v.audience,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: verify token: %v", ErrInvalidToken, err)
	}
	if claims == nil || strings.TrimSpace(claims.Subject) == "" {
		return nil, fmt.Errorf("%w: missing subject", ErrInvalidToken)
	}
	if claims.Raw == nil {
		claims.Raw = map[string]any{}
	}
	return claims, nil
}

// Warmup preloads the key provider cache.
func (v *OIDCVerifier) Warmup(ctx context.Context) error {
	if err := v.keys.Warmup(ctx); err != nil {
		return fmt.Errorf("%w: warmup: %v", ErrKeyProvider, err)
	}
	return nil
}

func (v *OIDCVerifier) isAllowedAlg(alg string) bool {
	_, ok := v.allowedSet[strings.ToUpper(strings.TrimSpace(alg))]
	return ok
}
