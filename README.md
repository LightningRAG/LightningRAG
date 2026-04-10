# LightningRAG

English | [简体中文](./README_zh.md)

[GitHub](https://github.com/LightningRAG/LightningRAG): https://github.com/LightningRAG/LightningRAG

## 1. Basic Introduction

### 1.1 Project introduction

> **LightningRAG** is built around **RAG**: **knowledge bases** (ingest, parse, chunk, vector retrieval), pluggable **LLMs, embeddings, vector stores, and rerankers**, and **Agent orchestration** on a canvas (retrieval, LLM, tools, control flow)—with optional **webhook channel connectors** (Feishu, DingTalk, Slack, etc.). It ships as a **[Vue](https://vuejs.org) + [Gin](https://gin-gonic.com)** full-stack starter with JWT, dynamic routes/menus, Casbin, a form builder, and code generation.

~~[Online Demo](https://demo.LightningRAG.com): https://demo.LightningRAG.com~~

~~username：admin~~

~~password：123456~~

*Public preview is not deployed yet; remove the strikethrough and turn the demo link back on when the server is ready.*

## 2. Features

- **Knowledge bases & RAG**: Ingest documents, parse and chunk, embed into pluggable vector stores; hybrid / multi-path retrieval (vector, keyword, PageIndex, and related retriever types); conversation and streaming chat APIs with optional **`references`**; pluggable **LLMs, embeddings, rerankers**, and **`rag:`** defaults in `config.yaml` (Section 3 below; [`server/rag/README.md`](server/rag/README.md)).
- **Agent orchestration**: Visual canvas flows with nodes such as Begin, Retrieval, LLM, Message, and **Agent** (with tools) for multi-step and branching automation; extensible **tool registry** for RAG chat ([`server/rag/tools/README.md`](server/rag/tools/README.md)).
- **Channel connectors (optional)**: Bind published agents to **webhook** endpoints for Feishu, DingTalk, WeChat, Slack, Teams, and other platforms (Section 3.5 below; [`docs/THIRD_PARTY_CHANNEL_CONNECTORS.md`](docs/THIRD_PARTY_CHANNEL_CONNECTORS.md)).
- Authority management: Authority management based on `jwt` and `casbin`. 
- File upload and download: implement file upload operations based on `Qiniuyun', `Aliyun 'and `Tencent Cloud` (please develop your own application for each platform corresponding to `token` or `key` ).
- Pagination Encapsulation：The frontend uses `mixins` to encapsulate paging, and the paging method can call `mixins` .
- User management: The system administrator assigns user roles and role permissions.
- Role management: Create the main object of permission control, and then assign different API permissions and menu permissions to the role.
- Menu management: User dynamic menu configuration implementation, assigning different menus to different roles.
- API management: Different users can call different API permissions.
- Configuration management: the configuration file can be modified in the foreground (may be disabled if you run a public demo instance).
- Conditional search: Add an example of conditional search.
- Restful example: You can see sample APIs in user management module.
  - Front-end file reference: [web/src/view/superAdmin/api/api.vue](https://github.com/LightningRAG/LightningRAG/blob/master/web/src/view/superAdmin/api/api.vue).
  - Stage reference: [server/router/sys_api.go](https://github.com/LightningRAG/LightningRAG/blob/master/server/router/sys_api.go).
- Multi-login restriction: set `use-multipoint` to `true` under `system` in `config.yaml` (configure Redis accordingly; report bugs if any).
- Chunk upload: examples for chunked and large file uploads.
- Form builder: based on [@Variant Form](https://github.com/vform666/variant-form).
- Code generator: Providing backend with basic logic and simple curd code generator.

## 3. RAG, knowledge bases, and agent orchestration

This section maps **knowledge bases → retrieval → chat / agents → channel webhooks**. For authoritative detail, see **`server/rag/README.md`**, the **`rag:` block in `config.yaml`**, and the linked docs below.

### 3.1 Knowledge bases and documents

- **Admin workflows**: Create knowledge bases, upload documents, manage metadata; **share / transfer** bases (tables such as `rag_knowledge_bases`, `rag_documents`, `rag_chunks` — see [server/rag/README.md](server/rag/README.md)).
- **Indexing**: Parse to plain text, chunk, embed, and write vectors; supported formats and PDF engine options are described in [docs/DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md](docs/DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md).
- **Storage**: Chunk metadata lives in the app database; **vectors** go to the configured VectorStore (e.g. PostgreSQL + **pgvector**, **Elasticsearch** `dense_vector`).

### 3.2 Retrieval and conversation

- **Retriever kinds**: Vector, **PageIndex**, **keyword**, and related types behind the `Retriever` interface under `server/rag/`.
- **APIs**: `/rag/conversation/chat`, `chatStream` (SSE; final frame may include retrieval mode, queries, **`references`**, etc.); **`queryData`** returns structured retrieval only (no LLM, no message persistence; Casbin must allow `/rag/conversation/queryData`).
- **Request tuning**: Body fields such as `queryMode`, `chunkTopK`, `topK`, `enableRerank`, `hlKeywords` / `llKeywords`, history, `maxRagContextTokens`, `cosineThreshold`, `minRerankScore`, `includeReferences`, and more — see `server/model/rag/request/conversation.go`.
- **Global defaults (`config.yaml` → `rag:`)**: Conversation and KB **top-k / candidate pool** keys (`default-conversation-chunk-top-k`, `default-knowledge-base-retrieve-top-n`, `max-retrieve-top-n`, `max-retrieve-candidate-top-k`, …), **hybrid fusion** weights and score floors, **`default-cosine-threshold`**, and **knowledge-graph** tuning (`kg-*`). A value of **`0` usually means “use built-in defaults”** — adjust per environment.

### 3.3 LLM, embedding, rerank, and vector store configuration

- **Admin DB config**: `rag_llm_providers`, `rag_embedding_providers`, `rag_vector_store_configs` (and the admin UI where exposed).
- **User models**: End users can add keys via **user models** (`rag_user_llms`).
- **Providers**: Pluggable LLM, embedding, **rerank**, speech, TTS, OCR, CV, etc. — full list in [server/rag/README.md](server/rag/README.md). Enable reranking per request with `enableRerank` when a rerank provider is configured.

### 3.4 Agent orchestration

- **Flow editor**: Canvas nodes such as Begin, Retrieval, LLM, Message, **Agent** (with tools) for multi-step and branching flows vs. a single fixed KB chat.
- **Plans**: [docs/AGENT_IMPLEMENTATION_PLAN.md](docs/AGENT_IMPLEMENTATION_PLAN.md), [docs/AGENT_COMPONENTS_DEVELOPMENT_PLAN.md](docs/AGENT_COMPONENTS_DEVELOPMENT_PLAN.md).
- **Tools**: Extensible tool registry for RAG chat — [server/rag/tools/README.md](server/rag/tools/README.md).

### 3.5 Channel connectors (Webhook)

- **Platforms**: Feishu, DingTalk, WeChat, WeCom, Discord, Slack, Telegram, Teams, WhatsApp, LINE, etc., bound to published **Agents** via public webhook URLs; **`X-Webhook-Secret`** or vendor signatures; per-channel **`extra` JSON** in the admin UI.
- **`rag:` ops tuning**: `channel-webhook-ip-limit-per-minute` (Redis), `channel-webhook-event-retention-days`, `channel-outbound-*` for outbound retry queues.
- **Full guide**: [docs/THIRD_PARTY_CHANNEL_CONNECTORS.md](docs/THIRD_PARTY_CHANNEL_CONNECTORS.md) (distinct from OAuth login — [docs/THIRD_PARTY_OAUTH_QUICK_LOGIN.md](docs/THIRD_PARTY_OAUTH_QUICK_LOGIN.md)).

## 4. Technology selection

- Frontend: [Vue](https://vuejs.org) with [Element Plus](https://github.com/element-plus/element-plus)-style UI (Element family).
- Backend: [Gin](https://gin-gonic.com/) for REST-style APIs.
- Database: `MySQL` > 5.7 with InnoDB, via [GORM](http://gorm.cn).
- Cache: `Redis` for JWT session tracking and optional multi-login limits.
- API docs: Swagger.
- Config: [fsnotify](https://github.com/fsnotify/fsnotify) and [Viper](https://github.com/spf13/viper) for YAML.
- Logging: [zap](https://github.com/uber-go/zap).

## 5. Project Architecture

### 5.1 Project Layout

```
    ├── server
        ├── api             (api entrance)
        │   └── v1          (v1 version interface)
        ├── config          (configuration package)
        ├── core            (core document)
        ├── docs            (swagger document directory)
        ├── global          (global object)                    
        ├── initialize      (initialization)                        
        │   └── internal    (initialize internal function)                            
        ├── middleware      (middleware layer)                        
        ├── model           (model layer)                    
        │   ├── request     (input parameter structure)                        
        │   └── response    (out-of-parameter structure)                            
        ├── packfile        (static file packaging)                        
        ├── resource        (static resource folder)                        
        │   ├── excel       (excel import and export default path)                        
        │   ├── page        (form generator)                        
        │   └── template    (template)                            
        ├── router          (routing layer)                    
        ├── service         (service layer)                    
        ├── source          (source layer)                    
        └── utils           (tool kit)                    
            ├── timer       (timer interface encapsulation)                        
            └── upload      (oss interface encapsulation)  
            
    └─web            （frontend）
        ├─public        （deploy templates）
        └─src           （source code）
            ├─api       （frontend APIs）
            ├─assets	（static files）
            ├─components（components）
            ├─router	（frontend routers）
            ├─store     （vuex state management）
            ├─style     （common styles）
            ├─utils     （frontend common utilitie）
            └─view      （pages）

```

## 6. Getting started

```
- node version > v18.16.0
- golang version >= v1.22
- IDE recommendation: GoLand
```

### 6.1 Server project

Open the `server` directory in GoLand or another editor. Do not open the LightningRAG repository root for backend work.

```bash
# clone the project
git clone https://github.com/LightningRAG/LightningRAG.git

# enter the server directory
cd server

# go mod and install Go dependencies
go generate

# run
go run .
```

### 6.2 Web project

```bash
# enter the web directory
cd web

# install dependencies
npm install

# start the web app
npm run serve
```

### 6.3 Embedded web UI (optional, `go:embed`)

To ship **one binary** with the built Vue app embedded:

1. From the **repository root**:

   ```bash
   make build-server-embed-local
   ```

   or `bash scripts/build-server-with-embed.sh` (runs `yarn build`, `scripts/sync-web-dist.sh`, then `go build` in `server/`).

2. Set `system.embed-web-ui: true` in `config.yaml` (default is `false` for the usual Nginx + API split).

3. With `router-prefix` empty and embed enabled, `/api/...` is rewritten to `/...` **before Gin routing** (HTTP handler wrapper), matching `VITE_BASE_API=/api` and the Nginx `rewrite` in this repo. Plain `Engine.Use` middleware cannot fix this because Gin matches routes before global middleware runs on 404. If `router-prefix` is set, no automatic `/api` strip is applied.

See also `scripts/sync-web-dist.sh` and `server/webui/`.

#### GoReleaser (multi-platform binaries on GitHub Releases)

Official release builds use [GoReleaser](https://goreleaser.com/) with `.goreleaser.yaml` at the repo root.

- **Before each compile:** `npm install` + `npm run build` in `web/`, then `scripts/sync-web-dist.sh` — same embedding path as above (`server/webui/webdist` → `go:embed`).
- **Go module:** `go.mod` lives in `server/`; the config sets `gomod.dir: server` so module detection works from the monorepo root.
- **Targets:** `CGO_ENABLED=0` cross-builds for Linux / Windows / macOS / FreeBSD (amd64, arm64, 386 where applicable, plus Linux armv7 and Windows arm64). See `.goreleaser.yaml` for the exact matrix.
- **Artifacts:** Each archive includes the `lightningrag` binary, `config.yaml` (copied from `server/config.docker.yaml`), and the `resource/` tree from `server/resource`.
- **CI:** Pushing a `v*` tag runs `.github/workflows/goreleaser.yml` and publishes to GitHub Releases (needs `contents: write` via `GITHUB_TOKEN`).

**Local dry run (no upload):**

```bash
goreleaser release --snapshot --clean --skip=publish
```

**Publish:** create and push a semver tag, e.g. `git tag v2.9.1 && git push origin v2.9.1`.

### 6.4 Swagger API docs

#### 6.4.1 Install swag

```shell
go install github.com/swaggo/swag/cmd/swag@latest
```

#### 6.4.2 Generate API docs

```shell
cd server
swag init
```

> After running the commands above, `docs.go`, `swagger.json`, and `swagger.yaml` under `server/docs` are updated. Start the Go server and open [http://localhost:8888/swagger/index.html](http://localhost:8888/swagger/index.html) to view the Swagger UI.

### 6.5 VS Code workspace

#### 6.5.1 Development

Open `LightningRAG.code-workspace` at the repo root in VS Code. The sidebar shows three virtual folders: `backend`, `frontend`, and `root`.

#### 6.5.2 Run / debug

Use the tasks `Backend`, `Frontend`, or `Both (Backend & Frontend)`. The last one starts backend and frontend together.

#### 6.5.3 Settings

The workspace file may define `go.toolsEnvVars` for Go tools in VS Code. On machines with multiple Go versions, set `go.gopath` and `go.goroot` as needed.

```json
    "go.gopath": null,
    "go.goroot": null,
```

### 6.6 Additional documentation

- [RAG backend module (providers, vector stores, API overview)](server/rag/README.md)
- [RAG tool-calling framework](server/rag/tools/README.md)
- [Third-party chat channel connectors (Webhook)](docs/THIRD_PARTY_CHANNEL_CONNECTORS.md)
- [Third-party OAuth quick login](docs/THIRD_PARTY_OAUTH_QUICK_LOGIN.md)
- [Knowledge-base document parsing vs. Ragflow](docs/DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md)
- [Agent orchestration plan](docs/AGENT_IMPLEMENTATION_PLAN.md) · [Agent components roadmap](docs/AGENT_COMPONENTS_DEVELOPMENT_PLAN.md)
- [简体中文：渠道连接器](docs/THIRD_PARTY_CHANNEL_CONNECTORS_zh.md) · [OAuth 快捷登录](docs/THIRD_PARTY_OAUTH_QUICK_LOGIN_zh.md)

## 7. Contributing Guide

Hi! Thank you for choosing LightningRAG.

LightningRAG is a full-stack (frontend and backend separation) framework for developers, designers and product managers.

We are excited that you are interested in contributing to LightningRAG. Before submitting your contribution though, please make sure to take a moment and read through the following guidelines.

#### 7.1 Issue guidelines

- Issues are for bug reports, feature requests, and design topics only. Other topics may be closed.

- Before opening an issue, search for existing ones.

#### 7.2 Pull request guidelines

- Fork the repository to your account. Do not create branches in the upstream repo.

- Commit messages should look like `[file]: description`, e.g. `README.md: fix typo`.

- If the PR fixes a bug, describe the bug in the PR.

- Merging needs two maintainers: one approves after review, the other reviews and merges.

## 8. Contributors

Thank you for considering your contribution to LightningRAG. See the full list on [GitHub Contributors](https://github.com/LightningRAG/LightningRAG/graphs/contributors).

## 9. Notices

Please strictly comply with Apache 2.0 and retain the work attribution. To remove copyright notices you must [obtain a license](https://plugin.LightningRAG.com/license).

Unauthorized removal of copyright notices may be subject to legal liability.
