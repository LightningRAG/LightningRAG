package component

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	Register("MCP", NewMCP)
}

// MCP MCP 工具调用组件，调用 MCP（Model Context Protocol）服务器上的工具
type MCP struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewMCP 创建 MCP 组件
func NewMCP(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &MCP{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (m *MCP) ComponentName() string {
	return "MCP"
}

// Invoke 执行 MCP 工具调用
func (m *MCP) Invoke(inputs map[string]any) error {
	m.mu.Lock()
	m.err = ""
	m.mu.Unlock()

	toolName := m.canvas.ResolveString(getStrParam(m.params, "tool_name"))
	if toolName == "" {
		toolName = getStrParam(m.params, "tool_name")
	}
	if toolName == "" {
		m.mu.Lock()
		m.err = "MCP tool_name 为空"
		m.mu.Unlock()
		return fmt.Errorf("MCP tool_name 为空")
	}

	serverURL := m.canvas.ResolveString(getStrParam(m.params, "server_url"))
	if serverURL == "" {
		serverURL = getStrParam(m.params, "server_url")
	}
	if serverURL == "" {
		// 使用内置 MCP 服务
		serverURL = fmt.Sprintf("http://127.0.0.1:%d%s", global.LRAG_CONFIG.System.Addr, global.LRAG_CONFIG.MCP.SSEPath)
	}

	serverName := m.canvas.ResolveString(getStrParam(m.params, "server_name"))
	if serverName == "" {
		serverName = getStrParam(m.params, "server_name")
	}
	if serverName == "" {
		serverName = global.LRAG_CONFIG.MCP.Name
	}

	arguments := m.resolveArguments()

	mcpClient, err := client.NewClient(serverURL, "agent-mcp", "1.0.0", serverName)
	if err != nil {
		m.mu.Lock()
		m.err = "创建 MCP 客户端失败: " + err.Error()
		m.mu.Unlock()
		return fmt.Errorf("创建 MCP 客户端失败: %w", err)
	}
	defer mcpClient.Close()

	ctx := context.Background()
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = arguments

	result, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		m.mu.Lock()
		m.err = "MCP 工具调用失败: " + err.Error()
		m.mu.Unlock()
		return fmt.Errorf("MCP 工具调用失败: %w", err)
	}

	// 提取文本结果
	var resultText string
	if len(result.Content) > 0 {
		if tc, ok := result.Content[0].(mcp.TextContent); ok {
			resultText = tc.Text
		} else {
			// 尝试 JSON 序列化
			b, _ := json.Marshal(result.Content)
			resultText = string(b)
		}
	}

	m.mu.Lock()
	m.output["result"] = resultText
	m.output["raw_content"] = result.Content
	m.mu.Unlock()
	return nil
}

// resolveArguments 解析并解析变量引用
func (m *MCP) resolveArguments() map[string]any {
	v, ok := m.params["arguments"]
	if !ok || v == nil {
		return map[string]any{}
	}
	var m0 map[string]any
	switch t := v.(type) {
	case map[string]any:
		m0 = t
	case string:
		_ = json.Unmarshal([]byte(t), &m0)
	default:
		return map[string]any{}
	}
	if m0 == nil {
		return map[string]any{}
	}
	out := make(map[string]any)
	for k, val := range m0 {
		var s string
		switch v := val.(type) {
		case string:
			s = m.canvas.ResolveString(v)
		default:
			s = m.canvas.ResolveString(fmt.Sprint(v))
		}
		out[k] = s
	}
	return out
}

// Output 获取输出
func (m *MCP) Output(key string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.output[key]
}

// OutputAll 获取所有输出
func (m *MCP) OutputAll() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range m.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (m *MCP) SetOutput(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.output[key] = value
}

// Error 返回错误
func (m *MCP) Error() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.err
}

// Reset 重置
func (m *MCP) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.output = make(map[string]any)
	m.err = ""
}
