# LightningRAG Agent implementation plan

**Language:** English | **中文：** [AGENT_IMPLEMENTATION_PLAN_zh.md](./AGENT_IMPLEMENTATION_PLAN_zh.md)

Use the `references` Agent samples to add Agent support: flow orchestration, template import, and component extensions.

## 1. Architecture overview

### 1.1 Core concepts

- **DSL**: JSON workflow with `components`, `path`, `globals`, `history`, `retrieval`
- **Canvas**: Executes `path` in order; variables `{component_id@variable}`, `{sys.query}`
- **Component**: Nodes such as Begin, Retrieval, LLM, Message, Agent, Categorize, Switch, Iteration, …

### 1.2 DSL example

(Same JSON structure as Chinese doc.)

```json
{
  "components": {
    "begin": {
      "obj": { "component_name": "Begin", "params": { "prologue": "Hi!" } },
      "downstream": ["retrieval_0"],
      "upstream": []
    },
    "retrieval_0": {
      "obj": { "component_name": "Retrieval", "params": { "kb_ids": ["xxx"], "top_n": 6 } },
      "downstream": ["generate_0"],
      "upstream": ["begin"]
    },
    "generate_0": {
      "obj": { "component_name": "LLM", "params": { "llm_id": "xxx", "sys_prompt": "..." } },
      "downstream": ["message_0"],
      "upstream": ["retrieval_0"]
    },
    "message_0": {
      "obj": { "component_name": "Message", "params": { "content": ["{generate_0@content}"] } },
      "downstream": [],
      "upstream": ["generate_0"]
    }
  },
  "path": ["begin", "retrieval_0", "generate_0", "message_0"],
  "globals": { "sys.query": "", "sys.user_id": "", "sys.conversation_turns": 0, "sys.files": [] },
  "history": [],
  "retrieval": []
}
```

### 1.3 Component categories

| Category | Components | Role |
|----------|------------|------|
| Entry | Begin | User input, Webhook |
| Retrieval | Retrieval | KB search |
| Generation | LLM | Model output |
| Output | Message | User-facing reply |
| Agent | Agent | LLM + tools |
| Control | Categorize, Switch | Branching |
| Loop | Iteration, Loop | Iteration |
| Tools | TavilySearch, Google, Wikipedia, … | External APIs |

### 1.4 Templates

- `agent/templates/*.json` with `title`, `description`, `dsl`
- Users create Agents from templates

## 2. Phased plan

### Phase 1 — DSL engine & core components

| # | Task | Notes |
|---|------|------|
| 1.1 | DSL structs | `agent/dsl/types.go` |
| 1.2 | Canvas | `agent/canvas/canvas.go`, variables + path |
| 1.3 | Base component | `agent/component/base.go` |
| 1.4–1.7 | Begin, Retrieval, LLM, Message | Wire to `rag/providers/*` |
| 1.8 | Registry | `agent/component/registry.go` |

**Goal:** `Begin → Retrieval → LLM → Message` RAG flow runs.

### Phase 2 — Templates & storage

Templates under `server/agent/templates/*.json`, `rag_agent` / `rag_agent_version` tables, APIs: templates list, createFromTemplate, CRUD, `POST /rag/agent/run`.

### Phase 3 — Frontend canvas

List page, template picker, Vue Flow editor, per-node forms, run/debug, import/export DSL.

### Phase 4 — More components

Agent+Tools, Categorize, Switch, Iteration/Loop, TavilySearch, HTTP, Code (sandbox).

## 3. Planned layout

```
server/agent/{dsl,canvas,component,templates}
model/rag/{rag_agent.go,rag_agent_version.go}
api/v1/rag/agent.go
web/src/view/rag/agent/{index.vue,editor/,templates/}
```

## 4. Recommended order

Phase 1 → Phase 2 → Phase 3 → Phase 4 as needed.

## 5. Phase 1 done (2025-03)

`dsl/types.go`, `canvas/canvas.go`, `canvas/run.go`, Begin/Retrieval/LLM/Message, `retrieval_and_generate.json`, `POST /rag/agent/run`, `POST /rag/agent/templates`.

## 6. References

Upstream under `references/`: `agent/canvas.py`, `agent/component/*.py`, `agent/test/dsl_examples/`, `agent/templates/*.json`.
