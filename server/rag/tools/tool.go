// Package tools 提供可扩展的工具调用框架，供 RAG 对话使用
// 设计原则：工具注册制、统一接口、易于扩展
package tools

import (
	"context"
	"encoding/json"
	"sync"
)

// contextKey 用于在 context 中传递用户 ID（web_search 等工具需要按用户获取配置）
type contextKey string

const userIDContextKey contextKey = "lrag_user_id"

// WithUserID 将用户 ID 注入 context，供需要用户配置的工具使用
func WithUserID(ctx context.Context, uid uint) context.Context {
	return context.WithValue(ctx, userIDContextKey, uid)
}

// UserIDFromContext 从 context 获取用户 ID，不存在返回 0
func UserIDFromContext(ctx context.Context) uint {
	if v := ctx.Value(userIDContextKey); v != nil {
		if uid, ok := v.(uint); ok {
			return uid
		}
	}
	return 0
}

// Tool 工具接口，所有工具需实现此接口
type Tool interface {
	// Name 工具唯一标识，用于 LLM 调用
	Name() string
	// Description 工具描述，供 LLM 理解何时调用
	Description() string
	// Parameters 返回 JSON Schema，定义工具参数
	Parameters() *ParameterSchema
	// Execute 执行工具，params 为 JSON 对象，返回结果字符串
	Execute(ctx context.Context, params map[string]any) (string, error)
}

// ParameterSchema JSON Schema 简化版，用于定义工具参数
type ParameterSchema struct {
	Type       string                    `json:"type"` // object
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema 属性定义
type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ToolToOpenAIFormat 将工具转为 OpenAI tools API 格式
func ToolToOpenAIFormat(t Tool) map[string]any {
	params := t.Parameters()
	if params == nil {
		params = &ParameterSchema{Type: "object", Properties: map[string]PropertySchema{}}
	}
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  params,
		},
	}
}

// Registry 工具注册中心，线程安全
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

var defaultRegistry = &Registry{tools: make(map[string]Tool)}

// Register 注册工具
func Register(t Tool) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.tools[t.Name()] = t
}

// Get 按名称获取工具
func Get(name string) Tool {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	return defaultRegistry.tools[name]
}

// List 返回所有已注册工具
func List() []Tool {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	out := make([]Tool, 0, len(defaultRegistry.tools))
	for _, t := range defaultRegistry.tools {
		out = append(out, t)
	}
	return out
}

// ToOpenAITools 将所有工具转为 OpenAI tools 数组
func ToOpenAITools() []map[string]any {
	tools := List()
	out := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		out = append(out, ToolToOpenAIFormat(t))
	}
	return out
}

// ToOpenAIToolsForNames 仅将指定名称的工具转为 OpenAI 格式；names 为空或 nil 时返回空（默认不使用任何工具）
func ToOpenAIToolsForNames(names []string) []map[string]any {
	if len(names) == 0 {
		return nil
	}
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	tools := List()
	out := make([]map[string]any, 0)
	for _, t := range tools {
		if nameSet[t.Name()] {
			out = append(out, ToolToOpenAIFormat(t))
		}
	}
	return out
}

// ToolMeta 工具元信息，用于 API 返回和前端展示
type ToolMeta struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// toolDisplayNames 默认展示名（英文）；前端按 locale 使用 rag.tools.names.* 覆盖
var toolDisplayNames = map[string]string{
	"web_search":                  "Web search",
	"requirement_analyzer":        "Requirements analysis",
	"list_all_menus":              "Menu list",
	"list_all_apis":               "API list",
	"lrag_analyze":                "LRAG analyze",
	"lrag_execute":                "LRAG execute",
	"lrag_review":                 "Code review",
	"create_menu":                 "Create menu",
	"create_api":                  "Create API",
	"query_dictionaries":          "Query dictionaries",
	"generate_dictionary_options": "Generate dictionary options",
}

// ListToolMeta 返回所有已注册工具的元信息
func ListToolMeta() []ToolMeta {
	tools := List()
	out := make([]ToolMeta, 0, len(tools))
	for _, t := range tools {
		out = append(out, ToolMeta{
			Name:        t.Name(),
			DisplayName: GetToolDisplayName(t.Name()),
			Description: t.Description(),
		})
	}
	return out
}

// GetToolDisplayName 获取工具显示名，未配置时返回 name
func GetToolDisplayName(name string) string {
	if s := toolDisplayNames[name]; s != "" {
		return s
	}
	return name
}

// ExecuteTool 执行指定工具
func ExecuteTool(ctx context.Context, name string, argsJSON string) (string, error) {
	t := Get(name)
	if t == nil {
		return "", &ErrToolNotFound{Name: name}
	}
	var params map[string]any
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
			params = map[string]any{"query": argsJSON}
		}
	}
	if params == nil {
		params = make(map[string]any)
	}
	return t.Execute(ctx, params)
}

// ErrToolNotFound 工具未找到
type ErrToolNotFound struct {
	Name string
}

func (e *ErrToolNotFound) Error() string {
	return "tool not found: " + e.Name
}
