// Package component 定义 Agent 流程编排的组件基类与接口
package component

import (
	"context"
	"regexp"
)

// VariableRefPattern 变量引用正则，匹配 {component_id@var}、{sys.xxx}、{iteration.current} 等
var VariableRefPattern = regexp.MustCompile(`(?i)\{* *\{([a-zA-Z:0-9_]+@[A-Za-z0-9_.-]+|sys\.[A-Za-z0-9_.]+|env\.[A-Za-z0-9_.]+|iteration\.[A-Za-z0-9_.]+)\} *\}*`)

// RetrieverFactoryOptions 画布 Retrieval 等传入的额外选项（对齐 Ragflow Agent 的 toc_enhance）
type RetrieverFactoryOptions struct {
	// PageIndexTocEnhance 对应 toc_enhance / tocEnhance；nil 表示不覆盖服务端默认（与 REST 不传 tocEnhance 一致）
	PageIndexTocEnhance *bool
}

// RetrieverFactory 根据知识库 ID 与本次期望返回条数创建检索器（finalTopN 来自工作流 Retrieval 的 top_n，便于与服务端默认候选池对齐）
type RetrieverFactory func(ctx context.Context, kbIDs []uint, userID uint, finalTopN int, opts RetrieverFactoryOptions) (Retriever, error)

// LLMResolvedConfig 从数据库解析出的 LLM 完整配置
type LLMResolvedConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
}

// LLMConfigResolver 根据 provider + modelName 从数据库查询完整 LLM 配置（含 API Key）
type LLMConfigResolver func(ctx context.Context, userID uint, provider, modelName string) (*LLMResolvedConfig, error)

// RunContext 运行上下文，用于流式与多轮对话
type RunContext interface {
	GetStreamCallback() func(string)
	GetHistory() []HistoryMessage
}

// HistoryMessage 历史消息，用于 LLM 多轮上下文
type HistoryMessage struct {
	Role    string
	Content string
}

// Canvas 画布接口，组件通过此接口访问上下文
type Canvas interface {
	GetComponent(cpnID string) (Component, bool)
	GetComponentObj(cpnID string) Component
	GetVariableValue(exp string) (any, bool)
	SetVariableValue(exp string, value any)
	IsVariableRef(exp string) bool
	GetGlobals() map[string]any
	GetTenantID() uint
	GetRetrieverFactory() RetrieverFactory
	GetLLMConfigResolver() LLMConfigResolver
	ResolveString(s string) string
	RunContext() RunContext
	// InvokeComponent 执行指定组件（供 Iteration 等循环组件调用）
	InvokeComponent(cpnID string) error
	// GetComponentDownstream 获取组件的下游 ID 列表
	GetComponentDownstream(cpnID string) []string
}

// Component 组件接口
type Component interface {
	ComponentName() string
	Invoke(inputs map[string]any) error
	Output(key string) any
	OutputAll() map[string]any
	SetOutput(key string, value any)
	Error() string
	Reset()
}

// Retriever 检索器接口，用于 Retrieval 组件
type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]Document, error)
}

// Document 检索文档
type Document struct {
	PageContent string
	Metadata    map[string]any
	Score       float32
}

// ParamBase 参数基类（组件可嵌入）
type ParamBase struct {
	MessageHistoryWindowSize int      `json:"message_history_window_size"`
	MaxRetries               int      `json:"max_retries"`
	DelayAfterError          float64  `json:"delay_after_error"`
	ExceptionMethod          string   `json:"exception_method"`
	ExceptionDefaultValue    string   `json:"exception_default_value"`
	ExceptionGoto            []string `json:"exception_goto"`
}
