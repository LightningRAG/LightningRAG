# RAG 对话工具框架

**English:** [README.md](./README.md)

本包提供可扩展的工具调用框架，供 RAG 知识库对话使用。LLM 可根据用户问题自动决定是否调用工具（如网页搜索、计算器等）。

## 设计原则

- **工具注册制**：通过 `init()` 自动注册，新增工具只需添加新文件
- **统一接口**：所有工具实现 `Tool` 接口
- **易于扩展**：新增工具时无需修改对话逻辑

## 已集成工具

| 工具名 | 显示名 | 说明 |
|--------|--------|------|
| web_search | 互联网搜索 | 接口实现，支持 DuckDuckGo（无需配置）和百度（需 API Key），在模型配置管理-互联网搜索中配置 |
| requirement_analyzer | 需求分析 | MCP 工具：智能需求分析与模块设计 |
| list_all_menus | 菜单列表 | MCP 工具：获取系统菜单信息 |
| list_all_apis | API 列表 | MCP 工具：获取系统 API 接口 |
| lrag_analyze | LRAG 分析 | MCP 工具：分析包、模块与字典 |
| lrag_execute | LRAG 执行 | MCP 工具：执行代码生成 |
| lrag_review | 代码审查 | MCP 工具：代码审查 |
| create_menu | 菜单创建 | MCP 工具：创建前端菜单 |
| create_api | API 创建 | MCP 工具：创建后端 API |
| query_dictionaries | 字典查询 | MCP 工具：查询字典 |
| generate_dictionary_options | 字典选项生成 | MCP 工具：生成字典选项 |

## 如何新增工具

1. 在 `server/rag/tools/` 下新建 `xxx.go`
2. 实现 `Tool` 接口：`Name()`, `Description()`, `Parameters()`, `Execute()`
3. 在 `init()` 中调用 `Register(&YourTool{})`
4. 在 `tool.go` 的 `toolDisplayNames` 中注册显示名（可选）

示例：

```go
package tools

func init() {
    Register(&MyTool{})
}

type MyTool struct{}

func (m *MyTool) Name() string { return "my_tool" }
func (m *MyTool) Description() string { return "工具描述，供 LLM 理解何时调用" }
func (m *MyTool) Parameters() *ParameterSchema {
    return &ParameterSchema{
        Type: "object",
        Properties: map[string]PropertySchema{
            "param1": {Type: "string", Description: "参数说明"},
        },
        Required: []string{"param1"},
    }
}
func (m *MyTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    // 实现逻辑
    return "结果", nil
}
```

## MCP 工具集成

MCP（Model Context Protocol）工具通过 `server/mcp/conversation_bridge.go` 自动注册到对话中。新增 MCP 工具时，在 `server/mcp/` 下实现并注册，对话即可使用。显示名在 `tool.go` 的 `toolDisplayNames` 中配置。

## 后续可扩展方向

- **工具开关**：按对话/用户配置启用/禁用特定工具
- **工具权限**：不同角色可见不同工具
- **更多工具**：计算器、天气、数据库查询、API 调用等
