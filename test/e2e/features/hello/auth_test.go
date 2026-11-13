//go:build e2e

package hello

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"strconv"
	"testing"
	"time"

	"go-fiber/test/e2e"
	"go-fiber/test/e2e/internal/helpers"
	"go-fiber/test/e2e/internal/steps"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func TestHelloMe_RejectsWithoutToken(t *testing.T) {
	t.Run("When_get_me_without_auth", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenGetHelloMeNoAuth(t, ctx)
		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_missing_header", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/missing-header")
		})
	})
}

func TestHelloMe_RejectsInvalidToken(t *testing.T) {
	t.Run("When_get_me_with_invalid_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenGetHelloMe(t, ctx, "invalid.jwt.token")
		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_invalid_token", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/invalid-token")
		})
	})
}

func TestHelloMe_RejectsTokenSignedByUnknownKey(t *testing.T) {
	token := mustSignTokenWithUnknownKey(t)
	t.Run("When_get_me_with_foreign_signed_token", func(t *testing.T) {
		ctx := e2e.NewTestContext(t)
		resp := steps.WhenGetHelloMe(t, ctx, token)
		t.Run("Then_unauthorized", func(t *testing.T) {
			helpers.ThenStatusIs(t, resp, http.StatusUnauthorized)
		})
		t.Run("And_error_invalid_token", func(t *testing.T) {
			steps.AndErrorCodeIs(t, resp, "auth/invalid-token")
		})
	})
}

func mustSignTokenWithUnknownKey(t *testing.T) string {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	now := time.Now().Unix()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.MapClaims{
		"sub":   "e2e-foreign-user",
		"email": "e2e-foreign@local.test",
		"iat":   now,
		"exp":   now + 600,
	})
	token.Header["kid"] = "e2e-fast-like-kid-" + strconv.FormatInt(now, 10)
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}
	return signed
}
