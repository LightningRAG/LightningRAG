// conversation_bridge 将 MCP 工具注册到 RAG 对话，使对话中可调用 MCP 工具
package mcpTool

import (
	"context"
	"encoding/json"

	ragtools "github.com/LightningRAG/LightningRAG/server/rag/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	for _, entry := range ListMcpTools() {
		adapter := &conversationToolAdapter{
			name:        entry.Name,
			description: entry.Description,
			inputSchema: entry.InputSchema,
		}
		ragtools.Register(adapter)
	}
}

// conversationToolAdapter 将 MCP 工具适配为 RAG Tool 接口
type conversationToolAdapter struct {
	name        string
	description string
	inputSchema mcp.ToolInputSchema
}

func (a *conversationToolAdapter) Name() string {
	return a.name
}

func (a *conversationToolAdapter) Description() string {
	return a.description
}

func (a *conversationToolAdapter) Parameters() *ragtools.ParameterSchema {
	return mcpSchemaToRagSchema(a.inputSchema)
}

func (a *conversationToolAdapter) Execute(ctx context.Context, params map[string]any) (string, error) {
	return ExecuteMcpTool(ctx, a.name, params)
}

func mcpSchemaToRagSchema(schema mcp.ToolInputSchema) *ragtools.ParameterSchema {
	b, err := json.Marshal(schema)
	if err != nil {
		return &ragtools.ParameterSchema{Type: "object", Properties: map[string]ragtools.PropertySchema{}}
	}
	var raw struct {
		Type       string                             `json:"type"`
		Properties map[string]ragtools.PropertySchema `json:"properties,omitempty"`
		Required   []string                           `json:"required,omitempty"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return &ragtools.ParameterSchema{Type: "object", Properties: map[string]ragtools.PropertySchema{}}
	}
	if raw.Type == "" {
		raw.Type = "object"
	}
	return &ragtools.ParameterSchema{
		Type:       raw.Type,
		Properties: raw.Properties,
		Required:   raw.Required,
	}
}
