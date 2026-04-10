package request

import (
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
)

// ConversationHistoryItem 单次请求附加的多轮消息（对齐 LightningRAG QueryParam.conversation_history）
// 仅注入 LLM 上下文，不参与向量检索；格式 {"role":"user|assistant|system","content":"..."}
type ConversationHistoryItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ConversationCreate 创建对话
type ConversationCreate struct {
	Title            string   `json:"title"`
	LLMProviderID    uint     `json:"llmProviderId"`                 // 可选，不传则按优先级使用默认：用户默认 > 角色默认
	LLMSource        string   `json:"llmSource"`                     // admin|user，不传则根据默认或选择推断
	SourceType       string   `json:"sourceType" binding:"required"` // knowledge_base|files
	SourceIDs        string   `json:"sourceIds" binding:"required"`  // JSON 数组字符串，如 ["1","2"] 或 ["kb:1","kb:2","file:3"]
	EnabledToolNames []string `json:"enabledToolNames"`              // 用户选中的工具名称，空则默认不使用任何工具
}

// ConversationChat 聊天
type ConversationChat struct {
	ConversationID uint   `json:"conversationId" binding:"required"`
	Content        string `json:"content" binding:"required"`
	// QueryMode 单次请求覆盖检索模式（与 LightningRAG QueryParam.mode 一致：local|global|hybrid|mix|bypass|vector|keyword|pageindex；naive 视为 vector）；空则仅用知识库配置或消息前缀
	QueryMode string `json:"queryMode"`
	// ChunkTopK 本轮注入知识库的切片条数上限（对齐 LightningRAG QueryParam.chunk_top_k）；0 表示使用服务默认
	ChunkTopK int `json:"chunkTopK"`
	// TopK 扩大检索候选池（对齐 LightningRAG QueryParam.top_k：实体/关系规模在无 KG 时映射为「先多取再截断」）；nil 或不大于 chunkTopK 时无效
	TopK *int `json:"topK"`
	// UserPrompt 附加给模型的说明（对齐 LightningRAG QueryParam.user_prompt），追加在系统提示中
	UserPrompt string `json:"userPrompt"`
	// ResponseType 回答形式偏好（对齐 LightningRAG QueryParam.response_type），如 Multiple Paragraphs、Bullet Points；空则不追加该段
	ResponseType string `json:"responseType"`
	// OnlyNeedPrompt 为 true 时不调用 LLM，仅返回本轮将发送的完整多段消息文本（含 RAG 上下文），便于对照 LightningRAG only_need_prompt 调试
	OnlyNeedPrompt bool `json:"onlyNeedPrompt"`
	// EnableRerank 单次请求覆盖是否 Rerank（对齐 LightningRAG QueryParam.enable_rerank）；nil=按知识库；false=关闭；true=在已配置 Rerank 时启用
	EnableRerank *bool `json:"enableRerank"`
	// HlKeywords / LlKeywords 检索用关键词（对齐 LightningRAG hl_keywords / ll_keywords），拼入检索 query，不改动对用户展示的问题句
	HlKeywords []string `json:"hlKeywords"`
	LlKeywords []string `json:"llKeywords"`
	// IncludeReferences 为 false 时响应中不返回 references（对齐 LightningRAG include_references）；nil 表示返回
	IncludeReferences *bool `json:"includeReferences"`
	// IncludeChunkContent 为 false 时引用条目中不含 content 正文（对齐 LightningRAG include_chunk_content）；nil 表示保留正文以兼容旧客户端
	IncludeChunkContent *bool `json:"includeChunkContent"`
	// CosineThreshold 向量相似度下限（约 0~1），仅作用于向量 / hybrid / mix 等含向量检索的路径
	CosineThreshold *float32 `json:"cosineThreshold"`
	// MinRerankScore Rerank 后分数下限，仅在启用 Rerank 时有效
	MinRerankScore *float32 `json:"minRerankScore"`
	// ConversationHistory 额外多轮上下文，仅送入模型、不参与检索（对齐 LightningRAG）；按顺序接在服务端已加载历史之后
	ConversationHistory []ConversationHistoryItem `json:"conversationHistory"`
	// MaxTotalTokens 本轮组装消息的总 token 预算（粗估），与模型上下文上限取较小值生效；0 或未传则仅用模型配置
	MaxTotalTokens *uint `json:"maxTotalTokens"`
	// MaxRagContextTokens 注入的「知识库切片正文」token 粗算上限（对齐 LightningRAG 对 chunk 上下文的预算）；0 或未传则可用服务端 rag.default-max-rag-context-tokens
	MaxRagContextTokens *uint `json:"maxRagContextTokens"`
	// MaxEntityTokens 注入 prompt 的「图谱实体摘要」token 粗算上限（对齐 LightRAG max_entity_tokens）；nil 或 0 不注入实体块
	MaxEntityTokens *uint `json:"maxEntityTokens"`
	// MaxRelationTokens 注入 prompt 的「图谱关系摘要」token 粗算上限（对齐 LightRAG max_relation_tokens）；nil 或 0 不注入关系块
	MaxRelationTokens *uint `json:"maxRelationTokens"`
	// ResponseLanguage 回答语言偏好（对齐 LightningRAG addon_params.summary_language 类用法），如 Chinese、English；空则不追加
	ResponseLanguage string `json:"responseLanguage"`
	// 临时覆盖模型：若传入则本次对话使用指定模型，不传则使用对话创建时的模型
	LLMProviderID   uint   `json:"llmProviderId"`
	LLMSource       string `json:"llmSource"`       // admin|user
	UseDeepThinking bool   `json:"useDeepThinking"` // 是否启用深度思考（仅当模型支持时有效）
	// TocEnhance 对齐 Ragflow toc_enhance：PageIndex 是否走「向量+TOC」混合；nil=默认启用；false=纯 TOC/树
	TocEnhance *bool `json:"tocEnhance,omitempty"`
	// TocEnhanceRagflow 与 Ragflow SDK / HTTP 示例中的 toc_enhance 同义；与 TocEnhance 同时传时以 TocEnhance 为准
	TocEnhanceRagflow *bool `json:"toc_enhance,omitempty"`
}

// ConversationQueryData 对话上下文的纯检索接口（对齐 references/LightRAG POST /query/data：无 LLM 生成，返回 chunks + references + metadata）
type ConversationQueryData struct {
	ConversationID  uint     `json:"conversationId" binding:"required"`
	Query           string   `json:"query" binding:"required"`
	QueryMode       string   `json:"queryMode"`
	ChunkTopK       int      `json:"chunkTopK"`
	TopK            *int     `json:"topK"`
	EnableRerank    *bool    `json:"enableRerank"`
	HlKeywords      []string `json:"hlKeywords"`
	LlKeywords      []string `json:"llKeywords"`
	CosineThreshold *float32 `json:"cosineThreshold"`
	MinRerankScore  *float32 `json:"minRerankScore"`
	// MaxRagContextTokens 0 或未传时可由服务端 default-max-rag-context-tokens 兜底
	MaxRagContextTokens *uint                     `json:"maxRagContextTokens"`
	MaxEntityTokens     *uint                     `json:"maxEntityTokens"`
	MaxRelationTokens   *uint                     `json:"maxRelationTokens"`
	ConversationHistory []ConversationHistoryItem `json:"conversationHistory"`
	IncludeReferences   *bool                     `json:"includeReferences"`
	IncludeChunkContent *bool                     `json:"includeChunkContent"`
	LLMProviderID       uint                      `json:"llmProviderId"`
	LLMSource           string                    `json:"llmSource"`
	// TocEnhance 对齐 Ragflow：PageIndex 向量+TOC；nil=默认；false=关闭混合
	TocEnhance *bool `json:"tocEnhance,omitempty"`
	// TocEnhanceRagflow 同义字段 toc_enhance（Ragflow SDK）
	TocEnhanceRagflow *bool `json:"toc_enhance,omitempty"`
}

// ConversationList 对话列表
type ConversationList struct {
	request.PageInfo
}

// ConversationGet 获取对话
type ConversationGet struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// ConversationUpdate 更新对话（如修改启用的工具）
type ConversationUpdate struct {
	ID               uint     `json:"id" binding:"required"`
	EnabledToolNames []string `json:"enabledToolNames"` // 用户选中的工具名称，空则默认不使用任何工具
}

// ConversationDelete 删除对话
type ConversationDelete struct {
	ID uint `json:"id" binding:"required"`
}

// ConversationMessageList 获取对话消息列表
type ConversationMessageList struct {
	ConversationID uint `json:"conversationId" binding:"required"`
	request.PageInfo
}
