# Agent 流程编排组件扩展开发计划

**English:** [AGENT_COMPONENTS_DEVELOPMENT_PLAN.md](./AGENT_COMPONENTS_DEVELOPMENT_PLAN.md)

参照 `references` 目录内上游编排组件设计与本文各 Phase 任务，为 LightningRAG 逐步添加 Agent 流程编排组件，实现前后端协同扩展。

---

## 一、现状与目标

### 1.1 当前已有组件（12 个）

| 组件 | 后端 | 前端 | 说明 |
|------|------|------|------|
| Begin | ✅ `component/begin.go` | ✅ componentTypes.js | 入口，接收 query/files |
| Retrieval | ✅ `component/retrieval.go` | ✅ | 知识库检索 |
| LLM | ✅ `component/llm.go` | ✅ | 大模型生成 |
| Message | ✅ `component/message.go` | ✅ | 输出到用户 |
| Switch | ✅ `component/switch.go` | ✅ | 条件分支 |
| Categorize | ✅ `component/categorize.go` | ✅ | 意图分类 |
| Agent | ✅ `component/agent.go` | ✅ | 智能体（LLM+Tools） |
| HTTPRequest | ✅ `component/http_request.go` | ✅ | HTTP API 调用 |
| Iteration | ✅ `component/iteration.go` | ✅ | 迭代处理 |
| TextProcessing | ✅ `component/text_processing.go` | ✅ | 文本合并/拆分 |
| ExecuteSQL | ✅ `component/execute_sql.go` | ✅ | 执行 SQL 查询 |
| DocsGenerator | ✅ `component/docs_generator.go` | ✅ | 文档生成（PDF/DOCX/TXT） |

### 1.2 LightningRAG 完整组件清单（编排相关）

组件能力划分与常见编排产品中的命名大致对应（细节以本文表格与本仓库实现为准）：

| 类别 | 组件 | LightningRAG | 说明 |
|------|------|--------------|------|
| **入口** | Begin | ✅ | 工作流起点 |
| **智能体** | Agent | ✅ | 推理、工具调用、多 Agent 协作 |
| **检索** | Retrieval | ✅ | 知识库检索 |
| **输出** | Message | ✅ | 静态/动态消息输出 |
| **交互** | Await Response | ❌ 待开发 | 暂停工作流，等待用户表单输入 |
| **控制流** | Switch | ✅ | 条件判断分支 |
| **控制流** | Iteration | ✅ | 拆分文本逐段迭代子流程 |
| **控制流** | Categorize | ✅ | LLM 意图分类分支 |
| **工具** | Code | ❌ 待开发 | Python/JS 沙箱执行（依赖 gVisor+Sandbox） |
| **工具** | Text Processing | ✅ | 文本合并/拆分 |
| **工具** | HTTP Request | ✅ | 调用远程服务 |
| **工具** | Execute SQL | ✅ | 执行 SQL 查询（独立组件） |
| **输出** | Docs Generator | ✅ | Markdown→PDF/DOCX/TXT 文档生成 |

**非编排类**（数据集处理，暂不纳入）：Parser、Chunker、Transformer、Indexer

---

## 二、架构约束与扩展点

### 2.1 后端扩展点

- **组件注册**：`server/agent/component/registry.go` 的 `Register(name, Factory)`
- **DSL 执行**：`server/agent/canvas/run.go` 当前按 `path` 线性执行，**控制流组件需改造 path 为动态分支**
- **DSL 结构**：`server/agent/dsl/types.go` 已支持 `components/path/globals`

### 2.2 前端扩展点

- **组件类型**：`web/src/view/rag/agent/components/flowEditor/componentTypes.js` 的 `COMPONENT_TYPES`
- **DSL 转换**：`web/src/view/rag/agent/components/flowEditor/dslConverter.js` 的 `labelMap`、`colorMap`
- **节点配置**：各组件需在 FlowEditor 的配置面板中增加对应表单

### 2.3 控制流组件的 path 改造

当前 `buildPath()` 返回线性 path，Switch/Categorize 需支持**多下游分支**：

- **方案 A**：path 改为「主路径」，遇到 Switch/Categorize 时根据条件/分类结果选择 downstream 之一继续执行
- **方案 B**：path 支持分支结构，如 `path: { "main": [...], "branch_1": [...] }`，执行时动态选择
- **推荐**：保持 path 为拓扑序，执行时遇到分支组件，根据输出选择**下一个组件 ID**，再继续按拓扑序执行后续节点

---

## 三、分阶段开发计划

### Phase 1：控制流组件（Switch、Categorize）

**目标**：支持条件分支与意图分类，实现多路径编排。

#### 1.1 Switch 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 1.1.1 | `component/switch.go` | componentTypes.js | 新增 Switch 类型 |
| 1.1.2 | 实现 Case 条件：Equals/Contains/Empty/StartsWith/EndsWith 等 | 配置面板：Case 列表、条件、逻辑 AND/OR | 参考 LightningRAG Switch |
| 1.1.3 | 输出 `next_id` 或 `selected_case` | dslConverter 支持多 downstream | 每个 Case 对应一个 downstream |
| 1.1.4 | canvas/run.go 改造 | - | 遇到 Switch 时根据 `selected_case` 选择下游，继续执行 |

**DSL 示例**：
```json
{
  "switch_0": {
    "obj": {
      "component_name": "Switch",
      "params": {
        "cases": [
          {
            "conditions": [
              { "ref": "retrieval_0@formalized_content", "op": "not_empty", "value": "" }
            ],
            "logic": "AND",
            "downstream": "generate_0"
          },
          {
            "conditions": [
              { "ref": "retrieval_0@formalized_content", "op": "is_empty", "value": "" }
            ],
            "logic": "AND",
            "downstream": "fallback_0"
          }
        ]
      }
    },
    "downstream": ["generate_0", "fallback_0"],
    "upstream": ["retrieval_0"]
  }
}
```

#### 1.2 Categorize 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 1.2.1 | `component/categorize.go` | componentTypes.js | 新增 Categorize 类型 |
| 1.2.2 | 调用 LLM 做意图分类，输出 category_name | 配置面板：Model、Categories（name/description/examples）、Input | 复用 LLM 创建逻辑 |
| 1.2.3 | 输出 `selected_category` → 映射到 downstream | 每个 Category 对应一个 downstream | 与 Switch 类似 |
| 1.2.4 | canvas/run.go | - | 遇到 Categorize 时根据 `selected_category` 选择下游 |

**DSL 示例**：
```json
{
  "categorize_0": {
    "obj": {
      "component_name": "Categorize",
      "params": {
        "input": "sys.query",
        "llm_id": "ollama@llama3.2",
        "categories": [
          { "name": "qa", "description": "知识问答", "examples": ["什么是X?"], "downstream": "retrieval_0" },
          { "name": "chat", "description": "闲聊", "examples": ["你好"], "downstream": "chat_llm_0" }
        ]
      }
    },
    "downstream": ["retrieval_0", "chat_llm_0"],
    "upstream": ["begin"]
  }
}
```

#### 1.3 Canvas 执行引擎改造

| 任务 | 文件 | 说明 |
|------|------|------|
| 1.3.1 | canvas/run.go | 将「线性 path 执行」改为「按 path 顺序，遇到分支组件时动态选择 next」 |
| 1.3.2 | 定义分支组件接口 | `BranchComponent`：`GetNextComponentID() string`，返回应执行的下游 ID |
| 1.3.3 | path 计算 | 分支组件的 downstream 可能有多个，执行时只走其中一个 |

---

### Phase 2：循环与工具组件（Iteration、HTTPRequest、AwaitResponse）

**目标**：支持迭代处理、外部 API 调用、多轮交互。

#### 2.1 Iteration 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 2.1.1 | `component/iteration.go` | componentTypes.js | 按 delimiter 拆分输入 |
| 2.1.2 | 内嵌子工作流 DSL | 画布支持「嵌套画布」或简化：单组件循环（如 LLM） | LightningRAG 有 IterationItem，我们可简化为「对每段调用同一组件」 |
| 2.1.3 | 输出合并 | 将各段结果合并为 `formalized_content` 或自定义 output | - |

**简化方案**：Iteration 只做「拆分 → 对每段调用一个 downstream 组件 → 合并结果」，不实现完整嵌套画布。

#### 2.2 HTTPRequest 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 2.2.1 | `component/http_request.go` | componentTypes.js | URL、Method、Headers、Params、Timeout |
| 2.2.2 | 支持变量引用 | 配置面板：URL/Params 支持 `{var}` | 复用 ResolveString |
| 2.2.3 | 输出 `result` | - | 响应 body 作为 string |

#### 2.3 AwaitResponse 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 2.3.1 | `component/await_response.go` | componentTypes.js | 暂停工作流，等待用户输入 |
| 2.3.2 | 对话 API 改造 | 支持「继续对话」：传入 conversation_id + 用户补充的变量 | 需扩展 Run/RunStream API |
| 2.3.3 | 前端多轮交互 | 展示 Message，收集表单变量后继续 | 与 conversation 深度集成 |

**注意**：AwaitResponse 依赖「工作流暂停/恢复」机制，实现复杂度较高，可放在 Phase 2 后期。

---

### Phase 3：Agent 组件与文本处理

**目标**：LLM+Tools 智能体、文本合并/拆分。

#### 3.1 Agent 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 3.1.1 | `component/agent.go` | componentTypes.js | 类似 LLM，但支持 Tools |
| 3.1.2 | Tools 定义 | Retrieval 作为 Tool、HTTP 作为 Tool | 需定义 Tool 接口 |
| 3.1.3 | 工具调用循环 | LLM 返回 tool_calls → 执行 Tool → 再调 LLM | 参考 LightningRAG Agent |
| 3.1.4 | 子 Agent | 可选，Agent 的 downstream 可以是另一个 Agent | 可延后 |

#### 3.2 TextProcessing 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 3.2.1 | `component/text_processing.go` | componentTypes.js | Method: merge/split |
| 3.2.2 | Merge：模板 + 变量拼接 | 配置：script 模板、delimiters | - |
| 3.2.3 | Split：按分隔符拆分 | 配置：split_ref、delimiters | - |

---

### Phase 4：Await Response 组件（多轮交互）

**目标**：支持工作流暂停，等待用户通过表单补充信息后继续执行。

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 4.1 | `component/await_response.go` | componentTypes.js | 新增 AwaitResponse 类型 |
| 4.2 | 输出 `pause_token`，存储工作流状态 | 配置面板：Message、Inputs（key/type/optional） | 变量类型：Boolean/Number/Paragraph/Single-line/Dropdown/File |
| 4.3 | Run API 扩展 | - | 新增 `ResumeRun(conversation_id, user_inputs)` 或扩展现有 API |
| 4.4 | 工作流状态持久化 | - | 暂停时保存 globals、path、current_id 到 DB/Redis |
| 4.5 | - | 多轮对话 UI | 展示 Message，渲染表单，提交后调用 Resume API |

**DSL 示例**：
```json
{
  "await_0": {
    "obj": {
      "component_name": "AwaitResponse",
      "params": {
        "message": "请补充以下信息：",
        "inputs": [
          { "key": "user_name", "type": "single_line", "optional": false },
          { "key": "preference", "type": "dropdown", "options": ["A", "B"], "optional": true }
        ]
      }
    },
    "downstream": ["llm_0"],
    "upstream": ["begin"]
  }
}
```

**依赖**：需扩展对话/会话模型，支持「暂停-恢复」语义。

---

### Phase 5：Docs Generator 与 Execute SQL

**目标**：文档生成、数据库查询能力。

#### 5.1 Docs Generator 组件

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 5.1.1 | `component/docs_generator.go` | componentTypes.js | Content、Output format（PDF/DOCX/TXT） |
| 5.1.2 | 集成 PDF 库（如 go-pdf、gofpdf） | 配置：Title、Logo、Font、Page size | 参考 LightningRAG 配置项 |
| 5.1.3 | 输出 `download`、`file_path`、`success` | Message 引用 `{docs_0@download}` 显示下载按钮 | - |

**依赖**：需引入 PDF/DOCX 生成库，注意 Unicode/CJK 字体支持。

#### 5.2 Execute SQL 组件（作为 Agent Tool 或独立组件）

| 任务 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 5.2.1 | `component/execute_sql.go` 或 Tool 实现 | componentTypes.js（若为独立组件） | SQL 语句、DB 类型、连接参数 |
| 5.2.2 | 支持 MySQL/PostgreSQL/MariaDB/MSSQL | 配置：Host、Port、Username、Password、Database | 使用 database/sql |
| 5.2.3 | 输出 `json`、`formalized_content` | - | 查询结果表格式展示 |
| 5.2.4 | Agent Tool 注册 | - | 若作为 Tool，Agent 可生成 SQL 后调用 |

---

### Phase 6：Code 组件（可选，依赖沙箱）

| 任务 | 说明 |
|------|------|
| 6.1 | 需引入代码执行沙箱（gVisor + LightningRAG Sandbox 或自建 executor） |
| 6.2 | 安全隔离、超时、资源限制、网络隔离 |
| 6.3 | 配置：Input variables、Code（Python/JS）、Output variables |
| 6.4 | 建议作为独立子项目，不纳入首期 |

---

## 四、前后端协同清单

每新增一个组件，需同步修改：

### 4.1 后端

1. `server/agent/component/<name>.go`：实现 `Component` 接口，`init()` 中 `Register`
2. `server/agent/component/registry.go`：无需改（自动通过 init 注册）
3. 若为分支组件：`canvas/run.go` 增加分支逻辑

### 4.2 前端

1. `web/src/view/rag/agent/components/flowEditor/componentTypes.js`：添加 `COMPONENT_TYPES.<Name>`
2. `web/src/view/rag/agent/components/flowEditor/dslConverter.js`：`labelMap`、`colorMap`
3. `web/src/view/rag/agent/components/flowEditor/nodeConfigPanel.vue`：若需独立配置表单，添加对应表单项
4. `web/src/view/rag/agent/components/flowEditor/componentPalette.vue`：若需新图标，在 `iconMap` 中注册
5. 节点拖拽：`COMPONENT_LIST` 由 `componentTypes.js` 自动生成，新组件会自动出现在面板

### 4.3 新增组件检查清单

每完成一个组件，确认：

- [ ] 后端：`component/<name>.go` 实现 `Component` 接口，`init()` 中 `Register`
- [ ] 后端：`params` 与 DSL 结构一致，支持变量引用 `{ref}`
- [ ] 前端：`componentTypes.js` 添加类型，含 `defaultParams`
- [ ] 前端：`dslConverter.js` 的 `labelMap`、`colorMap` 包含新组件
- [ ] 前端：`nodeConfigPanel.vue` 配置面板能正确编辑 params
- [ ] 测试：创建流程、保存、运行，验证输出正确

---

## 五、执行顺序建议

**Phase 1-3 已完成（2025-03）**：Canvas 分支改造、Switch、Categorize、HTTPRequest、Iteration、TextProcessing、Agent 均已实现。

**待开发阶段**：

| 阶段 | 组件 | 预估工作量 | 依赖 |
|------|------|------------|------|
| **Phase 4** | AwaitResponse | 3-5 天 | 对话 API 扩展、工作流状态持久化 |
| **Phase 5.1** | Docs Generator | 2-3 天 | PDF/DOCX 库、字体支持 |
| **Phase 5.2** | Execute SQL | 2-3 天 | database/sql、可选 Agent Tool 集成 |
| **Phase 6** | Code | 待定 | gVisor + Sandbox 基础设施 |

**推荐开发顺序**：
1. **Execute SQL**（Phase 5.2）— 无复杂依赖，可快速增强 Agent 数据查询能力
2. **Docs Generator**（Phase 5.1）— 独立组件，提升报告/文档生成场景
3. **Await Response**（Phase 4）— 需架构改造，多轮交互核心能力
4. **Code**（Phase 6）— 沙箱基础设施成熟后再考虑

---

## 六、参考

- 编排语义与组件参数速查见本文第七节及上文各 Phase 表格。
- 与历史实现的对照代码见仓库 `references` 目录内上游快照（本仓库不随该目录更名而修改其内容）。

---

## 七、附录：组件参数速查

| 组件 | 主要 params | 输出 |
|------|-------------|------|
| Switch | cases: [{ conditions, logic, downstream }] | selected_case, next_id |
| Categorize | input, llm_id, categories: [{ name, description, examples, downstream }] | selected_category, next_id |
| Iteration | input, delimiter, downstream | formalized_content（合并后） |
| HTTPRequest | url, method, headers, params, timeout, proxy | result |
| AwaitResponse | message, inputs: [{ key, type, optional, options? }] | 用户输入的变量（写入 globals） |
| TextProcessing | method, split_ref/script, delimiters | output |
| Agent | llm_id, sys_prompt, user_prompt, tools, creativity | content, tool_calls |
| Docs Generator | content, output_format, title, logo, font, page_size | download, file_path, success |
| Execute SQL | sql, db_type, host, port, username, password, database, max_records | json, formalized_content |
| Code | input_vars, code, output_vars, language | 自定义 output |
