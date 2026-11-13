package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeVerifier struct {
	claims *Claims
	err    error
}

func (f *fakeVerifier) Verify(context.Context, string) (*Claims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

func (f *fakeVerifier) Warmup(context.Context) error {
	return nil
}

func TestValidateAuth_MissingHeader(t *testing.T) {
	_, err := ValidateAuth(context.Background(), &fakeVerifier{}, "")
	if err == nil {
		t.Fatal("expected error for missing header")
	}
	if !errors.Is(err, ErrAuthMissingHeader) {
		t.Fatalf("err = %v, want ErrAuthMissingHeader", err)
	}
}

func TestValidateAuth_InvalidScheme(t *testing.T) {
	_, err := ValidateAuth(context.Background(), &fakeVerifier{}, "Basic xxx")
	if err == nil {
		t.Fatal("expected error for invalid scheme")
	}
	if !errors.Is(err, ErrAuthInvalidScheme) {
		t.Fatalf("err = %v, want ErrAuthInvalidScheme", err)
	}
}

func TestValidateAuth_InvalidToken(t *testing.T) {
	_, err := ValidateAuth(context.Background(), &fakeVerifier{err: ErrInvalidToken}, "Bearer invalid")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if !errors.Is(err, ErrAuthInvalidToken) {
		t.Fatalf("err = %v, want ErrAuthInvalidToken", err)
	}
}

func TestValidateAuth_KeyProviderUnavailable(t *testing.T) {
	_, err := ValidateAuth(context.Background(), &fakeVerifier{err: ErrKeyProvider}, "Bearer valid")
	if err == nil {
		t.Fatal("expected error for key provider unavailable")
	}
	if !errors.Is(err, ErrAuthServiceUnavailable) {
		t.Fatalf("err = %v, want ErrAuthServiceUnavailable", err)
	}
}

func TestValidateAuth_NilVerifier(t *testing.T) {
	_, err := ValidateAuth(context.Background(), nil, "Bearer token")
	if err == nil {
		t.Fatal("expected error for nil verifier")
	}
	if !errors.Is(err, ErrAuthInternalError) {
		t.Fatalf("err = %v, want ErrAuthInternalError", err)
	}
}

func TestValidateAuth_Success(t *testing.T) {
	principal, err := ValidateAuth(context.Background(), &fakeVerifier{
		claims: &Claims{Subject: "user-1", Email: "u@example.com", Raw: map[string]any{"role": "member"}},
	}, "Bearer valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if principal.UserID != "user-1" {
		t.Fatalf("UserID = %q, want %q", principal.UserID, "user-1")
	}
	if principal.Email != "u@example.com" {
		t.Fatalf("Email = %q, want %q", principal.Email, "u@example.com")
	}
}

func TestIsAuthError(t *testing.T) {
	if _, ok := IsAuthError(ErrAuthMissingHeader); !ok {
		t.Error("IsAuthError(ErrAuthMissingHeader) = false, want true")
	}
	if _, ok := IsAuthError(nil); ok {
		t.Error("IsAuthError(nil) = true, want false")
	}
	if _, ok := IsAuthError(ErrInvalidToken); !ok {
		t.Error("IsAuthError(ErrInvalidToken) = false, want true (ErrInvalidToken is AuthError)")
	}
}
