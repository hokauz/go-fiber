package auth

import (
	"context"
)

// AuthError represents an authentication failure with a machine-readable code.
type AuthError struct {
	Code string
}

// Claims is the normalized claims payload returned by the verifier.
type Claims struct {
	Subject string
	Email   string
	Raw     map[string]any
}

// Header contains the JWT header fields needed by the verifier.
type Header struct {
	KID string
	Alg string
}

// VerifyConfig holds claim constraints used by a token engine.
type VerifyConfig struct {
	AllowedAlgs []string
	Issuer      string
	Audience    string
}

// Verifier validates a token and returns normalized claims.
type Verifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
	Warmup(ctx context.Context) error
}

// TokenEngine is the port for the JWT parsing/verifying implementation.
// External JWT libraries live behind adapters implementing this interface.
type TokenEngine interface {
	ParseHeader(token string) (Header, error)
	Verify(token string, key any, cfg VerifyConfig) (*Claims, error)
}

// KeyProvider is the port for resolving keys by token kid.
// External key/JWKS providers live behind adapters implementing this interface.
type KeyProvider interface {
	PublicKey(ctx context.Context, kid string) (any, error)
	Warmup(ctx context.Context) error
}

// Principal stores the normalized authenticated user data in request scope.
type Principal struct {
	UserID string
	Email  string
	Claims map[string]any
}

// OIDCConfig defines dependencies and validation constraints for OIDC JWT verification.
type OIDCConfig struct {
	Engine      TokenEngine
	Keys        KeyProvider
	AllowedAlgs []string
	Issuer      string
	Audience    string
}

// Auth error codes.
var (
	ErrAuthMissingHeader      = &AuthError{Code: "auth/missing-header"}
	ErrAuthInvalidScheme      = &AuthError{Code: "auth/invalid-scheme"}
	ErrAuthRequired           = &AuthError{Code: "auth/required"}
	ErrAuthInvalidToken       = &AuthError{Code: "auth/invalid-token"}
	ErrAuthServiceUnavailable = &AuthError{Code: "auth/service-unavailable"}
	ErrAuthInternalError      = &AuthError{Code: "auth/internal-server-error"}
	ErrVerifierConfig         = &AuthError{Code: "auth/verifier-config"}
	ErrInvalidToken           = ErrAuthInvalidToken // alias for verifier use
	ErrKeyNotFound            = &AuthError{Code: "auth/key-not-found"}
	ErrKeyProvider            = &AuthError{Code: "auth/key-provider"}
)

func (e *AuthError) Error() string {
	return e.Code
}
