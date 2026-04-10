#!/usr/bin/env bash
# 在仓库根目录执行：构建前端、同步到 server/webui/webdist、编译 server 二进制（内含 embed 静态资源）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
if command -v make >/dev/null 2>&1; then
  make build-server-embed-local TAGS_OPT="${TAGS_OPT:-latest}"
else
  (cd web && yarn install && yarn build)
  bash "$ROOT/scripts/sync-web-dist.sh"
  cd "$ROOT/server"
  go env -w GO111MODULE=on
  go env -w GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
  go env -w CGO_ENABLED=0
  go mod tidy
  go build -ldflags "-B 0x$(head -c8 /dev/urandom | od -An -tx1 | tr -d ' \n') -X main.Version=${TAGS_OPT:-latest}" -v -o "${OUTPUT:-server}" .
fi
