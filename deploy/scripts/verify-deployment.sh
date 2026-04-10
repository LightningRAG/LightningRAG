#!/usr/bin/env bash
# 验证本机 docker-compose-online 或同类端口映射是否可用（需容器已启动）
# 用法：
#   ./deploy/scripts/verify-deployment.sh                    # 后端 + 前端 + /api 反代
#   ./deploy/scripts/verify-deployment.sh --api-only         # 仅后端（未起 lrag-web 时）
#   ./deploy/scripts/verify-deployment.sh --wait 120         # 启动后最多等待 120 秒再探测
#   ./deploy/scripts/verify-deployment.sh --wait 60 --api-only
# 覆盖端口：LRAG_SERVER_PORT=9888 LRAG_WEB_PORT=9080 ./deploy/scripts/verify-deployment.sh

set -euo pipefail

API_ONLY=false
WAIT_SECS=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-only)
      API_ONLY=true
      shift
      ;;
    --wait)
      WAIT_SECS="${2:?--wait 需要秒数}"
      shift 2
      ;;
    -h|--help)
      echo "用法: $0 [--api-only] [--wait 秒]"
      echo "环境变量: LRAG_SERVER_PORT, LRAG_WEB_PORT"
      exit 0
      ;;
    *)
      echo "未知参数: $1（试试 --help）" >&2
      exit 2
      ;;
  esac
done

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

LRAG_SERVER_PORT="${LRAG_SERVER_PORT:-8888}"
LRAG_WEB_PORT="${LRAG_WEB_PORT:-8080}"

check_server() {
  curl -sS -o /tmp/lrag-health.body -w '%{http_code}' "http://127.0.0.1:${LRAG_SERVER_PORT}/health" 2>/dev/null || echo "000"
}

check_web_root() {
  curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:${LRAG_WEB_PORT}/" 2>/dev/null || echo "000"
}

check_web_api() {
  curl -sS -o /tmp/lrag-api-health.body -w '%{http_code}' "http://127.0.0.1:${LRAG_WEB_PORT}/api/health" 2>/dev/null || echo "000"
}

WAIT_END=0
if [[ "$WAIT_SECS" -gt 0 ]]; then
  WAIT_END=$((SECONDS + WAIT_SECS))
  echo "==> 等待服务就绪（截止约 ${WAIT_SECS}s 内，间隔 3s）…"
  while [[ $SECONDS -lt $WAIT_END ]]; do
    code="$(check_server)"
    if [[ "$code" == "200" ]]; then
      echo "    后端已响应 /health"
      break
    fi
    sleep 3
  done
fi

echo "==> 检查服务端 GET /health (127.0.0.1:${LRAG_SERVER_PORT})"
code="$(check_server)"
if [[ "$code" != "200" ]]; then
  echo "失败: HTTP $code, body: $(cat /tmp/lrag-health.body 2>/dev/null || true)"
  exit 1
fi
echo "    HTTP $code OK"

if [[ "$API_ONLY" == true ]]; then
  echo "（--api-only）跳过 Web 与 Nginx 检查。"
  echo "全部检查通过。"
  exit 0
fi

if [[ "$API_ONLY" != true && "$WAIT_END" -gt 0 ]]; then
  echo "==> 等待 Web（同一截止时间）…"
  while [[ $SECONDS -lt $WAIT_END ]]; do
    code="$(check_web_root)"
    if [[ "$code" == "200" ]]; then
      echo "    Web 根路径已响应"
      break
    fi
    sleep 3
  done
fi

echo "==> 检查 Web 静态页 (127.0.0.1:${LRAG_WEB_PORT})"
code="$(check_web_root)"
if [[ "$code" != "200" ]]; then
  echo "失败: Web 根路径 HTTP $code"
  exit 1
fi
echo "    HTTP $code OK"

echo "==> 检查经 Nginx 反代的 API（/api/health -> 后端 /health）"
code="$(check_web_api)"
if [[ "$code" != "200" ]]; then
  echo "失败: /api/health HTTP $code, body: $(cat /tmp/lrag-api-health.body 2>/dev/null || true)"
  exit 1
fi
echo "    HTTP $code OK"

echo "全部检查通过。"
