package golangjwt

import (
	"fmt"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"go-fiber/pkg/auth"
)

// JWT header keys.
const (
	headerKID  = "kid"
	headerAlg  = "alg"
	claimEmail = "email"
)

// Engine adapts github.com/golang-jwt/jwt/v5 to the core TokenEngine port.
type Engine struct{}

// New creates a JWT engine adapter.
func New() *Engine {
	return &Engine{}
}

// ParseHeader parses the JWT header without claims verification.
func (e *Engine) ParseHeader(token string) (auth.Header, error) {
	parsed, _, err := new(jwtv5.Parser).ParseUnverified(token, jwtv5.MapClaims{})
	if err != nil {
		return auth.Header{}, fmt.Errorf("%w: parse header: %v", auth.ErrAuthInvalidToken, err)
	}

	return auth.Header{
		KID: getString(parsed.Header, headerKID),
		Alg: getString(parsed.Header, headerAlg),
	}, nil
}

// Verify validates signature and standard constraints and returns normalized claims.
func (e *Engine) Verify(token string, key any, cfg auth.VerifyConfig) (*auth.Claims, error) {
	opts := buildParserOptions(cfg)
	mapClaims := jwtv5.MapClaims{}

	parsed, err := jwtv5.ParseWithClaims(token, mapClaims, keyFunc(key), opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: verify: %v", auth.ErrAuthInvalidToken, err)
	}

	if !parsed.Valid {
		return nil, fmt.Errorf("%w: token invalid", auth.ErrAuthInvalidToken)
	}

	claims, err := extractClaims(mapClaims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// buildParserOptions builds parser options from VerifyConfig.
func buildParserOptions(cfg auth.VerifyConfig) []jwtv5.ParserOption {
	var opts []jwtv5.ParserOption

	if len(cfg.AllowedAlgs) > 0 {
		opts = append(opts, jwtv5.WithValidMethods(cfg.AllowedAlgs))
	}
	if cfg.Issuer != "" {
		opts = append(opts, jwtv5.WithIssuer(cfg.Issuer))
	}
	if cfg.Audience != "" {
		opts = append(opts, jwtv5.WithAudience(cfg.Audience))
	}
	opts = append(opts, jwtv5.WithExpirationRequired())

	return opts
}

// keyFunc returns a jwt.Keyfunc that always uses the given key.
func keyFunc(key any) jwtv5.Keyfunc {
	return func(_ *jwtv5.Token) (any, error) {
		return key, nil
	}
}

// extractClaims builds auth.Claims from MapClaims.
func extractClaims(mapClaims jwtv5.MapClaims) (*auth.Claims, error) {
	sub, err := mapClaims.GetSubject()
	if err != nil || sub == "" {
		return nil, fmt.Errorf("%w: missing subject", auth.ErrAuthInvalidToken)
	}

	return &auth.Claims{
		Subject: sub,
		Email:   getString(mapClaims, claimEmail),
		Raw:     copyMap(mapClaims),
	}, nil
}

// getString safely extracts a string from a map.
func getString(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

// copyMap returns a shallow copy of the map.
func copyMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
