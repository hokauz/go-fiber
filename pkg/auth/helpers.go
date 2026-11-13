package auth

import (
	"errors"
	"fmt"
	"strings"
)

// IsAuthError returns the AuthError if err is an AuthError.
func IsAuthError(err error) (*AuthError, bool) {
	var ae *AuthError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// normalizeAlgs normalizes and deduplicates algorithm names.
func normalizeAlgs(in []string) []string {
	if len(in) == 0 {
		return []string{"RS256"}
	}

	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, alg := range in {
		alg = strings.ToUpper(strings.TrimSpace(alg))
		if alg == "" {
			continue
		}
		if _, ok := seen[alg]; ok {
			continue
		}
		seen[alg] = struct{}{}
		out = append(out, alg)
	}
	return out
}

// classifyPublicKeyError maps key provider errors to the appropriate auth error.
func classifyPublicKeyError(err error) error {
	switch {
	case errors.Is(err, ErrKeyNotFound):
		return fmt.Errorf("%w: resolve key: %v", ErrInvalidToken, err)
	case errors.Is(err, ErrKeyProvider):
		return fmt.Errorf("%w: resolve key: %v", ErrKeyProvider, err)
	default:
		return fmt.Errorf("%w: resolve key: %v", ErrKeyProvider, err)
	}
}
