#!/usr/bin/env bash
# e2e-apply-migrations.sh: applies SQL migrations exactly once per database.

set -euo pipefail

DB_URL="${DB_URL:-}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./migrations}"

if [[ -z "$DB_URL" ]]; then
  echo "[e2e-migrations] DB_URL is required" >&2
  exit 1
fi

if [[ ! -d "$MIGRATIONS_DIR" ]]; then
  echo "[e2e-migrations] migrations directory not found: $MIGRATIONS_DIR" >&2
  exit 1
fi

shopt -s nullglob
files=("$MIGRATIONS_DIR"/*.sql)
shopt -u nullglob

if [[ ${#files[@]} -eq 0 ]]; then
  echo "[e2e-migrations] no migration files found in $MIGRATIONS_DIR"
  exit 0
fi

export PSQL_PAGER=off
psql_base=(psql "$DB_URL" -v ON_ERROR_STOP=1 -X -q)

extract_up_sql() {
  local file="$1"
  local out="$2"
  if grep -q '^-- +migrate Up' "$file"; then
    awk '
      BEGIN { in_up=0 }
      /^-- \+migrate Up$/ { in_up=1; next }
      /^-- \+migrate Down$/ { in_up=0; exit }
      { if (in_up==1) print }
    ' "$file" > "$out"
  else
    cat "$file" > "$out"
  fi
}

table_exists() {
  local table_name="$1"
  local exists
  exists="$("${psql_base[@]}" -tA -c "SELECT to_regclass('$table_name') IS NOT NULL;")"
  [[ "$exists" == "t" ]]
}

"${psql_base[@]}" <<'SQL'
CREATE TABLE IF NOT EXISTS app_e2e_migrations (
  filename   TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL

for file in "${files[@]}"; do
  filename="$(basename "$file")"
  tmp_sql="$(mktemp)"
  extract_up_sql "$file" "$tmp_sql"

  if [[ ! -s "$tmp_sql" ]]; then
    rm -f "$tmp_sql"
    echo "[e2e-migrations] empty up migration for $filename" >&2
    exit 1
  fi

  mapfile -t created_tables < <(grep -Eio 'CREATE[[:space:]]+TABLE[[:space:]]+(IF[[:space:]]+NOT[[:space:]]+EXISTS[[:space:]]+)?[A-Za-z0-9_\."]+' "$tmp_sql" \
    | sed -E 's/CREATE[[:space:]]+TABLE[[:space:]]+(IF[[:space:]]+NOT[[:space:]]+EXISTS[[:space:]]+)?//I' \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/"//g')

  applied="$("${psql_base[@]}" -tA -c "SELECT 1 FROM app_e2e_migrations WHERE filename = '$filename' LIMIT 1;")"
  if [[ "$applied" == "1" ]]; then
    missing=0
    for table_name in "${created_tables[@]}"; do
      if [[ -n "$table_name" ]] && ! table_exists "$table_name"; then
        missing=1
        break
      fi
    done
    if [[ "$missing" == "0" ]]; then
      rm -f "$tmp_sql"
      echo "[e2e-migrations] skip $filename (already applied)"
      continue
    fi
    echo "[e2e-migrations] drift detected for $filename, reapplying"
    "${psql_base[@]}" -c "DELETE FROM app_e2e_migrations WHERE filename = '$filename';"
  fi

  echo "[e2e-migrations] applying $filename"

  "${psql_base[@]}" -f "$tmp_sql"
  rm -f "$tmp_sql"
  "${psql_base[@]}" -c "INSERT INTO app_e2e_migrations (filename) VALUES ('$filename');"
done

echo "[e2e-migrations] done"
