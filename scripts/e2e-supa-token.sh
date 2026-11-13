#!/usr/bin/env bash

set -euo pipefail

SUPABASE_ENV_FILE=""
EMAIL=""
PASSWORD=""
TOKEN_OUT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --supabase-env)
      SUPABASE_ENV_FILE="$2"
      shift 2
      ;;
    --email)
      EMAIL="$2"
      shift 2
      ;;
    --password)
      PASSWORD="$2"
      shift 2
      ;;
    --token-out)
      TOKEN_OUT="$2"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$SUPABASE_ENV_FILE" || -z "$EMAIL" || -z "$PASSWORD" || -z "$TOKEN_OUT" ]]; then
  echo "required args: --supabase-env --email --password --token-out" >&2
  exit 1
fi

set -a
# shellcheck source=/dev/null
source "$SUPABASE_ENV_FILE"
set +a

if [[ -z "${API_URL:-}" || -z "${ANON_KEY:-}" || -z "${SERVICE_ROLE_KEY:-}" ]]; then
  echo "missing API_URL/ANON_KEY/SERVICE_ROLE_KEY in $SUPABASE_ENV_FILE" >&2
  exit 1
fi

SUPABASE_URL="${API_URL%/}"

CREATE_PAYLOAD="$(jq -nc \
  --arg email "$EMAIL" \
  --arg password "$PASSWORD" \
  '{"email":$email,"password":$password,"email_confirm":true}')"

CREATE_RESPONSE="$(curl -sS -w '\n%{http_code}' \
  -X POST "${SUPABASE_URL}/auth/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "apikey: ${SERVICE_ROLE_KEY}" \
  -H "Authorization: Bearer ${SERVICE_ROLE_KEY}" \
  -d "$CREATE_PAYLOAD")"

CREATE_BODY="$(printf '%s' "$CREATE_RESPONSE" | sed '$d')"
CREATE_STATUS="$(printf '%s' "$CREATE_RESPONSE" | tail -n 1)"

if [[ "$CREATE_STATUS" != "200" && "$CREATE_STATUS" != "201" ]]; then
  if ! printf '%s' "$CREATE_BODY" | grep -qiE 'already registered|already exists|already been registered'; then
    echo "failed to create e2e user (status=$CREATE_STATUS)" >&2
    echo "$CREATE_BODY" >&2
    exit 1
  fi
fi

LOGIN_PAYLOAD="$(jq -nc \
  --arg email "$EMAIL" \
  --arg password "$PASSWORD" \
  '{"email":$email,"password":$password}')"

LOGIN_RESPONSE="$(curl -sS -w '\n%{http_code}' \
  -X POST "${SUPABASE_URL}/auth/v1/token?grant_type=password" \
  -H "Content-Type: application/json" \
  -H "apikey: ${ANON_KEY}" \
  -d "$LOGIN_PAYLOAD")"

LOGIN_BODY="$(printf '%s' "$LOGIN_RESPONSE" | sed '$d')"
LOGIN_STATUS="$(printf '%s' "$LOGIN_RESPONSE" | tail -n 1)"

if [[ "$LOGIN_STATUS" != "200" ]]; then
  echo "failed to login e2e user (status=$LOGIN_STATUS)" >&2
  echo "$LOGIN_BODY" >&2
  exit 1
fi

ACCESS_TOKEN="$(printf '%s' "$LOGIN_BODY" | jq -r '.access_token // empty')"
if [[ -z "$ACCESS_TOKEN" ]]; then
  echo "missing access_token in auth response" >&2
  echo "$LOGIN_BODY" >&2
  exit 1
fi

TOKEN_META="$(python3 - "$ACCESS_TOKEN" <<'PY'
import base64
import json
import sys

token = sys.argv[1]
parts = token.split(".")
if len(parts) != 3:
    print("||")
    sys.exit(0)
header = parts[0]
payload = parts[1]
header += "=" * (-len(header) % 4)
payload += "=" * (-len(payload) % 4)
try:
    hdr = json.loads(base64.urlsafe_b64decode(header.encode("utf-8")))
    data = json.loads(base64.urlsafe_b64decode(payload.encode("utf-8")))
except Exception:
    print("||")
    sys.exit(0)
print(f'{data.get("sub", "")}|{hdr.get("alg", "")}|{hdr.get("kid", "")}')
PY
)"

SUBJECT="$(printf '%s' "$TOKEN_META" | cut -d'|' -f1)"
TOKEN_ALG="$(printf '%s' "$TOKEN_META" | cut -d'|' -f2)"
TOKEN_KID="$(printf '%s' "$TOKEN_META" | cut -d'|' -f3)"

if [[ -z "$SUBJECT" ]]; then
  echo "missing subject in JWT payload" >&2
  exit 1
fi

mkdir -p "$(dirname "$TOKEN_OUT")"
cat > "$TOKEN_OUT" <<EOF
E2E_JWT_TOKEN=$ACCESS_TOKEN
E2E_EXPECTED_SUB=$SUBJECT
E2E_USER_EMAIL=$EMAIL
E2E_TOKEN_ALG=$TOKEN_ALG
E2E_TOKEN_KID=$TOKEN_KID
EOF
