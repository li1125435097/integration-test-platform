#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MODE="${1:-build}"
# build: current platform | build-all: all platforms | release: all + zip
VERSION="${VERSION:-0.1.0-dev}"
if command -v git >/dev/null 2>&1; then
  VERSION="$(git describe --tags --always 2>/dev/null || echo "$VERSION")"
fi

LDFLAGS="-s -w -X main.version=${VERSION}"
PLATFORMS="${PLATFORMS:-linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64}"

"$ROOT/scripts/sync-web.sh"

build_one() {
  local goos goarch out name
  goos="${1%%/*}"
  goarch="${1##*/}"
  name="itp-${goos}-${goarch}"
  if [ "$goos" = "windows" ]; then
    out="dist/${name}.exe"
  else
    out="dist/${name}"
  fi
  echo "building $goos/$goarch -> $out"
  GOOS="$goos" GOARCH="$goarch" go build -tags release -trimpath \
    -ldflags "$LDFLAGS" -o "$out" ./server
}

mkdir -p dist dist/release

if [ "$MODE" = "release" ] || [ "$MODE" = "build-all" ]; then
  for p in $PLATFORMS; do
    build_one "$p"
    if [ "$MODE" = "release" ]; then
      goos="${p%%/*}"
      goarch="${p##*/}"
      zipname="itp-${VERSION}-${goos}-${goarch}"
      staged="dist/release/${zipname}"
      rm -rf "$staged"
      mkdir -p "$staged/config"
      if [ "$goos" = "windows" ]; then
        cp "dist/itp-${goos}-${goarch}.exe" "$staged/integration-test-platform.exe"
      else
        cp "dist/itp-${goos}-${goarch}" "$staged/integration-test-platform"
      fi
      cp config/menu.json "$staged/config/"
      cat > "$staged/RUN.txt" <<EOF
集成测试平台

启动:
  Windows: integration-test-platform.exe
  Linux/macOS: ./integration-test-platform

默认访问: http://127.0.0.1:8080
菜单配置: config/menu.json
EOF
      (cd dist/release && rm -f "${zipname}.zip" && zip -rq "${zipname}.zip" "${zipname}")
      echo "release zip: dist/release/${zipname}.zip"
    fi
  done
else
  build_one "$(go env GOOS)/$(go env GOARCH)"
fi
