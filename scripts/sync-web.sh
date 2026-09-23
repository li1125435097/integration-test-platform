#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/web"
DEST="$ROOT/server/embedded/web"
rm -rf "$DEST"
mkdir -p "$DEST"
cp -r "$SRC"/. "$DEST/"

echo "sync-web: $SRC -> $DEST"
