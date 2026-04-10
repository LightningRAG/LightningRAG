# docker-compose-online

**English:** [README.md](./README.md)

生产/预发风格的 Compose 栈：**MySQL、Valkey(Redis 协议)、可选 MinIO/PG/ES/Mongo** + **lrag-server、lrag-web**。

完整说明、排障与 K8s 对照见上级目录 **[../README_zh.md](../README_zh.md)**。

## 常用命令

```bash
cd deploy/docker-compose-online
cp -n .env.example .env

# 预构建镜像（默认）
docker compose up -d

# 本地构建前后端
DOCKER_BUILDKIT=1 docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d --build

# 仅后端 + 中间件（冒烟）
docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d mysql redis lrag-server
```

## 本机端口冲突

宿主机 **8888 / 8080** 被占用时，优先在 `.env` 中修改 `LRAG_SERVER_PORT`、`LRAG_WEB_PORT`，并对照设置验证脚本的环境变量。

也可将 `compose.override.example.yaml` 复制为 **`compose.override.yaml`**：在同一目录执行默认的 `docker compose up -d` 时会自动合并。若使用多条 `-f`（如 `docker-compose-base.yaml` + `docker-compose.local.yaml`），需自行追加 `-f compose.override.yaml` 才会加载覆盖文件。

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
