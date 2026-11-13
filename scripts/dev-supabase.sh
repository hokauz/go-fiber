#!/usr/bin/env bash

set -euo pipefail

ACTION="${1:-}"
if [[ -n "$ACTION" ]]; then
  shift
fi

SUPABASE_WORKDIR=""
ENV_FILE=""
WORKTREE_ID=""

usage() {
  cat <<'EOF' >&2
usage:
  dev-supabase.sh up --workdir <path> --env-file <path> --worktree-id <id>
  dev-supabase.sh stop --workdir <path>
  dev-supabase.sh reset --workdir <path> --env-file <path>
EOF
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workdir)
      SUPABASE_WORKDIR="$2"
      shift 2
      ;;
    --env-file)
      ENV_FILE="$2"
      shift 2
      ;;
    --worktree-id)
      WORKTREE_ID="$2"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      ;;
  esac
done

if [[ -z "$ACTION" ]]; then
  usage
fi

if ! command -v supabase >/dev/null 2>&1; then
  echo "supabase CLI not found. install it and try again." >&2
  exit 1
fi

if [[ -z "$SUPABASE_WORKDIR" ]]; then
  echo "--workdir is required" >&2
  usage
fi

if [[ "$ACTION" == "up" ]]; then
  if [[ -z "$ENV_FILE" || -z "$WORKTREE_ID" ]]; then
    echo "up requires --env-file and --worktree-id" >&2
    usage
  fi
fi

if [[ "$ACTION" == "reset" && -z "$ENV_FILE" ]]; then
  echo "reset requires --env-file" >&2
  usage
fi

MANAGED_KEYS=(
  APP_ENV
  AUTH_JWKS_URL
  DB_URL
  SUPABASE_URL
  SUPABASE_ANON_KEY
  SUPABASE_SERVICE_ROLE_KEY
  SUPABASE_PROJECT_ID
  SUPABASE_WORKDIR
  SUPABASE_API_PORT
  SUPABASE_DB_PORT
  SUPABASE_STUDIO_PORT
  SUPABASE_INBUCKET_PORT
  SUPABASE_ANALYTICS_PORT
  SUPABASE_EDGE_INSPECTOR_PORT
)

upsert_env() {
  local file="$1"
  local key="$2"
  local value="$3"
  local tmp_file

  mkdir -p "$(dirname "$file")"
  tmp_file="$(mktemp "${file}.tmp.XXXXXX")"

  if [[ -f "$file" ]]; then
    awk -v key="$key" -v value="$value" '
      BEGIN { replaced = 0 }
      $0 ~ "^" key "=" {
        print key "=" value
        replaced = 1
        next
      }
      { print }
      END {
        if (!replaced) {
          print key "=" value
        }
      }
    ' "$file" >"$tmp_file"
  else
    printf '%s=%s\n' "$key" "$value" >"$tmp_file"
  fi

  mv "$tmp_file" "$file"
}

remove_env_key() {
  local file="$1"
  local key="$2"
  local tmp_file

  [[ -f "$file" ]] || return 0
  tmp_file="$(mktemp "${file}.tmp.XXXXXX")"
  awk -v key="$key" '$0 !~ "^" key "=" { print }' "$file" >"$tmp_file"
  mv "$tmp_file" "$file"
}

configure_workdir() {
  local config_file="$SUPABASE_WORKDIR/supabase/config.toml"
  local port_seed project_seed checksum base_port
  local project_id

  mkdir -p "$SUPABASE_WORKDIR"
  if [[ ! -f "$config_file" ]]; then
    supabase init --workdir "$SUPABASE_WORKDIR" --yes >/dev/null
  fi

  port_seed="${WORKTREE_ID}-dev"
  project_seed="${WORKTREE_ID}-dev"
  checksum="$(printf '%s' "$port_seed" | cksum | awk '{print $1}')"
  base_port="$((21000 + (checksum % 20000)))"

  SUPABASE_API_PORT="$base_port"
  SUPABASE_DB_PORT="$((base_port + 1))"
  SUPABASE_STUDIO_PORT="$((base_port + 2))"
  SUPABASE_INBUCKET_PORT="$((base_port + 3))"
  SUPABASE_DB_SHADOW_PORT="$((base_port + 4))"
  SUPABASE_POOLER_PORT="$((base_port + 5))"
  SUPABASE_ANALYTICS_PORT="$((base_port + 6))"
  SUPABASE_EDGE_INSPECTOR_PORT="$((base_port + 7))"

  project_id="$(printf '%s' "$project_seed" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/_/g; s/^_+//; s/_+$//')"
  project_id="${project_id:0:40}"
  if [[ -z "$project_id" ]]; then
    project_id="go_fiber_dev"
  fi
  SUPABASE_PROJECT_ID="go_fiber_${project_id}"

  sed -i.bak \
    -e "s/^project_id = .*/project_id = \"${SUPABASE_PROJECT_ID}\"/" \
    -e "s/port = 54321/port = ${SUPABASE_API_PORT}/" \
    -e "s/port = 54322/port = ${SUPABASE_DB_PORT}/" \
    -e "s/shadow_port = 54320/shadow_port = ${SUPABASE_DB_SHADOW_PORT}/" \
    -e "s/port = 54329/port = ${SUPABASE_POOLER_PORT}/" \
    -e "s/port = 54323/port = ${SUPABASE_STUDIO_PORT}/" \
    -e "s/port = 54324/port = ${SUPABASE_INBUCKET_PORT}/" \
    -e "s/port = 54327/port = ${SUPABASE_ANALYTICS_PORT}/" \
    -e "s/inspector_port = 8083/inspector_port = ${SUPABASE_EDGE_INSPECTOR_PORT}/" \
    "$config_file"
  rm -f "${config_file}.bak"
}

status_env_file="$SUPABASE_WORKDIR/supabase-status.env"

stack_is_running() {
  supabase status --workdir "$SUPABASE_WORKDIR" -o env >"$status_env_file" 2>/dev/null
}

up_stack() {
  configure_workdir
  if ! stack_is_running; then
    supabase start --workdir "$SUPABASE_WORKDIR" --yes >/dev/null
    stack_is_running
  fi
}

stop_stack() {
  if [[ -d "$SUPABASE_WORKDIR" ]]; then
    supabase stop --workdir "$SUPABASE_WORKDIR" --no-backup >/dev/null 2>&1 || true
  fi
}

write_dev_env() {
  local api_url anon_key service_role_key db_url jwks_url

  set -a
  # shellcheck source=/dev/null
  source "$status_env_file"
  set +a

  api_url="${API_URL:-}"
  anon_key="${ANON_KEY:-}"
  service_role_key="${SERVICE_ROLE_KEY:-}"
  db_url="${DB_URL:-${POSTGRES_URL:-}}"
  if [[ -z "$db_url" ]]; then
    db_url="postgresql://postgres:postgres@127.0.0.1:${SUPABASE_DB_PORT}/postgres"
  fi

  if [[ -z "$api_url" || -z "$anon_key" || -z "$service_role_key" ]]; then
    echo "missing required supabase env keys (API_URL/ANON_KEY/SERVICE_ROLE_KEY)" >&2
    exit 1
  fi

  jwks_url="${api_url%/}/auth/v1/.well-known/jwks.json"

  upsert_env "$ENV_FILE" "APP_ENV" "development"
  upsert_env "$ENV_FILE" "AUTH_JWKS_URL" "$jwks_url"
  upsert_env "$ENV_FILE" "DB_URL" "$db_url"
  upsert_env "$ENV_FILE" "SUPABASE_URL" "$api_url"
  upsert_env "$ENV_FILE" "SUPABASE_ANON_KEY" "$anon_key"
  upsert_env "$ENV_FILE" "SUPABASE_SERVICE_ROLE_KEY" "$service_role_key"
  upsert_env "$ENV_FILE" "SUPABASE_PROJECT_ID" "$SUPABASE_PROJECT_ID"
  upsert_env "$ENV_FILE" "SUPABASE_WORKDIR" "$SUPABASE_WORKDIR"
  upsert_env "$ENV_FILE" "SUPABASE_API_PORT" "$SUPABASE_API_PORT"
  upsert_env "$ENV_FILE" "SUPABASE_DB_PORT" "$SUPABASE_DB_PORT"
  upsert_env "$ENV_FILE" "SUPABASE_STUDIO_PORT" "$SUPABASE_STUDIO_PORT"
  upsert_env "$ENV_FILE" "SUPABASE_INBUCKET_PORT" "$SUPABASE_INBUCKET_PORT"
  upsert_env "$ENV_FILE" "SUPABASE_ANALYTICS_PORT" "$SUPABASE_ANALYTICS_PORT"
  upsert_env "$ENV_FILE" "SUPABASE_EDGE_INSPECTOR_PORT" "$SUPABASE_EDGE_INSPECTOR_PORT"
}

cleanup_env() {
  local key
  [[ -n "$ENV_FILE" ]] || return 0
  for key in "${MANAGED_KEYS[@]}"; do
    remove_env_key "$ENV_FILE" "$key"
  done
}

case "$ACTION" in
  up)
    up_stack
    write_dev_env
    printf 'supabase dev ready: %s\n' "$ENV_FILE"
    ;;
  stop)
    stop_stack
    printf 'supabase dev stopped: %s\n' "$SUPABASE_WORKDIR"
    ;;
  reset)
    stop_stack
    rm -rf "$SUPABASE_WORKDIR"
    cleanup_env
    printf 'supabase dev reset: %s\n' "$SUPABASE_WORKDIR"
    ;;
  *)
    usage
    ;;
esac
