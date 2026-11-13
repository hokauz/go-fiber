#!/usr/bin/env bash

set -euo pipefail

SOURCE_DIR="${SOURCE_DIR:-/source}"
WORKSPACE_DIR="${WORKSPACE_DIR:-/workspace/api}"
REPORT_ROOT="${REPORT_ROOT:-/report}"
WORKTREE_ID="${WORKTREE_ID:-$(basename "$SOURCE_DIR")}"
RUN_ID="${RUN_ID:-$(date +%Y%m%d-%H%M%S)-$RANDOM}"
RUN_MODE="${RUN_MODE:-fast}"
SUPABASE_REUSE="${SUPABASE_REUSE:-1}"
E2E_TEST_EMAIL="${E2E_TEST_EMAIL:-e2e-user@local.test}"
E2E_TEST_PASSWORD="${E2E_TEST_PASSWORD:-Passw0rd!234}"
E2E_VERBOSE="${E2E_VERBOSE:-0}"
HEALTH_TIMEOUT_SECONDS="${HEALTH_TIMEOUT_SECONDS:-90}"
RUNNER_NETWORK_MODE="${RUNNER_NETWORK_MODE:-host}"
AUTH_ALLOWED_ALGS="${AUTH_ALLOWED_ALGS:-RS256,ES256}"
DB_WAIT_SECONDS="${DB_WAIT_SECONDS:-60}"

if [[ "$RUN_MODE" == "strict" ]]; then
  SUPABASE_REUSE="0"
fi

RUN_ROOT="/tmp/go-fiber-e2e-supa/runs/${WORKTREE_ID}/${RUN_ID}"
if [[ "$SUPABASE_REUSE" == "1" ]]; then
  SUPABASE_WORKDIR="/tmp/go-fiber-e2e-supa/shared/${WORKTREE_ID}/supabase"
  SUPABASE_STOP_ON_EXIT="0"
else
  SUPABASE_WORKDIR="${RUN_ROOT}/supabase"
  SUPABASE_STOP_ON_EXIT="1"
fi

SUPABASE_ENV_FILE="${RUN_ROOT}/supabase.env"
TOKEN_ENV_FILE="${RUN_ROOT}/token.env"
API_LOG_FILE="${RUN_ROOT}/api.log"
REPORT_DIR="${REPORT_ROOT}/${WORKTREE_ID}/${RUN_ID}"
LATEST_LINK_DIR="${REPORT_ROOT}/${WORKTREE_ID}"

API_PID=""

load_env_file() {
  local env_file="$1"
  set -a
  # shellcheck source=/dev/null
  source "$env_file"
  set +a
}

copy_reports() {
  mkdir -p "$REPORT_DIR"

  if [[ -f "$API_LOG_FILE" ]]; then
    cp "$API_LOG_FILE" "$REPORT_DIR/api.log"
  fi

  mkdir -p "$LATEST_LINK_DIR"
  (
    cd "$LATEST_LINK_DIR"
    ln -sfn "$RUN_ID" latest
  )
}

stop_api_process() {
  kill "$API_PID" >/dev/null 2>&1
}

stop_supabase_stack() {
  supabase stop --workdir "$SUPABASE_WORKDIR" --no-backup >/dev/null 2>&1
}

cleanup() {
  local exit_code=$?
  trap - EXIT INT TERM
  set +e

  if [[ -n "$API_PID" ]] && kill -0 "$API_PID" >/dev/null 2>&1; then
    stop_api_process || true
    wait "$API_PID" >/dev/null 2>&1 || true
  fi

  if [[ "$SUPABASE_STOP_ON_EXIT" == "1" ]] && [[ -d "$SUPABASE_WORKDIR" ]]; then
    stop_supabase_stack || true
  fi

  if [[ -d "$WORKSPACE_DIR" ]]; then
    rm -rf "$WORKSPACE_DIR"
  fi

  if [[ -d "$RUN_ROOT" ]]; then
    rm -rf "$RUN_ROOT"
  fi

  exit "$exit_code"
}

trap cleanup EXIT INT TERM

mkdir -p "$RUN_ROOT" "$WORKSPACE_DIR" "$REPORT_DIR"

echo "[e2e-supa] PREP: copy project"
/workspace/scripts/e2e-supa-copy-project.sh "$SOURCE_DIR" "$WORKSPACE_DIR"

echo "[e2e-supa] SUPABASE: start/check stack"
/workspace/scripts/e2e-supa-env.sh \
  --supabase-workdir "$SUPABASE_WORKDIR" \
  --supabase-env-out "$SUPABASE_ENV_FILE" \
  --worktree-id "$WORKTREE_ID" \
  --run-id "$RUN_ID" \
  --runner-network-mode "$RUNNER_NETWORK_MODE" \
  --reuse "$SUPABASE_REUSE"

load_env_file "$SUPABASE_ENV_FILE"

echo "[e2e-supa] DB: wait/migrate"
for i in $(seq 1 "$DB_WAIT_SECONDS"); do
  if psql "$DB_URL" -c 'SELECT 1' >/dev/null 2>&1; then
    break
  fi
  if [[ "$i" -eq "$DB_WAIT_SECONDS" ]]; then
    echo "[e2e-supa] DB failed to become ready" >&2
    exit 1
  fi
  sleep 1
done
DB_URL="$DB_URL" MIGRATIONS_DIR="$WORKSPACE_DIR/migrations" /workspace/scripts/e2e-apply-migrations.sh

echo "[e2e-supa] API: start"
(
  cd "$WORKSPACE_DIR"
  APP_ENV=development \
  AUTH_ISSUER= \
  AUTH_AUDIENCE= \
  AUTH_ALLOWED_ALGS="$AUTH_ALLOWED_ALGS" \
  AUTH_JWKS_URL="$SUPABASE_JWKS_URL" \
  DB_URL="${DB_URL:-}" \
  METRICS_ENABLED="${METRICS_ENABLED:-false}" \
  PORT=8080 \
  go run ./cmd/api >"$API_LOG_FILE" 2>&1
) &
API_PID=$!

API_READY="false"
for _ in $(seq 1 "$HEALTH_TIMEOUT_SECONDS"); do
  if ! kill -0 "$API_PID" >/dev/null 2>&1; then
    echo "[e2e-supa] API process exited before healthcheck passed" >&2
    tail -n 120 "$API_LOG_FILE" >&2 || true
    exit 1
  fi

  if curl --max-time 2 --connect-timeout 1 -fsS "http://127.0.0.1:8080/api/livez" >/dev/null 2>&1; then
    API_READY="true"
    break
  fi
  sleep 1
done
if [[ "$API_READY" != "true" ]]; then
  echo "[e2e-supa] API failed to boot" >&2
  tail -n 80 "$API_LOG_FILE" >&2 || true
  exit 1
fi

echo "[e2e-supa] AUTH: generate token"
/workspace/scripts/e2e-supa-token.sh \
  --supabase-env "$SUPABASE_ENV_FILE" \
  --email "$E2E_TEST_EMAIL" \
  --password "$E2E_TEST_PASSWORD" \
  --token-out "$TOKEN_ENV_FILE"

load_env_file "$TOKEN_ENV_FILE"
if [[ -z "${E2E_JWT_TOKEN:-}" || -z "${E2E_EXPECTED_SUB:-}" ]]; then
  echo "[e2e-supa] missing token or expected subject" >&2
  exit 1
fi
if [[ "$E2E_VERBOSE" == "1" ]]; then
  echo "[e2e-supa] token alg=${E2E_TOKEN_ALG:-unknown} kid=${E2E_TOKEN_KID:-missing} allowed_algs=${AUTH_ALLOWED_ALGS}"
fi

echo "[e2e-supa] E2E: run go tests"
verbose=""
if [[ "$E2E_VERBOSE" == "1" ]]; then
  verbose="-v"
fi

cd "$WORKSPACE_DIR"
pkgs=$(go list -tags=e2e -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./test/e2e/... 2>/dev/null | tr '\n' ' ')
if [[ -z "$pkgs" ]]; then
  pkgs="./test/e2e/..."
fi

API_BASE_URL="http://127.0.0.1:8080" \
E2E_JWT_TOKEN="$E2E_JWT_TOKEN" \
E2E_EXPECTED_SUB="$E2E_EXPECTED_SUB" \
E2E_EXPECT_DB_READY=true \
go test -count=1 -tags=e2e $verbose $pkgs

echo "[e2e-supa] REPORT: export"
copy_reports

echo "[e2e-supa] done"
