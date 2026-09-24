#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEB="$ROOT/web"

cd "$WEB"
if [ ! -f package.json ]; then
  echo "dev-web: missing web/package.json" >&2
  exit 1
fi
if [ ! -d node_modules ]; then
  echo "dev-web: installing dependencies (npm ci)..."
  npm ci
fi
exec npm run dev
