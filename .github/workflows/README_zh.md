# GitHub Actions 说明

**English:** [README.md](./README.md)

## CI（`ci.yaml`）

- **frontend**：Node 20 + **pnpm**，执行 `locale:check` 与生产构建（与仓库 `web/` 一致）。
- **backend**：按 `server/go.mod` 选择 Go 版本，`go vet` / `go test -short` / `go build -race`。
- **release-please**：在 `main` 分支或 `release` 事件上运行，变更日志路径为 **`docs/CHANGELOG.md`**。
- **devops-test / devops-prod**：仅在 `github.repository_owner == 'lightningrag'` 且配置了对应 Secrets 时执行 SSH 发布；fork 与其它组织不会执行。

## 自动发布 Docker 镜像（`docker-publish.yml`）

| 触发条件 | 说明 |
|----------|------|
| 推送 `v*` 标签 | 例如 `git tag v1.2.3 && git push origin v1.2.3` |
| GitHub Release **已发布** | `published` 事件，标签取自 `release.tag_name` |
| **workflow_dispatch** | 手动运行，填写 `image_tag`（如 `latest` 或 `v1.0.0`） |

### 推送到哪里

1. **GitHub Container Registry（GHCR）**（默认）  
   - 镜像：`ghcr.io/<小写 owner>/<小写 repo>/server:<tag>` 与 `.../web:<tag>`，并同时打 **`:latest`**。  
   - 使用工作流内置 `GITHUB_TOKEN`，需在仓库 **Settings → Actions → General** 中勾选 **Read and write** 权限（含 **packages**），或在组织策略中允许 `GITHUB_TOKEN` 写入 `packages`。

2. **阿里云容器镜像**（可选）  
   配置以下 Secrets 后，会在同一构建上额外推送 `lrag/server` 与 `lrag/web`（与 `deploy/docker-compose-online/.env.example` 中的路径一致）：  
   - `ALIYUN_REGISTRY`（如 `registry.cn-hangzhou.aliyuncs.com`）  
   - `ALIYUN_DOCKERHUB_USER`  
   - `ALIYUN_DOCKERHUB_PASSWORD`

### 架构

- **server**：`linux/amd64` + `linux/arm64`  
- **web**：`linux/amd64`（避免在 QEMU 下构建 arm64 前端导致超时或内存不足）

本地或 fork 验证流水线时，仅需通过 CI；发布镜像需对目标仓库具备 `packages: write` 或使用 PAT。

## 多平台二进制发布（`goreleaser.yml`）

| 触发条件 | 说明 |
|----------|------|
| 推送 `v*` 标签 | 例如 `git tag v1.2.3 && git push origin v1.2.3` |

- 使用 [GoReleaser](https://goreleaser.com/) v2，配置文件为仓库根目录 **`.goreleaser.yaml`**。
- **发布前步骤：** 在 `web/` 用 npm 构建前端，再执行 **`scripts/sync-web-dist.sh`**，将静态资源同步到 **`server/webui/webdist`**，与根目录 README「单二进制嵌入前端」一致，最后交叉编译 **`server/`** 下的 `lightningrag`。
- **发布产物：** GitHub Releases 上各平台压缩包内含 **`lightningrag`** 可执行文件、根目录 **`config.yaml`**（由 **`server/config.docker.yaml`** 复制）、以及 **`resource/`**（来自 **`server/resource`**）。
- **运行环境：** `ubuntu-latest`，**Node 20** + **`server/go.mod` 指定的 Go**；`GITHUB_TOKEN` 需具备 **`contents: write`** 以创建 Release。

本地试跑（不上传）：在仓库根目录执行 `goreleaser release --snapshot --clean --skip=publish`。详细平台列表与忽略规则见 **`.goreleaser.yaml`** 注释与配置。
