package golangjwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"go-fiber/pkg/auth"
)

func TestEngine_ParseHeader(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	token.Header["kid"] = "kid-1"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	engine := New()
	h, err := engine.ParseHeader(signed)
	if err != nil {
		t.Fatalf("parse header: %v", err)
	}
	if h.KID != "kid-1" {
		t.Fatalf("kid = %q, want kid-1", h.KID)
	}
	if h.Alg != "RS256" {
		t.Fatalf("alg = %q, want RS256", h.Alg)
	}
}

func TestEngine_Verify(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.MapClaims{
		"sub":   "user-1",
		"email": "u@example.com",
		"iss":   "issuer-1",
		"aud":   "api-1",
		"exp":   time.Now().Add(time.Minute).Unix(),
	})
	token.Header["kid"] = "kid-1"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	engine := New()
	claims, err := engine.Verify(signed, &privateKey.PublicKey, auth.VerifyConfig{
		AllowedAlgs: []string{"RS256"},
		Issuer:      "issuer-1",
		Audience:    "api-1",
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("subject = %q, want user-1", claims.Subject)
	}
	if claims.Email != "u@example.com" {
		t.Fatalf("email = %q, want u@example.com", claims.Email)
	}
}

func TestEngine_Verify_InvalidToken(t *testing.T) {
	engine := New()
	_, err := engine.Verify("invalid.jwt.token", nil, auth.VerifyConfig{})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestEngine_Verify_WrongAlg(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	engine := New()
	_, err = engine.Verify(signed, &privateKey.PublicKey, auth.VerifyConfig{
		AllowedAlgs: []string{"ES256"},
	})
	if err == nil {
		t.Fatal("expected error for disallowed algorithm")
	}
}

func TestEngine_ParseHeader_InvalidToken(t *testing.T) {
	engine := New()
	_, err := engine.ParseHeader("not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
