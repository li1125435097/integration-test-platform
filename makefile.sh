#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export VERSION="${VERSION:-0.1.0-dev}"
if command -v git >/dev/null 2>&1; then
  VERSION="$(git describe --tags --always 2>/dev/null || echo "$VERSION")"
  export VERSION
fi

targets=(sync-web run build build-all release tidy)
descriptions=(
  "同步前端到 server/embedded (scripts/sync-web.sh)"
  "开发运行 (go run ./server -config-dir ./config)"
  "构建当前平台 (sync-web + scripts/build.sh build)"
  "构建全平台 (scripts/build.sh build-all)"
  "发布构建 (scripts/build.sh release)"
  "整理 Go 依赖 (go mod tidy)"
)

run_target() {
  local target="$1"
  case "$target" in
    sync-web)
      bash scripts/sync-web.sh
      ;;
    run)
      go run ./server -config-dir ./config
      ;;
    build)
      bash scripts/sync-web.sh
      bash scripts/build.sh build
      ;;
    build-all)
      bash scripts/build.sh build-all
      ;;
    release)
      bash scripts/build.sh release
      ;;
    tidy)
      go mod tidy
      ;;
    *)
      echo "未知目标: $target" >&2
      return 1
      ;;
  esac
}

while true; do
  echo
  echo "integration-test-platform — 可用命令 (VERSION=${VERSION})"
  echo "--------------------------------------------------------"
  i=1
  for desc in "${descriptions[@]}"; do
    idx=$((i - 1))
    printf "  %d) %s\n" "$i" "$desc"
    printf "      make %s\n" "${targets[$idx]}"
    i=$((i + 1))
  done
  echo "  0) 退出"
  echo
  read -rp "请选择序号: " choice

  if [[ "$choice" == "0" ]]; then
    echo "已退出。"
    exit 0
  fi

  if ! [[ "$choice" =~ ^[0-9]+$ ]] || (( choice < 1 || choice > ${#targets[@]} )); then
    echo "无效选择，请输入 0–${#targets[@]} 之间的序号。"
    continue
  fi

  target="${targets[$((choice - 1))]}"
  echo
  echo ">>> 执行: make ${target}"
  echo
  run_target "$target"
  echo
  echo "完成: make ${target}"
done
