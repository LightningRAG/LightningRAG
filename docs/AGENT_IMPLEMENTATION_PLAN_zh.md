# LightningRAG Agent 实施计划

**English:** [AGENT_IMPLEMENTATION_PLAN.md](./AGENT_IMPLEMENTATION_PLAN.md)

参考 `references` 目录内的 Agent 参考实现，为 LightningRAG 添加 Agent 支持，实现流程编排、模板导入及组件扩展。

## 一、LightningRAG Agent 架构概览

### 1.1 核心概念

- **DSL (Domain Specific Language)**：JSON 格式描述工作流，包含 `components`、`path`、`globals`、`history`、`retrieval` 等
- **Canvas**：图执行引擎，按 `path` 顺序执行组件，支持变量引用 `{component_id@variable}`、`{sys.query}`
- **Component**：可编排节点，如 Begin、Retrieval、LLM、Message、Agent、Categorize、Switch、Iteration 等

### 1.2 DSL 结构示例

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

### 1.3 组件分类

| 类别 | 组件 | 说明 |
|------|------|------|
| 入口 | Begin | 接收用户输入、Webhook |
| 检索 | Retrieval | 知识库检索 |
| 生成 | LLM | 大模型生成 |
| 输出 | Message | 输出到用户 |
| 智能体 | Agent | LLM + Tools（检索、搜索等） |
| 控制流 | Categorize, Switch | 分支 |
| 循环 | Iteration, Loop | 迭代 |
| 工具 | TavilySearch, Google, Wikipedia 等 | 外部 API |

### 1.4 模板系统

- 模板位于 `agent/templates/*.json`，包含 `title`、`description`、`dsl`
- 用户可从模板创建 Agent，一键导入流程

---

## 二、LightningRAG 实施计划

### Phase 1：后端 DSL 引擎与核心组件（优先）

| 序号 | 任务 | 说明 |
|------|------|------|
| 1.1 | DSL 数据结构 | 定义 `agent/dsl/types.go`，解析 components/path/globals |
| 1.2 | Canvas 执行引擎 | `agent/canvas/canvas.go`，变量解析、路径执行 |
| 1.3 | 组件基类 | `agent/component/base.go`，Input/Output、变量引用 |
| 1.4 | Begin 组件 | 接收 query、files |
| 1.5 | Retrieval 组件 | 对接现有 `rag/providers/retriever` |
| 1.6 | LLM 组件 | 对接现有 `rag/providers/llm` |
| 1.7 | Message 组件 | 输出 content |
| 1.8 | 组件注册表 | `agent/component/registry.go`，按 component_name 创建实例 |

**产出**：可执行 `Begin → Retrieval → LLM → Message` 的 RAG 流程

---

### Phase 2：流程编排模板与导入

| 序号 | 任务 | 说明 |
|------|------|------|
| 2.1 | 模板存储 | `server/agent/templates/*.json`，内置 3–5 个示例 |
| 2.2 | Agent 数据模型 | `rag_agent`、`rag_agent_version` 表 |
| 2.3 | 模板列表 API | GET `/rag/agent/templates` |
| 2.4 | 从模板创建 API | POST `/rag/agent/createFromTemplate` |
| 2.5 | Agent CRUD API | 创建、列表、详情、更新、删除 |
| 2.6 | 运行 Agent API | POST `/rag/agent/run`，传入 agent_id、query |

**模板示例**：
- `retrieval_and_generate.json`：知识库检索 + 生成
- `knowledge_base_report.json`：知识库报告 Agent
- `web_search_assistant.json`：网页搜索助手（需 Phase 4 工具）

---

### Phase 3：前端画布与编排 UI

| 序号 | 任务 | 说明 |
|------|------|------|
| 3.1 | Agent 列表页 | `/rag/agent`，卡片列表、创建入口 |
| 3.2 | 模板选择弹窗 | 从模板创建时选择 |
| 3.3 | 画布编辑器 | 基于 React Flow / Vue Flow，拖拽节点、连线 |
| 3.4 | 节点配置表单 | 各组件参数配置（Begin/Retrieval/LLM/Message） |
| 3.5 | 运行与调试 | 输入 query，查看执行日志、输出 |
| 3.6 | 导入/导出 JSON | 导出 DSL、导入已有流程 |

**参考**：`references` 目录内上游项目的前端 Agent 页面（如 `web/src/pages/agent/`）

---

### Phase 4：扩展组件（逐步完善）

| 序号 | 组件 | 依赖 | 说明 |
|------|------|------|------|
| 4.1 | Agent（LLM+Tools） | Phase 1 LLM | 支持 Retrieval 作为 Tool |
| 4.2 | Categorize | Phase 1 LLM | 意图分类，多分支 |
| 4.3 | Switch | - | 条件分支 |
| 4.4 | Iteration / Loop | - | 循环迭代 |
| 4.5 | TavilySearch | 外部 API | 网页搜索 |
| 4.6 | HTTP 请求 | - | 通用 HTTP 调用 |
| 4.7 | Code 执行 | Sandbox | 代码执行（需沙箱） |

---

## 三、目录结构（规划）

```
server/
├── agent/
│   ├── dsl/
│   │   └── types.go           # DSL 数据结构
│   ├── canvas/
│   │   └── canvas.go          # 执行引擎
│   ├── component/
│   │   ├── base.go            # 组件基类
│   │   ├── registry.go        # 组件注册
│   │   ├── begin.go
│   │   ├── retrieval.go
│   │   ├── llm.go
│   │   └── message.go
│   └── templates/
│       ├── retrieval_and_generate.json
│       ├── knowledge_base_report.json
│       └── ...
├── model/rag/
│   ├── rag_agent.go
│   └── rag_agent_version.go
└── api/v1/rag/
    └── agent.go               # Agent API

web/src/view/rag/
├── agent/
│   ├── index.vue              # Agent 列表
│   ├── editor/                 # 画布编辑器
│   └── templates/              # 模板选择
```

---

## 四、执行顺序建议

1. **Phase 1**：先打通后端 DSL 执行，确保 `retrieval_and_generate` 流程可跑通
2. **Phase 2**：模板 + Agent 存储 + API，支持从模板创建、运行
3. **Phase 3**：前端画布，支持可视化编排
4. **Phase 4**：按需扩展 Categorize、Agent+Tools、外部搜索等

---

## 六、Phase 1 已完成（2025-03）

- `server/agent/dsl/types.go` - DSL 数据结构
- `server/agent/canvas/canvas.go` - 执行引擎
- `server/agent/canvas/run.go` - Run 方法
- `server/agent/component/` - Begin、Retrieval、LLM、Message 组件
- `server/agent/templates/retrieval_and_generate.json` - 示例模板
- API: `POST /rag/agent/run`、`POST /rag/agent/templates`

**下一步**：接入 RetrieverFactory（从知识库配置创建检索器），完善 Phase 2 模板与 Agent 存储。

---

## 七、参考文件

- `agent/canvas.py`（位于 `references` 内上游快照）- 执行引擎
- `agent/component/base.py` - 组件基类
- `agent/component/begin.py` - Begin
- `agent/component/llm.py` - LLM
- `agent/component/message.py` - Message
- `agent/test/dsl_examples/retrieval_and_generate.json` - DSL 示例
- `agent/templates/*.json` - 模板示例
