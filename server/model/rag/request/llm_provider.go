package request

// LLMProviderAddUser 用户添加自定义模型
type LLMProviderAddUser struct {
	Provider             string   `json:"provider" binding:"required"`
	ModelName            string   `json:"modelName" binding:"required"`
	ModelTypes           []string `json:"modelTypes"` // 适用场景 chat|embedding|rerank|speech2text|tts|ocr|cv，空则默认 chat
	BaseURL              string   `json:"baseUrl"`
	APIKey               string   `json:"apiKey"`
	Config               string   `json:"config"`               // JSON
	MaxContextTokens     uint     `json:"maxContextTokens"`     // 最大上下文token数，0表示不限制
	SupportsDeepThinking bool     `json:"supportsDeepThinking"` // 是否支持深度思考
	SupportsToolCall     bool     `json:"supportsToolCall"`     // 是否支持工具调用
}

// LLMProviderUpdateUser 更新用户模型；APIKey 非空时更新密钥，留空则保持原值
type LLMProviderUpdateUser struct {
	ID                   uint     `json:"id" binding:"required"`
	Provider             string   `json:"provider" binding:"required"`
	ModelName            string   `json:"modelName" binding:"required"`
	ModelTypes           []string `json:"modelTypes"` // 适用场景 chat|embedding|rerank|speech2text|tts|ocr|cv
	BaseURL              string   `json:"baseUrl"`
	APIKey               string   `json:"apiKey"`               // 留空不修改
	MaxContextTokens     uint     `json:"maxContextTokens"`     // 最大上下文token数，0表示不限制
	SupportsDeepThinking bool     `json:"supportsDeepThinking"` // 是否支持深度思考
	SupportsToolCall     bool     `json:"supportsToolCall"`     // 是否支持工具调用
}

// LLMProviderListReq 列出模型请求（可按场景过滤）
type LLMProviderListReq struct {
	ScenarioType string `json:"scenarioType"` // 场景类型 chat|embedding|rerank|speech2text|tts|ocr|cv，空则返回全部
}

// LLMProviderDeleteUser 删除用户模型
type LLMProviderDeleteUser struct {
	ID uint `json:"id" binding:"required"`
}

// LLMProviderListAvailableReq 获取可用提供商列表（按场景类型）
type LLMProviderListAvailableReq struct {
	ScenarioType  string   `json:"scenarioType"`  // 单场景类型，兼容旧参数
	ScenarioTypes []string `json:"scenarioTypes"` // 多场景类型，返回支持任一场景的提供商并集
}

// SetAuthorityDefaultLLMReq 设置角色默认模型（管理员）
type SetAuthorityDefaultLLMReq struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"`
	ModelType     string `json:"modelType" binding:"required"` // chat|embedding|rerank 等
	LLMProviderID uint   `json:"llmProviderId" binding:"required"`
	LLMSource     string `json:"llmSource"` // admin|user，默认 admin
}

// GetAuthorityDefaultLLMsReq 获取角色默认模型列表
type GetAuthorityDefaultLLMsReq struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"`
}

// ClearAuthorityDefaultLLMReq 清除角色某类型默认模型
type ClearAuthorityDefaultLLMReq struct {
	AuthorityId uint   `json:"authorityId" binding:"required"`
	ModelType   string `json:"modelType" binding:"required"`
}

// SetUserDefaultLLMReq 设置用户默认模型（用户为自己设置）
type SetUserDefaultLLMReq struct {
	ModelType     string `json:"modelType" binding:"required"` // chat|embedding|rerank 等
	LLMProviderID uint   `json:"llmProviderId" binding:"required"`
	LLMSource     string `json:"llmSource"` // admin|user，默认 user
}

// GetUserDefaultLLMsReq 获取用户默认模型列表
type GetUserDefaultLLMsReq struct {
	// 无参数，从 token 获取当前用户
}

// ClearUserDefaultLLMReq 清除用户某类型默认模型
type ClearUserDefaultLLMReq struct {
	ModelType string `json:"modelType" binding:"required"`
}

// WebSearchConfigSetReq 设置用户互联网搜索配置
type WebSearchConfigSetReq struct {
	UseSystemDefault *bool             `json:"useSystemDefault"` // 是否使用系统默认，为 true 时忽略 provider/config
	Provider         string            `json:"provider"`         // duckduckgo|baidu
	Config           map[string]string `json:"config"`           // 引擎配置，如 {"apiKey":"xxx"}
}
