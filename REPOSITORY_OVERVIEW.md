# LightningRAG — Project overview & Docker image guide

> This document is written to be **copied or published standalone**: in-repo references use **absolute GitHub URLs** only (no relative paths).

**Repository:** [https://github.com/LightningRAG/LightningRAG](https://github.com/LightningRAG/LightningRAG)

---

## 1. About the project

**LightningRAG** is an open-source, full-stack platform for **retrieval-augmented generation (RAG)** and **AI agent orchestration**. In one codebase you get:

- **Knowledge bases** — document ingest, parsing, chunking, embeddings, and hybrid / multi-path retrieval (vector, keyword, PageIndex-style retrievers, and related types).
- **Conversational RAG** — HTTP chat and **SSE streaming** APIs, optional **references** back to retrieved chunks, and retrieval-only APIs where needed.
- **Agent orchestration** — a **visual flow canvas** (Begin, Retrieval, LLM, Message, **Agent** with tools, branching).
- **Optional webhook channels** — publish agents to Feishu, DingTalk, WeChat, Slack, Microsoft Teams, and other platforms (see [docs/THIRD_PARTY_CHANNEL_CONNECTORS.md](https://github.com/LightningRAG/LightningRAG/blob/main/docs/THIRD_PARTY_CHANNEL_CONNECTORS.md)).

The admin and API layer is a **Vue** frontend plus **Go / Gin** backend with **JWT**, **Casbin**, dynamic menus, uploads, a form builder, and codegen scaffolding. Deep RAG configuration lives under the **`rag:`** block in YAML; module details are in [server/rag/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/server/rag/README.md).

For local development without Docker images, see [README.md](https://github.com/LightningRAG/LightningRAG/blob/main/README.md).

---

## 2. Official Docker images

Release builds publish a **single all-in-one image** (GoReleaser **`dockers_v2`**, see [.goreleaser.yaml](https://github.com/LightningRAG/LightningRAG/blob/main/.goreleaser.yaml) and [Dockerfile.goreleaser](https://github.com/LightningRAG/LightningRAG/blob/main/Dockerfile.goreleaser)):

| Registry | Image |
|----------|--------|
| GitHub Container Registry | `ghcr.io/lightningrag/lightningrag` |
| Docker Hub | `docker.io/lightningrag/lightningrag` |

**Tags:** `latest` tracks the latest release; versioned tags follow the Git tag (for example `v2.9.1`). Inspect available tags on the registry you use (e.g. [Docker Hub — lightningrag/lightningrag](https://hub.docker.com/r/lightningrag/lightningrag/tags) or the GHCR UI for `ghcr.io/lightningrag/lightningrag`).

**Platforms:** multi-arch **`linux/amd64`** and **`linux/arm64`** manifests.

**What the image contains:** Alpine-based runtime with `wget`, Python 3, **PyPDF** (for default PDF parsing paths), bundled **`resource/`**, default **`config.docker.yaml`**, and a single **`lightningrag`** binary. The process listens on **port `8888`** and serves **both the REST API and the embedded web UI** when `system.embed-web-ui: true` in the active config (the recommended Compose config does this).

**Entrypoint (simplified):** runs `./server -c config.docker.yaml` from the server working directory inside the image. Override behavior by mounting your own `config.docker.yaml` at the path documented below.

---

## 3. Recommended: run with Docker Compose (app + MySQL + Redis)

The maintained **production-style** stack is **`deploy/docker-compose-online`**: **MySQL**, **Valkey** (Redis-compatible), optional profiles (**MinIO**, **PostgreSQL + pgvector**, **Elasticsearch**, **MongoDB**), plus one application service **`lrag-server`** (`lightningrag-server`) using the image above.

**Authoritative layout and env matrix:** [deploy/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/README.md) and [deploy/docker-compose-online/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose-online/README.md).

### 3.1 Quick start

From a machine with **Docker** and **Docker Compose v2** (`docker compose`):

1. Clone the repository (or copy the `deploy/docker-compose-online` directory and its `config/` subtree).
2. Enter the stack directory and create env file:

   ```bash
   cd deploy/docker-compose-online
   cp -n .env.example .env
   ```

3. (Optional) Set **`COMPOSE_PROFILES`** in `.env` for optional middleware, comma-separated with **no spaces**, e.g. `COMPOSE_PROFILES=minio,elasticsearch`.

4. Start everything (Compose **v2.20+** supports **`--wait`** on healthchecks):

   ```bash
   docker compose up -d --wait
   ```

5. Open the app at **`http://localhost:8888`** (or `http://localhost:${LRAG_SERVER_PORT}` if you changed the port).

**Default image in Compose:** [`docker-compose.yaml`](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose-online/docker-compose.yaml) uses:

```text
${LRAG_IMAGE:-ghcr.io/lightningrag/lightningrag:latest}
```

Override **`LRAG_IMAGE`** in `.env` to pin a version, for example `ghcr.io/lightningrag/lightningrag:v2.9.1`, or switch registry to `docker.io/lightningrag/lightningrag:…`.

### 3.2 Ports and configuration

| Concern | Notes |
|--------|--------|
| **Host port** | **`LRAG_SERVER_PORT`** in [`.env.example`](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose-online/.env.example) (default **8888**). Change if the port is already in use. |
| **App config** | Compose mounts [config/config.compose-online.yaml](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose-online/config/config.compose-online.yaml) to the path inside the container expected by the image. **Do not mount this file read-only:** the app may write back after first-run / InitDB flows. |
| **`system.embed-web-ui`** | Must stay **`true`** for this image in the Compose stack so UI and API share **8888**. |
| **MySQL / Redis env** | Keep [`.env`](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose-online/.env.example) and **`config.compose-online.yaml`** in sync (passwords, hosts, DB name). See the “Config consistency” table in [deploy/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/README.md#config-consistency). |
| **Optional ES heap** | Tune **`ES_JAVA_HEAP_MB`** / limits when enabling the Elasticsearch profile. |

### 3.3 Verify the deployment

From the repository root (script reads `LRAG_SERVER_PORT` from `deploy/docker-compose-online/.env` when unset):

```bash
chmod +x deploy/scripts/verify-deployment.sh
./deploy/scripts/verify-deployment.sh
./deploy/scripts/verify-deployment.sh --api-only
./deploy/scripts/verify-deployment.sh --wait 120
```

Offline validation (no containers):

```bash
./deploy/scripts/check-deploy-config.sh
```

### 3.4 Dev-oriented multi-container build (optional)

To **build** separate `server` and `web` images from source instead of using the prebuilt all-in-one image, use **`deploy/docker-compose/`** — see [deploy/docker-compose/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose/README.md).

---

## 4. Running the image with `docker run` (advanced)

The all-in-one image **does not** bundle MySQL or Redis. Use **`docker run`** only when you already have reachable databases and a complete **`config.docker.yaml`** (paths, JWT, `mysql`, `redis`, `rag:` providers, `system.embed-web-ui`, etc.).

Example shape (adjust network, volume path, and image tag):

```bash
docker run -d --name lightningrag-server \
  --network your_backend_network \
  -p 8888:8888 \
  -v /absolute/path/to/config.docker.yaml:/go/src/github.com/LightningRAG/LightningRAG/server/config.docker.yaml \
  ghcr.io/lightningrag/lightningrag:latest
```

- **`mysql.path` / `redis.addr`** in your config must resolve from **inside** the container (Docker service names if on a user-defined bridge, or host addresses).
- Health endpoint inside the container: **`GET http://127.0.0.1:8888/health`** (same as Compose healthcheck in [docker-compose.yaml](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/docker-compose-online/docker-compose.yaml)).

For most users, **Compose in §3** is simpler and less error-prone.

---

## 5. Kubernetes and legacy Docker

- **Kubernetes:** example manifests under [deploy/kubernetes/](https://github.com/LightningRAG/LightningRAG/tree/main/deploy/kubernetes) — [deploy/kubernetes/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/kubernetes/README.md).
- **Legacy all-in-one CentOS-style image:** [deploy/docker/](https://github.com/LightningRAG/LightningRAG/tree/main/deploy/docker) — **not recommended** for new deployments ([deploy/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/deploy/README.md)).

---

## 6. Documentation map

| Topic | Link |
|--------|------|
| Full README (build from source, Swagger, embed UI) | [README.md](https://github.com/LightningRAG/LightningRAG/blob/main/README.md) |
| RAG module | [server/rag/README.md](https://github.com/LightningRAG/LightningRAG/blob/main/server/rag/README.md) |
| Webhook channels | [docs/THIRD_PARTY_CHANNEL_CONNECTORS.md](https://github.com/LightningRAG/LightningRAG/blob/main/docs/THIRD_PARTY_CHANNEL_CONNECTORS.md) |
| Contributing | [CONTRIBUTING.md](https://github.com/LightningRAG/LightningRAG/blob/main/CONTRIBUTING.md) |
| Security | [SECURITY.md](https://github.com/LightningRAG/LightningRAG/blob/main/SECURITY.md) |

---

## 7. License

LightningRAG is licensed under the **Apache License 2.0**. Notices and attribution: [README.md — Notices](https://github.com/LightningRAG/LightningRAG/blob/main/README.md#9-notices).
