# docker-compose-online

**简体中文：** [README_zh.md](./README_zh.md)

Production-style Compose stack: **MySQL, Valkey (Redis protocol), optional MinIO/PG/ES/Mongo** plus **lrag-server** and **lrag-web**.

Full guide, troubleshooting, and K8s notes: parent **[../README.md](../README.md)**.

## Common commands

```bash
cd deploy/docker-compose-online
cp -n .env.example .env

# Prebuilt images (default)
docker compose up -d

# Build server + web locally
DOCKER_BUILDKIT=1 docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d --build

# Backend + middleware only (smoke)
docker compose -f docker-compose-base.yaml -f docker-compose.local.yaml up -d mysql redis lrag-server
```

## Host port conflicts

If **8888 / 8080** are in use, change `LRAG_SERVER_PORT` and `LRAG_WEB_PORT` in `.env` and pass the same values to the verify script env.

Copy `compose.override.example.yaml` to **`compose.override.yaml`** in this directory; default `docker compose up -d` merges it. If you use multiple `-f` files, add `-f compose.override.yaml` explicitly.

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
