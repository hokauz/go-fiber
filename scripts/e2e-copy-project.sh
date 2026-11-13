#!/usr/bin/env bash
# e2e-copy-project.sh: rsync source to workspace for e2e runs

set -euo pipefail

SOURCE_DIR="${SOURCE_DIR:-.}"
WORKSPACE_DIR="${WORKSPACE_DIR:-/workspace/api}"

mkdir -p "$WORKSPACE_DIR"
rsync -a --delete \
  --exclude '.git' \
  --exclude 'tmp' \
  --exclude 'e2e-report' \
  --exclude '.env' \
  "$SOURCE_DIR/" "$WORKSPACE_DIR/"

