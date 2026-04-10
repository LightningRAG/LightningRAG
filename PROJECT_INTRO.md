# LightningRAG — Project overview

**LightningRAG** centers on **RAG**, **knowledge bases**, and **Agent orchestration**: document ingestion → chunking → vector search, conversational retrieval augmented by many **LLM / embedding / rerank / vector-store** providers, and canvas-style **Agent** flows (retrieval, LLM, tools). Optional **webhook channels** bind those agents to Feishu, DingTalk, Slack, and similar apps. The stack remains a **Vue + Gin** full-stack starter with JWT, Casbin, and admin scaffolding.

**More detail:** root [README.md](./README.md) / [README_zh.md](./README_zh.md) (RAG section), [server/rag/README.md](./server/rag/README.md), and `docs/THIRD_PARTY_CHANNEL_CONNECTORS.md` (channels) / `docs/THIRD_PARTY_OAUTH_QUICK_LOGIN.md` (OAuth login).

## Tech stack (summary)

- **Frontend**: Vue (Vite)
- **Backend**: Go / Gin
- **Platform**: JWT, dynamic routes and menus, Casbin, form builder, code generation, and sample scaffolding
- **RAG**: Modular LLM / embedding / rerank / vector-store interfaces; multiple cloud and local providers (see `server/rag/`)

## Getting started

**Requirements:** **Node.js > v18.16.0**, **Go ≥ 1.22**. Full setup and run instructions:

- English: [README.md](./README.md)
- 简体中文: [README_zh.md](./README_zh.md)

**Demo (planned):** ~~<https://demo.LightningRAG.com>~~ — *Preview server not online yet; the main README will restore the link and guest credentials when it is available.*

## License and commercial use

**Apache 2.0.** Commercial use and attribution requirements are described in the **Notices** section at the end of the main README.

---

*Short intro for GitHub and similar contexts; for full documentation, deployment, and contribution guidelines, use the root `README.md` / `README_zh.md`.*
