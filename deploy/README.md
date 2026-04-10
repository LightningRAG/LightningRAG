# LightningRAG deployment

**简体中文：** [README_zh.md](./README_zh.md)

This document describes the `deploy/` layout, recommended paths, environment variables, and how to verify a deployment. Install [Docker](https://docs.docker.com/get-docker/) and [Docker Compose V2](https://docs.docker.com/compose/) (`docker compose` subcommand) first.

## Directory layout

| Path | Purpose |
|------|---------|
| `docker-compose/` | Dev-oriented multi-container stack: build images from repo `server/` and `web/`; fixed subnet `177.7.0.0/16`. Split files, profiles, and env vars: **[docker-compose/README.md](./docker-compose/README.md)**. |
| `docker-compose-online/` | Typical prod/staging: middleware + prebuilt or **local** app images; see [README.md](./docker-compose-online/README.md) in that folder. |
| `docker-compose-online/config/` | Mounted as server `config.docker.yaml` (must match DB passwords etc. in `.env`). |
| `docker-compose-online/nginx/` | Front Nginx: `/api` → `lrag-server:8888`, long timeouts, `proxy_buffering off` (helps SSE/long requests). |
| `kubernetes/` | Example Deployment / Service / ConfigMap / Ingress; apply with `kubectl apply -k deploy/kubernetes`. |
| `docker/` | Legacy all-in-one image (CentOS 7 + MySQL + Redis + Nginx + app); **not recommended** for new setups. |
| `scripts/verify-deployment.sh` | HTTP checks: `/health`, web root, `/api/health`; `--api-only`, `--wait`. |
| `../scripts/sync-web-dist.sh` | At **repo root**: copy `web/dist` → `server/webui/webdist` for `go:embed` (`make build-server-embed-local`). |
| `../scripts/build-server-with-embed.sh` | At **repo root**: web build → sync → `go build` in `server/` (see root README). |
| `scripts/check-deploy-config.sh` | Offline: compose merge, script syntax, optional `kubectl kustomize`. |
| `kubernetes/README.md` | Namespace, TLS, differences vs. Compose. |

## Compose scenarios

| Scenario | Command (from `deploy/docker-compose-online`) |
|----------|-----------------------------------------------|
| Use registry images for app | `docker compose up -d` |
| Build server + web locally | `DOCKER_BUILDKIT=1 docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d --build` |
| Middleware + backend only (smoke / save resources) | `docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d mysql redis lrag-server` |

`lrag-server` starts after MySQL and Redis are healthy; `lrag-web` starts after `lrag-server` passes **`GET /health`**, reducing 502s from Nginx starting too early.

### Optional middleware (profiles)

Set `COMPOSE_PROFILES` in `.env` with **comma-separated** profile names (no spaces), e.g.:

```env
COMPOSE_PROFILES=minio,elasticsearch
```

Equivalent to:

```bash
docker compose --profile minio --profile elasticsearch up -d
```

After enabling MinIO / PostgreSQL / Elasticsearch / Mongo, configure connections and feature flags in `config/config.compose-online.yaml` (e.g. `system.oss-type: minio`).

### Ports and overrides

- Host ports come from `.env`: `LRAG_SERVER_PORT`, `LRAG_WEB_PORT`, `EXPOSE_*`.  
- If **8888 / 8080** are taken, change `.env` and export the same values when running verify scripts.  
- Copy `docker-compose-online/compose.override.example.yaml` to **`compose.override.yaml`** in the same directory; `docker compose up -d` merges it automatically. Root **`.gitignore`** ignores `deploy/docker-compose-online/compose.override.yaml`.

### Container logs

`docker-compose-base.yaml` and related files use the **`json-file`** driver (~**50MB** per file, **5** files) to limit disk use; adjust `max-size` / `max-file` as needed.

### Elasticsearch memory

**`ES_JAVA_HEAP_MB`** in `.env` / `.env.example` feeds `ES_JAVA_OPTS` (MB). Keep it below container/host free RAM; increase for heavier vector workloads.

## Dev multi-container (fixed IPs)

From the **repository root**:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up -d --build
```

- Web **8080**, API **8888**, MySQL **13306**, Redis **16379**.  
- The stack is split into `docker-compose.yaml`, `compose.middleware.core.yaml` (MySQL/Redis), and `compose.middleware.optional.yaml` (MinIO, PostgreSQL, Elasticsearch, MongoDB via **profiles**). Full details: **[docker-compose/README.md](./docker-compose/README.md)**.  
- On **ARM64** (e.g. Apple Silicon), if MySQL misbehaves, change the `mysql` `image` in `compose.middleware.core.yaml` (see that README).

## Kubernetes

```bash
kubectl apply -k deploy/kubernetes
```

Or apply `deploy/kubernetes/server/` and `web/` separately. See **[kubernetes/README.md](./kubernetes/README.md)** for namespaces, TLS, and dependencies.

Before production:

1. Point Deployment `image` fields at your registry.  
2. Move secrets out of ConfigMap into **Secret** (JWT, DB passwords, etc.).  
3. Ingress uses `spec.ingressClassName: nginx`; install [Ingress-NGINX](https://kubernetes.github.io/ingress-nginx/) or equivalent; tune `rules.host` and TLS.  
4. Probes use `GET /health`; if `system.router-prefix` is set, probe path should be `/<prefix>/health`.  
5. Long-lived connections / large uploads: raise timeouts on Ingress (e.g. `proxy-read-timeout`) in addition to Nginx ConfigMap values.

## Config consistency

After editing `docker-compose-online/.env`, align `config/config.compose-online.yaml`:

| `.env` | Typical config keys |
|--------|---------------------|
| `MYSQL_*`, `MYSQL_ROOT_PASSWORD` | `mysql.path`, `mysql.username`, `mysql.password`, `mysql.db-name` |
| `REDIS_PASSWORD` | `redis.password`, entries in `redis-list` |
| MinIO-related | `minio.endpoint`, `access-key-*`, `bucket-name`, etc. |

## Verify scripts

From the **repository root** (omit env vars if ports match `.env`):

```bash
chmod +x deploy/scripts/verify-deployment.sh

./deploy/scripts/verify-deployment.sh
./deploy/scripts/verify-deployment.sh --api-only
./deploy/scripts/verify-deployment.sh --wait 120
LRAG_SERVER_PORT=18888 ./deploy/scripts/verify-deployment.sh --api-only
```

`--wait` polls until the backend (and web unless `--api-only`) responds, useful right after `docker compose up -d`.

### Offline check (no containers)

```bash
./deploy/scripts/check-deploy-config.sh
```

Use in CI or before push to validate compose merge, Kustomize render, and script syntax.

## Data and upgrades

- **Volumes:** online stack volumes look like `docker-compose-online_lrag_mysql_data` (name may vary). `docker compose down -v` **deletes data** — backup in production.  
- **Config upgrades:** after changing `config.compose-online.yaml` or images, `docker compose up -d` rolls affected containers.  
- **Upload size:** front Nginx sets **`client_max_body_size 100m`** (aligned with Ingress `proxy-body-size`); raise Nginx/Ingress and backend limits for larger files.

## Images and builds

- **GitHub Actions:** pushing a **`v*`** tag or publishing a Release runs **`.github/workflows/docker-publish.yml`**, pushing to **GHCR** (`ghcr.io/<owner>/<repo>/server|web`). Optional Aliyun secrets also push `lrag/server` and `lrag/web`. See [`.github/workflows/README.md`](../.github/workflows/README.md).  
- **Prebuilt images:** `registry.cn-hangzhou.aliyuncs.com/lrag/*` may be private; you can switch `.env` to GHCR or build locally.  
- **Web image:** `vite build` is memory-heavy; if Docker exits **137** / `Killed`, give Docker **≥ 8 GiB** RAM or run `cd web && pnpm install && pnpm run build` locally and adjust Dockerfile to copy `dist` only.  
- **BuildKit:** prefer `DOCKER_BUILDKIT=1` for front-end image builds.

## Troubleshooting

- **Web OK but API 502:** `docker compose logs lrag-server`; confirm Nginx `proxy_pass` targets `lrag-server:8888`.  
- **Slow first API start:** wait for MySQL init and healthchecks; increase server `healthcheck.start_period` if needed.  
- **Front build fails:** check network and `pnpm`; ensure production Dockerfile stages use `--ignore-scripts` if `patch-package` is dev-only.  
- **CORS:** set `cors.whitelist` on the server for production (see `config.compose-online.yaml`) to match the browser origin.

## All-in-one image (`deploy/docker/`)

Based on EOL CentOS 7, single container with MySQL, Redis, Nginx, and apps — legacy only; prefer `docker-compose-online` or Kubernetes.
