# GitHub Actions

**简体中文：** [README_zh.md](./README_zh.md)

## CI (`ci.yaml`)

- **frontend:** Node 20 + **pnpm** — `locale:check` and production build (same as repo `web/`).  
- **backend:** Go version from `server/go.mod` — `go vet`, `go test -short`, `go build -race`.  
- **release-please:** runs on `main` or `release` events; changelog path is **`docs/CHANGELOG.md`**.  
- **devops-test / devops-prod:** SSH deploy only when `github.repository_owner == 'lightningrag'` and Secrets are set; forks and other orgs skip these jobs.

## Docker image publish (`docker-publish.yml`)

| Trigger | Notes |
|---------|--------|
| Push `v*` tag | e.g. `git tag v1.2.3 && git push origin v1.2.3` |
| GitHub Release **published** | tag from `release.tag_name` |
| **workflow_dispatch** | manual run with `image_tag` (e.g. `latest` or `v1.0.0`) |

### Registries

1. **GHCR (default)**  
   Images: `ghcr.io/<lowercase owner>/<lowercase repo>/server:<tag>` and `.../web:<tag>`, also **`:latest`**.  
   Enable **Read and write** for `GITHUB_TOKEN` under **Settings → Actions → General** (including **packages**), or allow org policy for `packages: write`.

2. **Aliyun Container Registry (optional)**  
   With Secrets set, also push `lrag/server` and `lrag/web` (same paths as `deploy/docker-compose-online/.env.example`):  
   - `ALIYUN_REGISTRY` (e.g. `registry.cn-hangzhou.aliyuncs.com`)  
   - `ALIYUN_DOCKERHUB_USER`  
   - `ALIYUN_DOCKERHUB_PASSWORD`

### Architectures

- **server:** `linux/amd64` + `linux/arm64`  
- **web:** `linux/amd64` only (avoids QEMU arm64 front builds timing out or OOM)

Forks only need CI to pass; publishing images requires `packages: write` or a PAT on the target repo.

## Binary releases (`goreleaser.yml`)

| Trigger | Notes |
|---------|--------|
| Push `v*` tag | e.g. `git tag v1.2.3 && git push origin v1.2.3` |

- Uses [GoReleaser](https://goreleaser.com/) v2 with repo-root `.goreleaser.yaml`.
- **Hooks:** build `web/` with npm, run `scripts/sync-web-dist.sh`, then cross-compile from `server/` with embedded UI (`go:embed` in `server/webui/`).
- **Artifacts:** per-platform archives on GitHub Releases — `lightningrag` binary, `config.yaml` (from `server/config.docker.yaml`), and `resource/` from `server/resource`.
- **Runner:** Ubuntu latest; **Node 20** + **Go** from `server/go.mod`; `GITHUB_TOKEN` with `contents: write`.

Local snapshot (no publish): `goreleaser release --snapshot --clean --skip=publish` from the repo root.
