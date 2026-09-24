#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [[ "$(basename "$SCRIPT_DIR")" == "scripts" ]]; then
  ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
else
  ROOT="$SCRIPT_DIR"
fi
cd "$ROOT"

export VERSION="${VERSION:-0.1.0-dev}"
if command -v git >/dev/null 2>&1; then
  VERSION="$(git describe --tags --always 2>/dev/null || echo "$VERSION")"
  export VERSION
fi

descriptions=(
  "同步前端到 server/embedded"
  "开发运行服务（Go 后端）"
  "运行前端开发服务器（Vite）"
  "构建当前平台"
  "构建全平台"
  "发布构建（全平台 + zip）"
  "整理 Go 依赖"
)
invoke=(
  "bash scripts/sync-web.sh"
  "bash scripts/run.sh"
  "bash scripts/dev-web.sh"
  "bash scripts/build.sh build"
  "bash scripts/build.sh build-all"
  "bash scripts/build.sh release"
  "bash scripts/tidy.sh"
)

run_choice() {
  local choice="$1"
  case "$choice" in
    1) bash scripts/sync-web.sh ;;
    2) bash scripts/run.sh ;;
    3) bash scripts/dev-web.sh ;;
    4) bash scripts/build.sh build ;;
    5) bash scripts/build.sh build-all ;;
    6) bash scripts/build.sh release ;;
    7) bash scripts/tidy.sh ;;
    *)
      echo "未知选项: $choice" >&2
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
    printf "      %s\n" "${invoke[$idx]}"
    i=$((i + 1))
  done
  echo "  0) 退出"
  echo
  read -rp "请选择序号: " choice

  if [[ "$choice" == "0" ]]; then
    echo "已退出。"
    exit 0
  fi

  if ! [[ "$choice" =~ ^[0-9]+$ ]] || (( choice < 1 || choice > ${#descriptions[@]} )); then
    echo "无效选择，请输入 0–${#descriptions[@]} 之间的序号。"
    continue
  fi

  cmd="${invoke[$((choice - 1))]}"
  echo
  echo ">>> 执行: ${cmd}"
  echo
  run_choice "$choice"
  echo
  echo "完成: ${cmd}"
done
