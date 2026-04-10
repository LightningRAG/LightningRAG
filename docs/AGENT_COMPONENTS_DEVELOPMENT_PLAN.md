# Agent orchestration components — development plan

**Language:** English | **中文：** [AGENT_COMPONENTS_DEVELOPMENT_PLAN_zh.md](./AGENT_COMPONENTS_DEVELOPMENT_PLAN_zh.md)

This plan extends LightningRAG’s Agent flow using upstream patterns under `references` and phased work below. Frontend and backend should evolve together.

---

## 1. Current state and goals

### 1.1 Implemented components (12)

| Component | Backend | Frontend | Role |
|-----------|---------|----------|------|
| Begin | ✅ `component/begin.go` | ✅ componentTypes.js | Entry: query / files |
| Retrieval | ✅ `component/retrieval.go` | ✅ | KB search |
| LLM | ✅ `component/llm.go` | ✅ | Generation |
| Message | ✅ `component/message.go` | ✅ | User-facing output |
| Switch | ✅ `component/switch.go` | ✅ | Conditional branch |
| Categorize | ✅ `component/categorize.go` | ✅ | Intent routing |
| Agent | ✅ `component/agent.go` | ✅ | LLM + tools |
| HTTPRequest | ✅ `component/http_request.go` | ✅ | HTTP calls |
| Iteration | ✅ `component/iteration.go` | ✅ | Iteration |
| TextProcessing | ✅ `component/text_processing.go` | ✅ | Merge/split text |
| ExecuteSQL | ✅ `component/execute_sql.go` | ✅ | SQL execution |
| DocsGenerator | ✅ `component/docs_generator.go` | ✅ | PDF/DOCX/TXT from Markdown |

### 1.2 Full component map (orchestration)

| Category | Component | Status | Notes |
|----------|-----------|--------|-------|
| **Entry** | Begin | ✅ | Workflow start |
| **Agent** | Agent | ✅ | Reasoning, tools, multi-agent |
| **Retrieval** | Retrieval | ✅ | KB retrieval |
| **Output** | Message | ✅ | Static/dynamic messages |
| **Interaction** | Await Response | ❌ planned | Pause for form input |
| **Control** | Switch | ✅ | Branches |
| **Control** | Iteration | ✅ | Split text, iterate subgraph |
| **Control** | Categorize | ✅ | LLM intent branches |
| **Tools** | Code | ❌ planned | Sandboxed Python/JS |
| **Tools** | Text Processing | ✅ | Merge/split |
| **Tools** | HTTP Request | ✅ | Remote APIs |
| **Tools** | Execute SQL | ✅ | Standalone SQL |
| **Output** | Docs Generator | ✅ | Markdown → PDF/DOCX/TXT |

**Out of scope here:** Parser, Chunker, Transformer, Indexer (dataset pipeline).

---

## 2. Architecture (summary)

### 2.1 Backend

- **Registry:** `server/agent/component/registry.go` — `Register(name, Factory)`
- **Execution:** `server/agent/canvas/run.go` — today follows `path` linearly; **branching components need dynamic `path`**
- **DSL:** `server/agent/dsl/types.go` — `components` / `path` / `globals`

### 2.2 Frontend

- **Types:** `web/src/view/rag/agent/components/flowEditor/componentTypes.js` — `COMPONENT_TYPES`
- **DSL mapping:** `dslConverter.js` — `labelMap`, `colorMap`
- **Forms:** each component needs a config panel in the flow editor

### 2.3 Branching and `path`

`buildPath()` is linear today. For Switch/Categorize, either treat `path` as a topological order and at branch nodes pick the **next component id** from output, or evolve DSL (e.g. structured branches). **Recommended:** keep a topological `path`; at branch components, select one downstream and continue.

---

## 3. Phased roadmap (detail in Chinese doc)

Phases cover control-flow (Switch, Categorize), canvas runner changes, Iteration / HTTP / Await Response, Agent+Tools, Code sandbox, frontend canvas, testing, and docs. Task tables, DSL JSON examples, and acceptance criteria are long and maintained here:

**→ [AGENT_COMPONENTS_DEVELOPMENT_PLAN_zh.md — section 三、分阶段开发计划](./AGENT_COMPONENTS_DEVELOPMENT_PLAN_zh.md#三分阶段开发计划)**

---

## 4. Layout (reference)

```
server/agent/{dsl,canvas,component,templates}
web/src/view/rag/agent/...
```

Align new files with existing `references` samples where applicable.
