# docker-compose-online

**English:** [README.md](./README.md)

生产/预发风格的 Compose 栈：**MySQL、Valkey(Redis 协议)、可选 MinIO/PG/ES/Mongo** + **一体化应用**（Compose 服务名 **`lrag-server`**，容器名 **`lightningrag-server`**；[GoReleaser](https://goreleaser.com/) 镜像内置前端与 API，单端口 **8888**，无独立 web 镜像）。

**网络与默认值**与 **`deploy/docker-compose`** 中中间件对齐：自定义子网 **`177.7.0.0/16`**，应用固定 **`177.7.0.12`**（开发环境中 **`server`** 占用同地址；在线栈无 **`177.7.0.11`** Web 容器）；MySQL **`177.7.0.13`**、Redis **`177.7.0.14`**，其余见 [docker-compose/README_zh.md](../docker-compose/README_zh.md)。默认账号如 **`rag@lightningrag`**、MinIO **`minioadmin`**、PostgreSQL 库名 **`lightningrag`**、宿主机端口 **MinIO 19000/19001**、**Mongo 17017** 等与开发 compose 一致。

镜像名与仓库根目录 **`.goreleaser.yaml`** 中 `dockers_v2` 一致，例如 **`ghcr.io/lightningrag/lightningrag`**、**`docker.io/lightningrag/lightningrag`**。

完整说明、排障与 K8s 对照见上级目录 **[../README_zh.md](../README_zh.md)**。

## 常用命令

```bash
cd deploy/docker-compose-online
cp -n .env.example .env

# 预构建一体化镜像（默认）；--wait 需 Docker Compose v2.20+，会等待 healthcheck 通过
docker compose up -d --wait
```

本地从源码分别构建前后端、多容器开发环境请使用 **`deploy/docker-compose/`**（见该目录 README）。

## 本机端口冲突

宿主机 **8888** 被占用时，在 `.env` 中修改 `LRAG_SERVER_PORT`，并对照设置验证脚本的环境变量。

也可将 `compose.override.example.yaml` 复制为 **`compose.override.yaml`**：在同一目录执行默认的 `docker compose up -d` 时会自动合并。

## 验证

在**仓库根目录**执行：

```bash
./deploy/scripts/verify-deployment.sh
./deploy/scripts/verify-deployment.sh --api-only
./deploy/scripts/verify-deployment.sh --wait 120    # 最多等待 120 秒再探测
```

## 离线校验

不启动容器，检查 Compose 合并与脚本语法：

```bash
./deploy/scripts/check-deploy-config.sh
```

## 日志与磁盘

中间件与应用服务默认使用 **json-file 日志轮转**（约 50MB×5 文件），详见上级 [README_zh.md](../README_zh.md)「容器日志」一节。
