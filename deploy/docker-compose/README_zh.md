# 开发环境 Docker Compose（`deploy/docker-compose`）

**English:** [README.md](./README.md)

面向本地/开发的多容器编排：从本仓库 **`web/`**、**`server/`** 构建镜像，默认拉起 **MySQL、Redis** 与前后端；**MinIO、PostgreSQL、Elasticsearch、MongoDB** 通过 **Compose Profile** 按需启用。

## 文件结构

| 文件 | 说明 |
|------|------|
| `docker-compose.yaml` | 主入口：网络、应用服务（`web`、`server`），并通过 `include` 引入下方片段。 |
| `compose.middleware.core.yaml` | 核心中间件：**MySQL**、**Redis（Valkey）**；默认始终参与 `up`。 |
| `compose.middleware.optional.yaml` | 可选中间件：**MinIO**、**PostgreSQL（pgvector）**、**Elasticsearch**、**MongoDB**；仅带对应 profile 时创建。 |

Compose 使用 **`include`** 合并上述文件；合并后**不支持跨文件 YAML 锚点**，因此各文件内各自定义了相同的 `x-logging`（`json-file`，单文件约 50MB、保留 5 个文件）。

## 快速开始

在**仓库根目录**执行：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up -d --build
```

仅校验合并结果（不启动容器）：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml config
```

## 网络与固定 IP

自定义网络子网为 **`177.7.0.0/16`**，便于开发时固定容器地址：

| 服务 | 容器内 IPv4 |
|------|-------------|
| web | 177.7.0.11 |
| server | 177.7.0.12 |
| mysql | 177.7.0.13 |
| redis | 177.7.0.14 |
| minio | 177.7.0.15 |
| postgres | 177.7.0.16 |
| elasticsearch | 177.7.0.17 |
| mongo | 177.7.0.18 |

## 默认端口（宿主机）

| 服务 | 环境变量 | 默认宿主机端口 |
|------|----------|----------------|
| Web（Nginx） | `LRAG_WEB_PORT` | 8080 |
| API（server） | `LRAG_SERVER_PORT` | 8888 |
| MySQL | `EXPOSE_MYSQL_PORT` | 13306 |
| Redis | `EXPOSE_REDIS_PORT` | 16379 |
| MinIO API | `EXPOSE_MINIO_PORT` | 19000 |
| MinIO 控制台 | `EXPOSE_MINIO_CONSOLE_PORT` | 19001 |
| PostgreSQL | `EXPOSE_POSTGRES_PORT` | 15432 |
| Elasticsearch HTTP | `EXPOSE_ES_PORT` | 19200 |
| MongoDB | `EXPOSE_MONGO_PORT` | 17017 |

## 可选中间件（Profiles）

默认 **`docker compose up -d`** 只启动 **web、server、mysql、redis**。需要其他组件时：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml --profile minio --profile pgsql up -d
```

或在 shell / `.env` 中设置（多个 profile 用**英文逗号**，不要空格）：

```env
COMPOSE_PROFILES=minio,pgsql,elasticsearch,mongo
```

| Profile | 服务 | 说明 |
|---------|------|------|
| `minio` | minio | 对象存储；控制台端口见上表。 |
| `pgsql` | postgres | PostgreSQL 16 + pgvector；`shm_size: 256mb`。 |
| `elasticsearch` | elasticsearch | 单节点、开发向关闭 xpack 安全；见下文镜像变量。 |
| `mongo` | mongo | MongoDB 7。 |

启用后，应用侧仍需在服务端配置中填写对应连接信息；本目录 Compose **仅提供容器**。

## 环境变量参考

### MySQL（核心）

| 变量 | 默认值 |
|------|--------|
| `MYSQL_ROOT_PASSWORD` | `rag@lightningrag` |
| `MYSQL_DATABASE` | `qmPlus` |
| `MYSQL_USER` | `lrag` |
| `MYSQL_PASSWORD` | `rag@lightningrag` |

健康检查使用 **root + `MYSQL_ROOT_PASSWORD`**，与业务用户/密码解耦。

### MinIO

| 变量 | 默认值 |
|------|--------|
| `MINIO_ROOT_USER` | `minioadmin` |
| `MINIO_ROOT_PASSWORD` | `minioadmin` |

### PostgreSQL

| 变量 | 默认值 |
|------|--------|
| `POSTGRES_USER` | `lrag` |
| `POSTGRES_PASSWORD` | `rag@lightningrag` |
| `POSTGRES_DB` | `lightningrag` |

### Elasticsearch

| 变量 | 说明 |
|------|------|
| `ES_IMAGE` | **可选**。完整镜像引用（含 tag），用于内网镜像站或替代仓库；设置后优先生效。 |
| `ES_STACK_VERSION` | 未设置 `ES_IMAGE` 时，仅覆盖官方镜像 tag（默认 `8.15.3`）。 |
| `ES_JAVA_HEAP_MB` | JVM 堆（MB），默认 `512`；负载高时请酌情调大并保证宿主机内存充足。 |

拉取 `docker.elastic.co` 失败时，可设置例如：

```bash
export ES_IMAGE=你的镜像仓库/elasticsearch:8.15.3
export COMPOSE_PROFILES=elasticsearch
docker compose -f deploy/docker-compose/docker-compose.yaml up -d elasticsearch
```

### MongoDB

| 变量 | 默认值 |
|------|--------|
| `MONGO_INITDB_ROOT_USERNAME` | `lrag` |
| `MONGO_INITDB_ROOT_PASSWORD` | `rag@lightningrag` |

## 启动顺序与健康检查

- **server** 在 **mysql、redis** 健康后启动。  
- **web** 在 **server** 通过 **`http://127.0.0.1:8888/health`** 后启动。  

请勿在生产环境直接使用本编排中的默认密码与 ES 关闭安全等开发向设置；生产请使用 **`deploy/docker-compose-online/`** 或 Kubernetes，见上级 [README_zh.md](../README_zh.md)。

## 数据卷与清理

卷名随 Compose 项目名变化（例如 `docker-compose_mysql`）。删除数据：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml down -v
```

**`-v` 会删除卷内数据**，请谨慎使用。

## ARM64（Apple Silicon）与 MySQL

若官方 `mysql:8.0.x` 镜像异常，可将 `compose.middleware.core.yaml` 中 `mysql` 的 `image` 换为注释中常见的替代（如 `mysql/mysql-server:8.0`），以实际镜像文档为准。

## 与 `docker-compose-online` 的差异

| 项目 | `deploy/docker-compose`（本文） | `deploy/docker-compose-online` |
|------|--------------------------------|--------------------------------|
| 应用 | 本地 `build` **web** + **server** 分离容器 | **GoReleaser 一体化镜像**（服务 `lrag-server`，无独立 web 容器） |
| 网络与中间件 | 子网 **177.7.0.0/16**、固定 IP、默认密码/宿主机端口 | **与上文中间件对齐**（同一子网、账号、端口约定；PostgreSQL 服务名 **postgres**） |
| 配置 | 使用仓库内 server 默认配置流程 | `config/config.compose-online.yaml` + `.env` 强一致 |
| 典型用途 | 本机开发、联调 | 预发/生产形态演练 |

在线栈说明见 [docker-compose-online/README_zh.md](../docker-compose-online/README_zh.md)。
