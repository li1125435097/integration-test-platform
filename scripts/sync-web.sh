#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEB="$ROOT/web"
DEST="$ROOT/server/embedded/web"

cd "$WEB"
if [ ! -f package.json ]; then
  echo "sync-web: missing web/package.json" >&2
  exit 1
fi
npm ci
npm run build

rm -rf "$DEST"
mkdir -p "$DEST"
cp -r "$WEB/dist"/. "$DEST/"

echo "sync-web: $WEB/dist -> $DEST"
