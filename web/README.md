# LightningRAG web

**简体中文：** [README_zh.md](./README_zh.md)

CI and local development use **pnpm** (Node 20). You can still use `npm` where scripts are compatible.

## Setup

```bash
pnpm install
```

## Development (Vite, hot reload)

```bash
pnpm run serve
```

## Production build

```bash
pnpm run build
```

To embed the built UI into the Go binary (`go:embed`), from the **repository root** run `make sync-web-dist` or `bash scripts/sync-web-dist.sh` after this step, then follow the root README (“Embedded web UI” / 单二进制嵌入前端).

## Locale checks

```bash
pnpm run locale:check
pnpm run locale:zh-tw
```

## Layout (overview)

```text
web
 ├── Dockerfile
 ├── index.html
 ├── package.json
 ├── src
 │   ├── api
 │   ├── App.vue
 │   ├── assets
 │   ├── components
 │   ├── core              # site config, globals
 │   ├── directive
 │   ├── main.js
 │   ├── permission.js
 │   ├── pinia
 │   ├── router
 │   ├── style
 │   ├── utils
 │   └── view              # feature pages (login, dashboard, RAG, etc.)
 ├── vite.config.js
 └── ...
```

See [README_zh.md](./README_zh.md) for a longer annotated tree (Chinese).
