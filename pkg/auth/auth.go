package auth

import (
	"context"
	"errors"
	"strings"
)

// ValidateAuth validates bearer JWT from the Authorization header.
// It is framework-agnostic and returns Principal or an AuthError.
func ValidateAuth(ctx context.Context, verifier Verifier, authHeader string) (Principal, error) {
	if verifier == nil {
		return Principal{}, ErrAuthInternalError
	}

	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return Principal{}, ErrAuthMissingHeader
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return Principal{}, ErrAuthInvalidScheme
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return Principal{}, ErrAuthInvalidToken
	}

	claims, err := verifier.Verify(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			return Principal{}, ErrAuthInvalidToken
		case errors.Is(err, ErrKeyProvider):
			return Principal{}, ErrAuthServiceUnavailable
		default:
			return Principal{}, ErrAuthInternalError
		}
	}

	return Principal{
		UserID: claims.Subject,
		Email:  claims.Email,
		Claims: claims.Raw,
	}, nil
}
