#!/usr/bin/env bash

set -euo pipefail

SUPABASE_WORKDIR=""
SUPABASE_ENV_OUT=""
WORKTREE_ID=""
RUN_ID=""
RUNNER_NETWORK_MODE="host"
SUPABASE_REUSE="0"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --supabase-workdir)
      SUPABASE_WORKDIR="$2"
      shift 2
      ;;
    --supabase-env-out)
      SUPABASE_ENV_OUT="$2"
      shift 2
      ;;
    --worktree-id)
      WORKTREE_ID="$2"
      shift 2
      ;;
    --run-id)
      RUN_ID="$2"
      shift 2
      ;;
    --runner-network-mode)
      RUNNER_NETWORK_MODE="$2"
      shift 2
      ;;
    --reuse)
      SUPABASE_REUSE="$2"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$SUPABASE_WORKDIR" || -z "$SUPABASE_ENV_OUT" || -z "$WORKTREE_ID" || -z "$RUN_ID" ]]; then
  echo "required args: --supabase-workdir --supabase-env-out --worktree-id --run-id" >&2
  exit 1
fi

mkdir -p "$SUPABASE_WORKDIR"
if [[ ! -f "$SUPABASE_WORKDIR/supabase/config.toml" ]]; then
  supabase init --workdir "$SUPABASE_WORKDIR" --yes >/dev/null
fi

CONFIG_FILE="$SUPABASE_WORKDIR/supabase/config.toml"

if [[ "$SUPABASE_REUSE" == "1" ]]; then
  PORT_SEED="${WORKTREE_ID}-shared"
  PROJECT_ID_SEED="${WORKTREE_ID}-shared"
else
  PORT_SEED="${WORKTREE_ID}-${RUN_ID}"
  PROJECT_ID_SEED="${WORKTREE_ID}-${RUN_ID}"
fi

PROJECT_ID="$(printf '%s' "$PROJECT_ID_SEED" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/_/g; s/^_+//; s/_+$//')"
PROJECT_ID="${PROJECT_ID:0:40}"
if [[ -z "$PROJECT_ID" ]]; then
  PROJECT_ID="go_fiber"
fi
PROJECT_ID="go_fiber_${PROJECT_ID}"

CHECKSUM="$(printf '%s' "$PORT_SEED" | cksum | awk '{print $1}')"
BASE_PORT="$((21000 + (CHECKSUM % 20000)))"
SUPABASE_API_PORT="$BASE_PORT"
SUPABASE_DB_PORT="$((BASE_PORT + 1))"
SUPABASE_STUDIO_PORT="$((BASE_PORT + 2))"
SUPABASE_INBUCKET_PORT="$((BASE_PORT + 3))"
SUPABASE_DB_SHADOW_PORT="$((BASE_PORT + 4))"
SUPABASE_POOLER_PORT="$((BASE_PORT + 5))"
SUPABASE_ANALYTICS_PORT="$((BASE_PORT + 6))"
SUPABASE_EDGE_INSPECTOR_PORT="$((BASE_PORT + 7))"

sed -i \
  -e "s/^project_id = .*/project_id = \"${PROJECT_ID}\"/" \
  -e "s/port = 54321/port = ${SUPABASE_API_PORT}/" \
  -e "s/port = 54322/port = ${SUPABASE_DB_PORT}/" \
  -e "s/shadow_port = 54320/shadow_port = ${SUPABASE_DB_SHADOW_PORT}/" \
  -e "s/port = 54329/port = ${SUPABASE_POOLER_PORT}/" \
  -e "s/port = 54323/port = ${SUPABASE_STUDIO_PORT}/" \
  -e "s/port = 54324/port = ${SUPABASE_INBUCKET_PORT}/" \
  -e "s/port = 54327/port = ${SUPABASE_ANALYTICS_PORT}/" \
  -e "s/inspector_port = 8083/inspector_port = ${SUPABASE_EDGE_INSPECTOR_PORT}/" \
  "$CONFIG_FILE"

RAW_ENV_FILE="${SUPABASE_WORKDIR}/supabase-status.env"

is_running() {
  supabase status --workdir "$SUPABASE_WORKDIR" -o env >"$RAW_ENV_FILE" 2>/dev/null
}

start_stack() {
  supabase start --workdir "$SUPABASE_WORKDIR" --yes >/dev/null
}

if [[ "$SUPABASE_REUSE" == "1" ]]; then
  if ! is_running; then
    if ! start_stack; then
      supabase stop --workdir "$SUPABASE_WORKDIR" --no-backup >/dev/null 2>&1 || true
      start_stack
    fi
    is_running
  fi
else
  start_stack
  is_running
fi

set -a
# shellcheck source=/dev/null
source "$RAW_ENV_FILE"
set +a

if [[ -z "${API_URL:-}" || -z "${ANON_KEY:-}" || -z "${SERVICE_ROLE_KEY:-}" ]]; then
  echo "missing required supabase status variables (API_URL/ANON_KEY/SERVICE_ROLE_KEY)" >&2
  exit 1
fi

DB_URL_VALUE="${DB_URL:-${POSTGRES_URL:-}}"
if [[ -z "$DB_URL_VALUE" ]]; then
  DB_URL_VALUE="postgresql://postgres:postgres@127.0.0.1:${SUPABASE_DB_PORT}/postgres"
fi

rewrite_host() {
  printf '%s' "$1" | sed -E 's#(://)(127\.0\.0\.1|localhost)#\1host.docker.internal#g'
}

if [[ "$RUNNER_NETWORK_MODE" == "host" ]]; then
  API_URL_HOST="$API_URL"
  DB_URL_HOST="$DB_URL_VALUE"
else
  API_URL_HOST="$(rewrite_host "$API_URL")"
  DB_URL_HOST="$(rewrite_host "$DB_URL_VALUE")"
fi
SUPABASE_JWKS_URL="${API_URL_HOST%/}/auth/v1/.well-known/jwks.json"

mkdir -p "$(dirname "$SUPABASE_ENV_OUT")"
cat > "$SUPABASE_ENV_OUT" <<EOF
API_URL=$API_URL_HOST
SUPABASE_URL=$API_URL_HOST
ANON_KEY=$ANON_KEY
SUPABASE_ANON_KEY=$ANON_KEY
SERVICE_ROLE_KEY=$SERVICE_ROLE_KEY
SUPABASE_SERVICE_ROLE_KEY=$SERVICE_ROLE_KEY
SUPABASE_JWKS_URL=$SUPABASE_JWKS_URL
DB_URL=$DB_URL_HOST
METRICS_ENABLED=false
SUPABASE_API_PORT=$SUPABASE_API_PORT
SUPABASE_DB_PORT=$SUPABASE_DB_PORT
SUPABASE_STUDIO_PORT=$SUPABASE_STUDIO_PORT
SUPABASE_INBUCKET_PORT=$SUPABASE_INBUCKET_PORT
SUPABASE_ANALYTICS_PORT=$SUPABASE_ANALYTICS_PORT
SUPABASE_EDGE_INSPECTOR_PORT=$SUPABASE_EDGE_INSPECTOR_PORT
EOF
