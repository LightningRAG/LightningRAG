// Package interfaces 定义 RAG 核心接口，参考 langchaingo 与 references 目录内上游设计
// 支持多家模型提供商，新提供商只需实现对应接口即可接入
package interfaces

import (
	"context"
)

// MessageRole 消息角色
type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleHuman     MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

// MessageRoleTool 工具结果消息角色
const MessageRoleTool MessageRole = "tool"

// MessageContent 消息内容
type MessageContent struct {
	Role       MessageRole
	Parts      []ContentPart
	ToolCalls  []ToolCall // 仅当 Role=assistant 且模型返回工具调用时
	ToolCallID string     // 仅当 Role=tool 时，OpenAI 用 id
	ToolName   string     // 仅当 Role=tool 时，Ollama 用 tool_name
}

// ContentPart 内容片段接口
type ContentPart interface {
	// PartType 返回内容类型，如 "text", "image"
	PartType() string
}

// TextContent 文本内容
type TextContent struct {
	Text string
}

func (t TextContent) PartType() string { return "text" }

// ContentResponse 生成响应
type ContentResponse struct {
	Choices []Choice
	Usage   *Usage
}

// Choice 选项
type Choice struct {
	Content   string
	ToolCalls []ToolCall // 当模型决定调用工具时非空
}

// Usage token 使用量
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// CallOption 调用选项
type CallOption func(*CallOptions)

// WithStreamCallback 设置流式回调，启用后每收到一块内容即调用
func WithStreamCallback(cb func(string)) CallOption {
	return func(o *CallOptions) {
		o.Stream = true
		o.StreamCallback = cb
	}
}

// WithTemperature 设置温度（0-2，越高越随机）
func WithTemperature(t float32) CallOption {
	return func(o *CallOptions) { o.Temperature = t }
}

// WithTopP 设置 Top P 采样
func WithTopP(p float32) CallOption {
	return func(o *CallOptions) { o.TopP = p }
}

// WithMaxTokens 设置最大生成 token 数
func WithMaxTokens(n int) CallOption {
	return func(o *CallOptions) { o.MaxTokens = n }
}

// CallOptions 调用选项结构
type CallOptions struct {
	Model          string
	Temperature    float32
	MaxTokens      int
	TopP           float32
	Stop           []string
	Stream         bool
	StreamCallback func(string)
	// Tools 工具定义，OpenAI 格式：[]map[string]any，每个元素含 type/function/name/description/parameters
	Tools []map[string]any
	// ReasoningEffort 深度思考/推理强度，如 "low"|"medium"|"high"，仅部分模型（如 o1）支持
	ReasoningEffort string
}

// ToolCall LLM 返回的工具调用
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// WithTools 设置工具列表（用于 tool calling）
func WithTools(tools []map[string]any) CallOption {
	return func(o *CallOptions) {
		o.Tools = tools
	}
}

// WithReasoningEffort 设置深度思考/推理强度（如 "low"|"medium"|"high"），仅部分模型支持
func WithReasoningEffort(effort string) CallOption {
	return func(o *CallOptions) {
		o.ReasoningEffort = effort
	}
}

// LLM 大语言模型接口，参考 langchaingo Model 与 references 目录内 chat_model
// 新提供商只需实现此接口即可接入
type LLM interface {
	// GenerateContent 根据消息序列生成内容，支持多模态
	GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)

	// Call 简化接口：单 prompt 生成单响应（向后兼容）
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)

	// ProviderName 返回提供商名称，如 "openai", "ollama"
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string
}
