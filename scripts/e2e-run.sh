#!/usr/bin/env bash
# e2e-run.sh: orchestrator for Go e2e tests with local JWKS auth.
# Flow: copy -> generate jwks/token -> api -> go test

set -euo pipefail

SOURCE_DIR="${SOURCE_DIR:-/source}"
WORKSPACE_DIR="${WORKSPACE_DIR:-/workspace/api}"
API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:8080}"
REPORT_DIR="${REPORT_DIR:-$WORKSPACE_DIR/e2e-report}"
AUTH_DIR="${AUTH_DIR:-/tmp/e2e-auth}"
AUTH_HOST="${AUTH_HOST:-127.0.0.1}"
AUTH_PORT="${AUTH_PORT:-9010}"
E2E_VERBOSE="${E2E_VERBOSE:-}"
DB_URL="${DB_URL:-}"
DB_WAIT_SECONDS="${DB_WAIT_SECONDS:-30}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$DB_URL" ]; then
  echo "E2E: DB_URL is required"
  exit 1
fi

# 1. PREP: copy project to workspace
SOURCE_DIR="$SOURCE_DIR" WORKSPACE_DIR="$WORKSPACE_DIR" "$SCRIPT_DIR/e2e-copy-project.sh"
cd "$WORKSPACE_DIR"

# 2. AUTH: generate local JWKS and JWT token
OUT_DIR="$AUTH_DIR" "$SCRIPT_DIR/e2e-generate-jwks-token.sh"
source "$AUTH_DIR/token.env"
export E2E_JWT_TOKEN E2E_EXPECTED_SUB

python3 -m http.server "$AUTH_PORT" --bind "$AUTH_HOST" --directory "$AUTH_DIR" > jwks.log 2>&1 &
JWKS_PID=$!

wait_for_db() {
  for i in $(seq 1 "$DB_WAIT_SECONDS"); do
    if psql "$DB_URL" -c 'SELECT 1' >/dev/null 2>&1; then
      return 0
    fi
    if [ "$i" -eq "$DB_WAIT_SECONDS" ]; then
      echo "E2E: DB failed to become ready"
      return 1
    fi
    sleep 1
  done
}

wait_for_db
psql "$DB_URL" -v ON_ERROR_STOP=1 -c 'DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;'
DB_URL="$DB_URL" MIGRATIONS_DIR="$WORKSPACE_DIR/migrations" "$SCRIPT_DIR/e2e-apply-migrations.sh"

# 3. API: start in background, wait for health
APP_ENV=development \
AUTH_ISSUER= \
AUTH_AUDIENCE= \
AUTH_ALLOWED_ALGS=RS256 \
AUTH_JWKS_URL="http://${AUTH_HOST}:${AUTH_PORT}/jwks.json" \
DB_URL="$DB_URL" \
PORT=8080 \
go run ./cmd/api > api.log 2>&1 &
API_PID=$!

cleanup() {
  kill "$API_PID" 2>/dev/null || true
  kill "$JWKS_PID" 2>/dev/null || true
}
trap cleanup EXIT

for i in $(seq 1 30); do
  if curl -sf "$API_BASE_URL/api/livez" > /dev/null 2>&1; then
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "E2E: API failed to start"
    cat api.log
    exit 1
  fi
  sleep 1
done

# 4. E2E: go test only packages with test files
echo "E2E: running tests..."
verbose=""
[ -n "$E2E_VERBOSE" ] && verbose="-v"
pkgs=$(go list -tags=e2e -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./test/e2e/... 2>/dev/null | tr '\n' ' ')
if [ -z "$pkgs" ]; then
  pkgs="./test/e2e/..."
fi
if ! API_BASE_URL="$API_BASE_URL" E2E_JWT_TOKEN="$E2E_JWT_TOKEN" E2E_EXPECTED_SUB="$E2E_EXPECTED_SUB" E2E_EXPECT_DB_READY=true go test -count=1 -tags=e2e $verbose $pkgs; then
  exit 1
fi
echo "E2E: all tests passed"

# 5. REPORT
mkdir -p "$REPORT_DIR"
cp -f api.log "$REPORT_DIR/" 2>/dev/null || true
cp -f jwks.log "$REPORT_DIR/" 2>/dev/null || true
