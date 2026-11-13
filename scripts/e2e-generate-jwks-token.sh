#!/usr/bin/env bash
# e2e-generate-jwks-token.sh: generates local JWKS + signed JWT for e2e auth.

set -euo pipefail

OUT_DIR="${OUT_DIR:-/tmp/e2e-auth}"
SUBJECT="${SUBJECT:-user-e2e-123}"
EMAIL="${EMAIL:-e2e@local.test}"
KID="${KID:-e2e-kid-1}"
EXPIRES_SECONDS="${EXPIRES_SECONDS:-600}"

mkdir -p "$OUT_DIR"

GEN_FILE="/tmp/e2e-jwks-gen.go"
cat > "$GEN_FILE" <<'EOF'
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func getenvDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func main() {
	outDir := getenvDefault("OUT_DIR", "/tmp/e2e-auth")
	subject := getenvDefault("SUBJECT", "user-e2e-123")
	email := getenvDefault("EMAIL", "e2e@local.test")
	kid := getenvDefault("KID", "e2e-kid-1")

	expSecondsStr := getenvDefault("EXPIRES_SECONDS", "600")
	expSeconds, err := strconv.Atoi(expSecondsStr)
	if err != nil || expSeconds <= 0 {
		panic(fmt.Errorf("invalid EXPIRES_SECONDS: %q", expSecondsStr))
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.FromRaw(&key.PublicKey)
	if err != nil {
		panic(err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, kid); err != nil {
		panic(err)
	}
	if err := jwkKey.Set(jwk.AlgorithmKey, "RS256"); err != nil {
		panic(err)
	}
	if err := jwkKey.Set(jwk.KeyUsageKey, "sig"); err != nil {
		panic(err)
	}

	set := jwk.NewSet()
	set.AddKey(jwkKey)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}

	jwksBytes, err := json.Marshal(set)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(outDir+"/jwks.json", jwksBytes, 0o644); err != nil {
		panic(err)
	}

	now := time.Now()
	claims := jwtv5.MapClaims{
		"sub":   subject,
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(expSeconds) * time.Second).Unix(),
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(outDir+"/token.txt", []byte(signed), 0o600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(outDir+"/token.env", []byte("E2E_JWT_TOKEN="+signed+"\nE2E_EXPECTED_SUB="+subject+"\n"), 0o600); err != nil {
		panic(err)
	}
}
EOF

OUT_DIR="$OUT_DIR" SUBJECT="$SUBJECT" EMAIL="$EMAIL" KID="$KID" EXPIRES_SECONDS="$EXPIRES_SECONDS" go run "$GEN_FILE"

