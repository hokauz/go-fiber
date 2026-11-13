#!/usr/bin/env bash
# e2e-supa-copy-project.sh: copy source to workspace for supabase e2e runs

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <source_dir> <dest_dir>" >&2
  exit 1
fi

SOURCE_DIR="$1"
DEST_DIR="$2"

if [[ ! -d "$SOURCE_DIR" ]]; then
  echo "source dir does not exist: $SOURCE_DIR" >&2
  exit 1
fi

mkdir -p "$DEST_DIR"

rsync -a --delete \
  --exclude ".git/" \
  --exclude ".env*" \
  --exclude "tmp/" \
  --exclude "e2e-report/" \
  --exclude "e2e-report-supa/" \
  "$SOURCE_DIR"/ "$DEST_DIR"/

