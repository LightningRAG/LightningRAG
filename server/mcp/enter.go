package mcpTool

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// McpTool 定义了MCP工具必须实现的接口
type McpTool interface {
	// Handle 返回工具调用信息
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	// New 返回工具注册信息
	New() mcp.Tool
}

// 工具注册表
var toolRegister = make(map[string]McpTool)

// RegisterTool 供工具在init时调用，将自己注册到工具注册表中
func RegisterTool(tool McpTool) {
	mcpTool := tool.New()
	toolRegister[mcpTool.Name] = tool
}

// RegisterAllTools 将所有注册的工具注册到MCP服务中
func RegisterAllTools(mcpServer *server.MCPServer) {
	for _, tool := range toolRegister {
		mcpServer.AddTool(tool.New(), tool.Handle)
	}
}

// McpToolEntry 供 RAG 对话等外部调用时使用的工具条目
type McpToolEntry struct {
	Name        string
	Description string
	InputSchema mcp.ToolInputSchema
	Handler     McpTool
}

// ListMcpTools 返回所有已注册的 MCP 工具
func ListMcpTools() []McpToolEntry {
	out := make([]McpToolEntry, 0, len(toolRegister))
	for name, handler := range toolRegister {
		t := handler.New()
		out = append(out, McpToolEntry{
			Name:        name,
			Description: t.Description,
			InputSchema: t.InputSchema,
			Handler:     handler,
		})
	}
	return out
}

// ExecuteMcpTool 执行指定的 MCP 工具
func ExecuteMcpTool(ctx context.Context, name string, arguments map[string]any) (string, error) {
	handler, ok := toolRegister[name]
	if !ok {
		return "", fmt.Errorf("MCP 工具不存在: %s", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = arguments
	result, err := handler.Handle(ctx, req)
	if err != nil {
		return "", err
	}
	// 将 CallToolResult 转为字符串
	if len(result.Content) == 0 {
		return "", nil
	}
	var sb strings.Builder
	for _, c := range result.Content {
		if tc, ok := mcp.AsTextContent(c); ok && tc != nil {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String(), nil
}
