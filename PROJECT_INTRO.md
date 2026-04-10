# LightningRAG — Project overview

**LightningRAG** is a full-stack **Vue + Gin** starter with a decoupled frontend and backend, plus built-in, extensible **RAG (retrieval-augmented generation)**: knowledge bases, vector search, and integrations with many LLM and vector-store providers—suited for Q&A and document-centric features behind one admin surface.

## Tech stack (summary)

- **Frontend**: Vue (Vite)
- **Backend**: Go / Gin
- **Platform**: JWT, dynamic routes and menus, Casbin, form builder, code generation, and sample scaffolding
- **RAG**: Modular LLM / embedding / rerank / vector-store interfaces; multiple cloud and local providers (see `server/rag/`)

## Getting started

**Requirements:** **Node.js > v18.16.0**, **Go ≥ 1.22**. Full setup and run instructions:

- English: [README.md](./README.md)
- 简体中文: [README_zh.md](./README_zh.md)

**Demo:** <https://demo.LightningRAG.com> (test credentials in the main README)

## License and commercial use

**Apache 2.0.** Commercial use and attribution requirements are described in the “Important notices” section of the main README.

---

*Short intro for GitHub and similar contexts; for full documentation, deployment, and contribution guidelines, use the root `README.md` / `README_zh.md`.*
