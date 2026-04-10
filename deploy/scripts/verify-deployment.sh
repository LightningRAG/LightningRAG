#!/usr/bin/env bash
# 验证本机 docker-compose-online 或同类端口映射是否可用（需容器已启动）
# 一体化镜像：API 与内置 Web 均在 LRAG_SERVER_PORT（默认 8888），/api 前缀由服务端剥离。
# 用法：
#   ./deploy/scripts/verify-deployment.sh                    # /health + 根路径 + /api/health（同端口）
#   ./deploy/scripts/verify-deployment.sh --api-only         # 仅 /health
#   ./deploy/scripts/verify-deployment.sh --wait 120         # 启动后最多等待 120 秒再探测
#   ./deploy/scripts/verify-deployment.sh --wait 60 --api-only
# 覆盖端口：LRAG_SERVER_PORT=9888 ./deploy/scripts/verify-deployment.sh
# 未导出 LRAG_SERVER_PORT 时，会尝试读取 deploy/docker-compose-online/.env 中的 LRAG_SERVER_PORT=

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
      echo "环境变量: LRAG_SERVER_PORT（未设置时读取 deploy/docker-compose-online/.env，默认 8888）"
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

ONLINE_ENV="$ROOT/deploy/docker-compose-online/.env"
if [[ -z "${LRAG_SERVER_PORT:-}" ]] && [[ -f "$ONLINE_ENV" ]]; then
  line="$(grep -E '^LRAG_SERVER_PORT=' "$ONLINE_ENV" | tail -1 || true)"
  if [[ -n "$line" ]]; then
    val="${line#LRAG_SERVER_PORT=}"
    val="${val%$'\r'}"
    val="${val#"${val%%[![:space:]]*}"}"
    val="${val%"${val##*[![:space:]]}"}"
    if [[ -n "$val" ]]; then
      LRAG_SERVER_PORT="$val"
    fi
  fi
fi
LRAG_SERVER_PORT="${LRAG_SERVER_PORT:-8888}"
BASE_URL="http://127.0.0.1:${LRAG_SERVER_PORT}"

BODY_FILE="$(mktemp "${TMPDIR:-/tmp}/lrag-verify.XXXXXX")"
trap 'rm -f "$BODY_FILE"' EXIT

check_server() {
  curl -sS -o "$BODY_FILE" -w '%{http_code}' "${BASE_URL}/health" 2>/dev/null || echo "000"
}

check_web_root() {
  curl -sS -o /dev/null -w '%{http_code}' "${BASE_URL}/" 2>/dev/null || echo "000"
}

check_web_api() {
  curl -sS -o "$BODY_FILE" -w '%{http_code}' "${BASE_URL}/api/health" 2>/dev/null || echo "000"
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

echo "==> 检查 GET /health (${BASE_URL}/health)"
code="$(check_server)"
if [[ "$code" != "200" ]]; then
  echo "失败: HTTP $code, body: $(cat "$BODY_FILE" 2>/dev/null || true)"
  exit 1
fi
echo "    HTTP $code OK"

if [[ "$API_ONLY" == true ]]; then
  echo "（--api-only）跳过内置 Web 检查。"
  echo "全部检查通过。"
  exit 0
fi

if [[ "$WAIT_END" -gt 0 ]]; then
  echo "==> 等待内置 Web（同一截止时间）…"
  while [[ $SECONDS -lt $WAIT_END ]]; do
    code="$(check_web_root)"
    if [[ "$code" == "200" ]]; then
      echo "    Web 根路径已响应"
      break
    fi
    sleep 3
  done
fi

echo "==> 检查内置 Web 静态入口 (${BASE_URL}/)"
code="$(check_web_root)"
if [[ "$code" != "200" ]]; then
  echo "失败: Web 根路径 HTTP $code（请确认 config 中 system.embed-web-ui: true 且镜像为 GoReleaser 构建）"
  exit 1
fi
echo "    HTTP $code OK"

echo "==> 检查 /api/health（与前端一致的 /api 前缀，应转发至 /health）"
code="$(check_web_api)"
if [[ "$code" != "200" ]]; then
  echo "失败: /api/health HTTP $code, body: $(cat "$BODY_FILE" 2>/dev/null || true)"
  exit 1
fi
echo "    HTTP $code OK"

echo "全部检查通过。"
