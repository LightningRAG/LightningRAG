# LightningRAG 部署说明

**English:** [README.md](./README.md)

本文说明仓库内 `deploy/` 的用途、推荐路径、环境变量与验证方式。部署前请安装 [Docker](https://docs.docker.com/get-docker/) 与 [Docker Compose V2](https://docs.docker.com/compose/)（`docker compose` 子命令）。

## 目录结构

| 路径 | 说明 |
|------|------|
| `docker-compose/` | 开发向多容器编排：从本仓库 `server/`、`web/` 构建镜像，固定子网 `177.7.0.0/16`；**文件拆分、Profile、环境变量**见 [docker-compose/README_zh.md](./docker-compose/README_zh.md)。 |
| `docker-compose-online/` | 生产/预发常用：中间件 + 预构建镜像或**本地构建**；详见该目录 [README_zh.md](./docker-compose-online/README_zh.md)。 |
| `docker-compose-online/config/` | 挂载为服务端 `config.docker.yaml`（与 `.env` 中数据库密码等**必须一致**）。 |
| `docker-compose-online/nginx/` | 前端 Nginx：`/api` → `lrag-server:8888`，已配置长超时与 `proxy_buffering off`（利于 SSE/长请求）。 |
| `kubernetes/` | 示例 Deployment / Service / ConfigMap / Ingress；可用 `kubectl apply -k deploy/kubernetes` 一次性应用。 |
| `docker/` | 一体化旧版镜像（CentOS 7 + MySQL + Redis + Nginx + 应用），**不推荐**新环境使用。 |
| `scripts/verify-deployment.sh` | HTTP 探测：`/health`、Web 根路径、`/api/health`；支持 `--api-only` 与 `--wait`。 |
| `../scripts/sync-web-dist.sh` | （仓库根目录）将 `web/dist` 同步到 `server/webui/webdist`，供 `go:embed` 单二进制发布，与 `make build-server-embed-local` 配套。 |
| `../scripts/build-server-with-embed.sh` | （仓库根目录）前端构建 → 同步 webdist → 编译 `server`（详见根目录 README 2.6）。 |
| `scripts/check-deploy-config.sh` | 离线校验：各 Compose `config`、验证脚本语法、可选 `kubectl kustomize`。 |
| `kubernetes/README_zh.md` | K8s 命名空间、TLS、与 Compose 差异等补充说明。 |

## Compose 编排对照

| 场景 | 命令（在 `deploy/docker-compose-online` 下） |
|------|-----------------------------------------------|
| 使用镜像仓库中的前后端镜像 | `docker compose up -d` |
| 本地构建前后端 | `DOCKER_BUILDKIT=1 docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d --build` |
| 仅中间件 + 后端（冒烟 / 节省资源） | `docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d mysql redis lrag-server` |

`lrag-server` 在 MySQL、Redis 健康后启动；`lrag-web` 在 `lrag-server` **通过 HTTP `/health` 健康检查**后再启动，避免 Nginx 先起而后端未就绪导致批量 502。

### 可选中间件（Profiles）

在 `.env` 中设置 `COMPOSE_PROFILES`，多个 profile 用**英文逗号**分隔（不要空格），例如：

```env
COMPOSE_PROFILES=minio,elasticsearch
```

等价于：

```bash
docker compose --profile minio --profile elasticsearch up -d
```

启用 MinIO / PostgreSQL / Elasticsearch / Mongo 后，须在 `config/config.compose-online.yaml` 中配置对应连接与业务开关（如 `system.oss-type: minio`）。

### 端口与覆盖文件

- 默认映射由 `.env` 中 `LRAG_SERVER_PORT`、`LRAG_WEB_PORT`、`EXPOSE_*` 控制。  
- 宿主机 **8888 / 8080 已被占用**时，修改 `.env` 中上述变量，并在运行验证脚本时导出相同端口。  
- 可将 `docker-compose-online/compose.override.example.yaml` 复制为 **`compose.override.yaml`**，在同目录执行 `docker compose up -d` 时会自动合并（用于本机端口或卷挂载）。仓库根目录 **`.gitignore` 已忽略** `deploy/docker-compose-online/compose.override.yaml`，避免误提交本机路径。

### 容器日志

`docker-compose-base.yaml` 与在线/本地应用编排中为各服务配置了 **`json-file` 日志驱动**（单文件约 **50MB**、保留 **5** 个文件），降低日志占满磁盘风险；可按需在 Compose 中调整 `max-size` / `max-file`。

### Elasticsearch 内存

`.env` / `.env.example` 中的 **`ES_JAVA_HEAP_MB`** 会写入 `ES_JAVA_OPTS`（单位 MB）。请保持小于容器与主机可用内存；向量检索负载高时可酌情调大。

## 方式三：开发多容器（固定 IP）

在**仓库根目录**：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up -d --build
```

- Web `8080`，API `8888`，MySQL `13306`，Redis `16379`。  
- 编排已拆分为 `docker-compose.yaml` + `compose.middleware.core.yaml`（MySQL/Redis）+ `compose.middleware.optional.yaml`（MinIO、PostgreSQL、Elasticsearch、MongoDB，按需 **profile** 启用）。完整说明、端口与变量表见 **[docker-compose/README_zh.md](./docker-compose/README_zh.md)**。  
- **ARM64**（如 Apple Silicon）若 MySQL 异常，可调整 `compose.middleware.core.yaml` 中 `mysql` 的 `image`（参见该目录 README）。

## Kubernetes

```bash
kubectl apply -k deploy/kubernetes
```

或按资源类型分批 `apply -f deploy/kubernetes/server/`、`web/`。命名空间、TLS 与依赖说明见 **[kubernetes/README_zh.md](./kubernetes/README_zh.md)**。

部署前建议：

1. 修改各 Deployment 的 `image` 为你的镜像仓库地址。  
2. 将敏感配置从 ConfigMap 迁出，使用 **Secret**（JWT、数据库密码等）。  
3. Ingress 已设置 `spec.ingressClassName: nginx`，集群需安装 [Ingress-NGINX](https://kubernetes.github.io/ingress-nginx/)（或 class 名为 `nginx` 的同类组件）；按需修改 `rules.host` 与 TLS。  
4. 探针使用 `GET /health`；若 `system.router-prefix` 非空，探针路径应为 `/<prefix>/health`。  
5. 长连接 / 大文件上传：除 Nginx ConfigMap 中的超时外，可在 Ingress 上增加 `proxy-read-timeout` 等注解。

## 配置一致性清单

修改 `docker-compose-online/.env` 后，请同步 `config/config.compose-online.yaml` 中对应项：

| .env | 配置文件中典型键 |
|------|------------------|
| `MYSQL_*`、`MYSQL_ROOT_PASSWORD` | `mysql.path` / `mysql.username` / `mysql.password` / `mysql.db-name` |
| `REDIS_PASSWORD` | `redis.password`、`redis-list` 中各项 `password` |
| MinIO 相关 | `minio.endpoint`、`access-key-*`、`bucket-name` 等 |

## 验证脚本

在**仓库根目录**执行（端口与 `.env` 一致时可省略环境变量）：

```bash
chmod +x deploy/scripts/verify-deployment.sh

./deploy/scripts/verify-deployment.sh
./deploy/scripts/verify-deployment.sh --api-only
./deploy/scripts/verify-deployment.sh --wait 120
LRAG_SERVER_PORT=18888 ./deploy/scripts/verify-deployment.sh --api-only
```

`--wait`：在总等待时间内轮询后端（及非 `--api-only` 时的 Web），适合 `docker compose up -d` 后立即执行脚本。

### 离线校验（不启动容器）

```bash
./deploy/scripts/check-deploy-config.sh
```

用于 CI 或提交前确认 Compose 合并、Kustomize 渲染与脚本语法正常。

## 数据与升级

- **数据卷**：在线栈默认命名形如 `docker-compose-online_lrag_mysql_data`（视项目目录名略有不同），`docker compose down -v` 会删除卷内数据，生产请谨慎并做好 **备份**。  
- **配置升级**：更新 `config.compose-online.yaml` 或镜像后，执行 `docker compose up -d` 即可滚动重建依赖变更的容器。  
- **上传大小**：前端 Nginx 已设置 **`client_max_body_size 100m`**（与 Ingress 示例中的 `proxy-body-size` 对齐），更大文件请同步改 Nginx / Ingress 与后端限制。

## 镜像与构建说明

- **GitHub Actions 自动构建**：推送 **`v*`** 标签或发布 Release 后，工作流 **`.github/workflows/docker-publish.yml`** 会将镜像推送到 **GHCR**（`ghcr.io/<owner>/<repo>/server|web`）。可选配置阿里云 Secrets 后同步推送至 `lrag/server`、`lrag/web`。详见 [`.github/workflows/README_zh.md`](../.github/workflows/README_zh.md)。  
- **预构建镜像**：`registry.cn-hangzhou.aliyuncs.com/lrag/*` 可能为私有仓库；也可在 `.env` 中改为上述 GHCR 地址。拉取失败时请改用 GHCR、自有镜像或「本地构建」编排。  
- **Web 镜像构建**：`vite build` 内存占用高；Docker 构建出现 `Killed` / exit **137** 时，请为 Docker 分配 **≥ 8 GiB** 内存，或在本机执行 `cd web && pnpm install && pnpm run build` 后自定义 Dockerfile 仅拷贝 `dist`。  
- **BuildKit**：推荐 `DOCKER_BUILDKIT=1` 构建前端镜像。

## 常见问题

- **Web 正常但接口 502**：`docker compose logs lrag-server`；确认 Nginx `proxy_pass` 上游为 `lrag-server:8888`。  
- **首次启动 API 较慢**：等待 MySQL 初始化与健康检查通过；可适当增大 server 的 `healthcheck.start_period`。  
- **前端构建失败**：检查网络与 `pnpm` 拉包；确认 `web/Dockerfile` 中生产依赖阶段使用 `--ignore-scripts` 避免仅 dev 存在的 `patch-package` 报错。  
- **CORS**：生产环境请在服务端配置 `cors.whitelist`（参见 `config.compose-online.yaml`），与浏览器访问域名一致。

## 一体化镜像（`deploy/docker/`）

基于已 EOL 的 CentOS 7，单容器内运行 MySQL、Redis、Nginx 与应用，仅作历史兼容；新项目请使用 `docker-compose-online` 或 Kubernetes。
