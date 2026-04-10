# Dev Docker Compose (`deploy/docker-compose`)

**简体中文：** [README_zh.md](./README_zh.md)

Multi-container stack for **local development**: builds **`web/`** and **`server/`** from this repo, starts **MySQL** and **Redis** by default, and enables **MinIO**, **PostgreSQL**, **Elasticsearch**, and **MongoDB** via **Compose profiles**.

## File layout

| File | Role |
|------|------|
| `docker-compose.yaml` | Entrypoint: network, app services (`web`, `server`), and `include` of the fragments below. |
| `compose.middleware.core.yaml` | Core middleware: **MySQL**, **Redis (Valkey)**; always part of `up` unless you scale services explicitly. |
| `compose.middleware.optional.yaml` | Optional middleware: **MinIO**, **PostgreSQL (pgvector)**, **Elasticsearch**, **MongoDB**; created only when their profiles are active. |

Compose **`include`** merges these files. **YAML anchors do not work across included files**, so each file defines the same `x-logging` (`json-file`, ~50MB per file, 5 files).

## Quick start

From the **repository root**:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up -d --build
```

Validate the merged project without starting containers:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml config
```

## Network and static IPs

Custom bridge subnet **`177.7.0.0/16`**:

| Service | IPv4 |
|---------|------|
| web | 177.7.0.11 |
| server | 177.7.0.12 |
| mysql | 177.7.0.13 |
| redis | 177.7.0.14 |
| minio | 177.7.0.15 |
| postgres | 177.7.0.16 |
| elasticsearch | 177.7.0.17 |
| mongo | 177.7.0.18 |

## Default host ports

| Service | Env var | Default host port |
|---------|---------|-------------------|
| Web (Nginx) | `LRAG_WEB_PORT` | 8080 |
| API (server) | `LRAG_SERVER_PORT` | 8888 |
| MySQL | `EXPOSE_MYSQL_PORT` | 13306 |
| Redis | `EXPOSE_REDIS_PORT` | 16379 |
| MinIO API | `EXPOSE_MINIO_PORT` | 19000 |
| MinIO console | `EXPOSE_MINIO_CONSOLE_PORT` | 19001 |
| PostgreSQL | `EXPOSE_POSTGRES_PORT` | 15432 |
| Elasticsearch HTTP | `EXPOSE_ES_PORT` | 19200 |
| MongoDB | `EXPOSE_MONGO_PORT` | 17017 |

## Optional middleware (profiles)

By default, **`docker compose up -d`** starts **web**, **server**, **mysql**, and **redis** only. Enable extras:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml --profile minio --profile pgsql up -d
```

Or set (comma-separated profile names, **no spaces**):

```env
COMPOSE_PROFILES=minio,pgsql,elasticsearch,mongo
```

| Profile | Service | Notes |
|---------|---------|--------|
| `minio` | minio | S3-compatible; console port see table above. |
| `pgsql` | postgres | PostgreSQL 16 + pgvector; `shm_size: 256mb`. |
| `elasticsearch` | elasticsearch | Single-node, dev-oriented xpack security off; see image vars below. |
| `mongo` | mongo | MongoDB 7. |

You still need server-side config for connections; this folder only runs containers.

## Environment variables

### MySQL (core)

| Variable | Default |
|----------|---------|
| `MYSQL_ROOT_PASSWORD` | `rag@lightningrag` |
| `MYSQL_DATABASE` | `qmPlus` |
| `MYSQL_USER` | `lrag` |
| `MYSQL_PASSWORD` | `rag@lightningrag` |

Healthcheck uses **root** + **`MYSQL_ROOT_PASSWORD`**, independent of app user/password.

### MinIO

| Variable | Default |
|----------|---------|
| `MINIO_ROOT_USER` | `minioadmin` |
| `MINIO_ROOT_PASSWORD` | `minioadmin` |

### PostgreSQL

| Variable | Default |
|----------|---------|
| `POSTGRES_USER` | `lrag` |
| `POSTGRES_PASSWORD` | `rag@lightningrag` |
| `POSTGRES_DB` | `lightningrag` |

### Elasticsearch

| Variable | Purpose |
|----------|---------|
| `ES_IMAGE` | Optional full image reference (including tag) for mirrors/private registries; wins over the default. |
| `ES_STACK_VERSION` | When `ES_IMAGE` is unset, overrides the tag on the Elastic image (default `8.15.3`). |
| `ES_JAVA_HEAP_MB` | JVM heap in MB (default `512`); raise for heavier workloads if RAM allows. |

If `docker.elastic.co` is unreachable:

```bash
export ES_IMAGE=your.registry.example/elasticsearch:8.15.3
export COMPOSE_PROFILES=elasticsearch
docker compose -f deploy/docker-compose/docker-compose.yaml up -d elasticsearch
```

### MongoDB

| Variable | Default |
|----------|---------|
| `MONGO_INITDB_ROOT_USERNAME` | `lrag` |
| `MONGO_INITDB_ROOT_PASSWORD` | `rag@lightningrag` |

## Startup order

- **server** waits for **mysql** and **redis** to be healthy.  
- **web** waits until **server** passes **`http://127.0.0.1:8888/health`**.  

Do not use default passwords or dev-only Elasticsearch settings in production. Prefer **`deploy/docker-compose-online/`** or Kubernetes; see parent [README.md](../README.md).

## Volumes and cleanup

Volume names depend on the Compose project name. Remove containers **and** volumes:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml down -v
```

**`-v` deletes data** in those volumes.

## ARM64 (Apple Silicon) and MySQL

If the official `mysql:8.0.x` image misbehaves, switch the `mysql` `image` in `compose.middleware.core.yaml` to a documented alternative (e.g. `mysql/mysql-server:8.0`).

## vs `docker-compose-online`

| | `deploy/docker-compose` (this doc) | `deploy/docker-compose-online` |
|--|-----------------------------------|--------------------------------|
| App images | Local `build` of server/web | Prebuilt or local override compose files |
| Config | Repo default server workflow | `config/` + `.env` must match |
| Typical use | Local dev | Staging/prod-like stacks |

See [docker-compose-online/README.md](../docker-compose-online/README.md).
