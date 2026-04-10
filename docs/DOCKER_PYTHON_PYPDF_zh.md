# Docker 镜像中的 Python 与 pypdf（PDF 解析）

## 背景

服务端在解析 PDF 时，默认会优先通过内嵌脚本 `pypdf_plain_extract.py` 调用 **pypdf**（与 Ragflow PlainParser 对齐），再视情况回退到 `ledongthuc/pdf`。该路径需要：

1. 可执行的 **Python 3**（可通过环境变量 `LIGHTNINGRAG_PYTHON` 指定解释器，未设置时一般为 `python3`）。
2. 可导入的 **pypdf** 包，且 Go 侧能定位到「含 `pypdf/__init__.py` 的上一级目录」作为包根目录（与本地开发时的 `references/pypdf` 用法一致）。容器内通常通过 **`LIGHTNINGRAG_PYPDF_SRC`** 指向系统/虚拟环境中 **site-packages** 目录。

仓库根目录下的 `references/` 在 `.gitignore` 中，**不会进入镜像**；若不在镜像里安装 pypdf 并设置 `LIGHTNINGRAG_PYPDF_SRC`，pypdf 路径会失败并依赖回退解析，行为与「完整默认链」不一致。

## 已修改的镜像定义

| 文件 | 说明 |
|------|------|
| `server/Dockerfile` | 开发/CI 常用的 server 镜像（如 `deploy/docker-compose` 构建 `server` 服务）。最终阶段基于 **Alpine 3.21**，安装 `python3`、`py3-pypdf`（Alpine **community** 源），并在构建时写入 `/etc/lightningrag-pypdf.env`。 |
| `Dockerfile.goreleaser` | **GoReleaser** 组装的在线一体化镜像（如 `ghcr.io/lightningrag/lightningrag`），与上述相同策略。 |

启动时通过 `ENTRYPOINT` 在 `exec` 启动进程前 **source** `/etc/lightningrag-pypdf.env`，为进程注入：

- `LIGHTNINGRAG_PYTHON=/usr/bin/python3`
- `LIGHTNINGRAG_PYPDF_SRC`：构建阶段用 Python 的 `site.getsitepackages()[0]` 得到的 **site-packages** 绝对路径（该目录下存在 `pypdf/` 包）

这样在任意工作目录下运行二进制，都能找到 pypdf 包根。

## 技术说明

- **为何用 `apk` 的 `py3-pypdf` 而非镜像内 `pip install`**：减少对外网 PyPI 的依赖，构建更可复现；版本与 Alpine 分支锁定，升级随基础镜像迭代。
- **为何改 `ENTRYPOINT` 为 shell + `exec`**：需在启动 `./server` 前加载环境文件，同时用 `exec` 保证 **PID 1** 仍为服务进程，便于信号与容器停止行为正常。
- **`server/Dockerfile` 基础镜像**：最终阶段由 `alpine:latest` 固定为 **`alpine:3.21`**，与 `Dockerfile.goreleaser` 一致，避免未来 Alpine 主版本漂移导致包名或 Python 小版本难以预期。

## 自建镜像或覆盖配置

- 若使用**自定义基础镜像**，请自行安装 Python 3 与 pypdf，并设置 `LIGHTNINGRAG_PYPDF_SRC` 为 **site-packages 绝对路径**（该目录下应存在 `pypdf/__init__.py`）。
- 可选：`LIGHTNINGRAG_PYTHON` 指向其他解释器（例如虚拟环境中的 `python`）。
- PDF 引擎与更多变量见主文档及 `server/service/rag/docparse/pypdfplain` 包内注释（如 `LIGHTNINGRAG_PDF_ENGINE`）。

## 未改动的镜像

- `deploy/docker/Dockerfile`（CentOS 7 一体化旧流程）未在本次调整；若仍在该镜像内运行带 PDF 的 server 二进制，需自行安装 Python/pypdf 并配置上述环境变量。

## 验证建议

构建镜像后可在容器内执行：

```sh
python3 -c "import pypdf; print(pypdf.__version__)"
```

并确认进程环境中已包含 `LIGHTNINGRAG_PYPDF_SRC`（例如 `docker exec <container> sh -c 'wget -qO- http://127.0.0.1:8888/health'` 前用 `docker exec` 查看 `env | grep LIGHTNINGRAG_`）。
