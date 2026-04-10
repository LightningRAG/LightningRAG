# RAG conversation tools

**简体中文：** [README_zh.md](./README_zh.md)

Extensible tool-calling for RAG chat: the LLM may invoke tools (web search, calculators, etc.) as needed.

## Design

- **Registry:** tools self-register in `init()`; add a file to add a tool  
- **Common interface:** every tool implements `Tool`  
- **Isolation:** conversation logic does not need edits for each new tool  

## Built-in tools

| Name | Display | Description |
|------|---------|-------------|
| web_search | Web search | DuckDuckGo (no key) or Baidu (API key); configure under model admin → web search |
| requirement_analyzer | Requirements | MCP: analysis and module design |
| list_all_menus | Menus | MCP: menu tree |
| list_all_apis | APIs | MCP: API list |
| lrag_analyze | LRAG analyze | MCP: packages, modules, dictionaries |
| lrag_execute | LRAG execute | MCP: codegen run |
| lrag_review | Code review | MCP: review |
| create_menu | Create menu | MCP |
| create_api | Create API | MCP |
| query_dictionaries | Dictionaries | MCP |
| generate_dictionary_options | Dict options | MCP |

## Adding a tool

1. Add `xxx.go` under `server/rag/tools/`.  
2. Implement `Tool`: `Name()`, `Description()`, `Parameters()`, `Execute()`.  
3. In `init()`, call `Register(&YourTool{})`.  
4. Optionally add a display name in `tool.go` → `toolDisplayNames`.

Example:

```go
package tools

func init() {
    Register(&MyTool{})
}

type MyTool struct{}

func (m *MyTool) Name() string { return "my_tool" }
func (m *MyTool) Description() string { return "When the LLM should call this tool" }
func (m *MyTool) Parameters() *ParameterSchema {
    return &ParameterSchema{
        Type: "object",
        Properties: map[string]PropertySchema{
            "param1": {Type: "string", Description: "parameter help"},
        },
        Required: []string{"param1"},
    }
}
func (m *MyTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    return "result", nil
}
```

## MCP tools

MCP (Model Context Protocol) tools are wired through `server/mcp/conversation_bridge.go`. Implement and register under `server/mcp/`; display names live in `tool.go` → `toolDisplayNames`.

## Future ideas

- Per-conversation or per-user tool toggles  
- Role-based tool visibility  
- More tools: calculator, weather, DB, generic HTTP, etc.
