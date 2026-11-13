package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeTokenEngine struct {
	header Header
	claims *Claims
	err    error
}

func (f *fakeTokenEngine) ParseHeader(token string) (Header, error) {
	if token == "parse-error" {
		return Header{}, errors.New("parse error")
	}
	return f.header, nil
}

func (f *fakeTokenEngine) Verify(string, any, VerifyConfig) (*Claims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

type fakeKeyProvider struct {
	key    any
	err    error
	warmed bool
}

func (f *fakeKeyProvider) PublicKey(context.Context, string) (any, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.key, nil
}

func (f *fakeKeyProvider) Warmup(context.Context) error {
	f.warmed = true
	return nil
}

func TestNewOIDCVerifier_ConfigValidation(t *testing.T) {
	_, err := NewOIDCVerifier(OIDCConfig{})
	if !errors.Is(err, ErrVerifierConfig) {
		t.Fatalf("expected ErrVerifierConfig, got %v", err)
	}
}

func TestOIDCVerifier_Verify_RejectsInvalidAlg(t *testing.T) {
	v, err := NewOIDCVerifier(OIDCConfig{
		Engine:      &fakeTokenEngine{header: Header{KID: "k1", Alg: "HS256"}},
		Keys:        &fakeKeyProvider{key: "key"},
		AllowedAlgs: []string{"RS256"},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	if _, err := v.Verify(context.Background(), "token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestOIDCVerifier_Verify_KeyNotFoundMappedToInvalidToken(t *testing.T) {
	v, err := NewOIDCVerifier(OIDCConfig{
		Engine:      &fakeTokenEngine{header: Header{KID: "k1", Alg: "RS256"}},
		Keys:        &fakeKeyProvider{err: ErrKeyNotFound},
		AllowedAlgs: []string{"RS256"},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	if _, err := v.Verify(context.Background(), "token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestOIDCVerifier_Verify_KeyProviderErrorMappedToKeyProvider(t *testing.T) {
	v, err := NewOIDCVerifier(OIDCConfig{
		Engine:      &fakeTokenEngine{header: Header{KID: "k1", Alg: "RS256"}},
		Keys:        &fakeKeyProvider{err: ErrKeyProvider},
		AllowedAlgs: []string{"RS256"},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	if _, err := v.Verify(context.Background(), "token"); !errors.Is(err, ErrKeyProvider) {
		t.Fatalf("expected ErrKeyProvider, got %v", err)
	}
}

func TestOIDCVerifier_Verify_Success(t *testing.T) {
	v, err := NewOIDCVerifier(OIDCConfig{
		Engine: &fakeTokenEngine{
			header: Header{KID: "k1", Alg: "RS256"},
			claims: &Claims{Subject: "user-1", Email: "u@example.com"},
		},
		Keys:        &fakeKeyProvider{key: "key"},
		AllowedAlgs: []string{"RS256"},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	claims, err := v.Verify(context.Background(), "token")
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("subject = %q, want user-1", claims.Subject)
	}
	if claims.Raw == nil {
		t.Fatal("raw claims should be initialized")
	}
}

func TestOIDCVerifier_Warmup_DelegatesToProvider(t *testing.T) {
	provider := &fakeKeyProvider{key: "key"}
	v, err := NewOIDCVerifier(OIDCConfig{
		Engine:      &fakeTokenEngine{header: Header{KID: "k1", Alg: "RS256"}},
		Keys:        provider,
		AllowedAlgs: []string{"RS256"},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	if err := v.Warmup(context.Background()); err != nil {
		t.Fatalf("warmup: %v", err)
	}
	if !provider.warmed {
		t.Fatal("expected provider to be warmed")
	}
}
