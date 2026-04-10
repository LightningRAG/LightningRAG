#!/usr/bin/env bash
# 校验 deploy 下 Compose / Kustomize / 脚本语法（不写集群、不启动容器）
# 在仓库根目录执行: ./deploy/scripts/check-deploy-config.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"
DCO="deploy/docker-compose-online"
fail=0

run() {
  if ! "$@"; then
    echo "失败: $*" >&2
    fail=1
  fi
}

echo "==> docker-compose-online（预构建镜像 + 中间件）"
run docker compose -f "$DCO/docker-compose-base.yaml" -f "$DCO/docker-compose.yaml" config >/dev/null

echo "==> docker-compose-online（本地构建）"
run docker compose -f "$DCO/docker-compose-base.yaml" -f "$DCO/docker-compose.local.yaml" config >/dev/null

echo "==> docker-compose-online（示例 profile）"
run env COMPOSE_PROFILES=elasticsearch,minio docker compose -f "$DCO/docker-compose-base.yaml" -f "$DCO/docker-compose.yaml" config >/dev/null

echo "==> deploy/docker-compose（开发固定 IP）"
run docker compose -f deploy/docker-compose/docker-compose.yaml config >/dev/null

echo "==> verify-deployment.sh 语法"
run bash -n deploy/scripts/verify-deployment.sh

if command -v kubectl >/dev/null 2>&1; then
  echo "==> kubectl kustomize deploy/kubernetes"
  run kubectl kustomize deploy/kubernetes >/dev/null
else
  echo "==> 未找到 kubectl，跳过 kustomize"
fi

if [[ "$fail" -ne 0 ]]; then
  echo "部分检查未通过。" >&2
  exit 1
fi
echo "全部检查通过。"
