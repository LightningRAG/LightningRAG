#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/web/dist"
DST="$ROOT/server/webui/webdist"
if [[ ! -f "$SRC/index.html" ]]; then
  echo "error: 缺少 $SRC/index.html，请先构建前端（例如: cd web && yarn install && yarn build）" >&2
  exit 1
fi
mkdir -p "$DST"
rsync -a --delete --exclude '.gitkeep' "$SRC/" "$DST/"
echo "已同步前端产物: $SRC -> $DST"
