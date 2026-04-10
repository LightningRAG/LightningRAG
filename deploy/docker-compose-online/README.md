# docker-compose-online

**简体中文：** [README_zh.md](./README_zh.md)

Production-style Compose stack: **MySQL, Valkey (Redis protocol), optional MinIO/PG/ES/Mongo** plus a **single app** (Compose service **`lrag-server`**, container **`lightningrag-server`**; [GoReleaser](https://goreleaser.com/) image with embedded web UI and API on port **8888** — no separate web image).

**Networking and defaults** match **`deploy/docker-compose`** middleware: subnet **`177.7.0.0/16`**, app at **`177.7.0.12`** (same as dev **`server`**; no **`177.7.0.11`** web node here). Passwords and host ports (e.g. MinIO **19000/19001**, Mongo **17017**, Postgres service name **`postgres`**) follow the dev compose docs.

Image names match **`dockers_v2`** in **`.goreleaser.yaml`** at the repo root, e.g. **`ghcr.io/lightningrag/lightningrag`** and **`docker.io/lightningrag/lightningrag`**.

Full guide, troubleshooting, and K8s notes: parent **[../README.md](../README.md)**.

## Common commands

```bash
cd deploy/docker-compose-online
cp -n .env.example .env

# Prebuilt all-in-one image (default); --wait needs Compose v2.20+ (waits for healthchecks)
docker compose up -d --wait
```

For local multi-container builds from `server/` and `web/`, use **`deploy/docker-compose/`** (see that README).

## Host port conflicts

If **8888** is in use, change `LRAG_SERVER_PORT` in `.env` and pass the same value to the verify script.

Copy `compose.override.example.yaml` to **`compose.override.yaml`** in this directory; default `docker compose up -d` merges it.

## Verify

From the **repository root**:

```bash
./deploy/scripts/verify-deployment.sh
./deploy/scripts/verify-deployment.sh --api-only
./deploy/scripts/verify-deployment.sh --wait 120
```

## Offline check

Without starting containers:

```bash
./deploy/scripts/check-deploy-config.sh
```

## Logs and disk

Services default to **json-file** log rotation (~50MB × 5 files); see parent [README.md](../README.md) § Container logs.
