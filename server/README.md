## Server layout

**简体中文：** [README_zh.md](./README_zh.md)

```text
├── api
│   └── v1
├── config
├── core
├── docs
├── global
├── initialize
│   └── internal
├── middleware
├── model
│   ├── request
│   └── response
├── packfile
├── webui
│   └── webdist
├── resource
│   ├── excel
│   ├── page
│   └── template
├── router
├── service
├── source
└── utils
    ├── timer
    └── upload
```

| Folder | Role | Notes |
|--------|------|------|
| `api` | HTTP API layer | |
| `-- v1` | API version | |
| `config` | Config structs | Maps to `config.yaml` |
| `core` | Core bootstrap | zap, viper, server init |
| `docs` | Swagger output | |
| `global` | Global singletons | |
| `initialize` | App init | router, redis, gorm, validator, timer |
| `-- internal` | Init helpers | e.g. custom gorm logger; only called from `initialize` |
| `middleware` | Gin middleware | |
| `model` | Domain models | DB tables |
| `-- request` | Request DTOs | |
| `-- response` | Response DTOs | |
| `packfile` | Embedded static assets | |
| `webui` | Embedded SPA build output | `webdist` is synced from `web/dist` (`scripts/sync-web-dist.sh`) for `go:embed`; see repo root README (embedded UI + GoReleaser releases) |
| `resource` | Static resources | |
| `-- excel` | Default Excel import/export paths | |
| `-- page` | Form generator bundled `dist` | |
| `-- template` | Code generator templates | |
| `router` | Route registration | |
| `service` | Business logic | |
| `source` | Seed / migration data loaders | |
| `utils` | Utilities | |
| `-- timer` | Timer abstractions | |
| `-- upload` | OSS helpers | |
