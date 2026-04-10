package request

import "github.com/LightningRAG/LightningRAG/server/model/common/request"

// AdminModelCreate 管理员添加系统模型
type AdminModelCreate struct {
	Name                 string   `json:"name" binding:"required"`
	ModelName            string   `json:"modelName" binding:"required"`
	ModelTypes           []string `json:"modelTypes"`
	BaseURL              string   `json:"baseUrl"`
	APIKey               string   `json:"apiKey"`
	MaxContextTokens     uint     `json:"maxContextTokens"`
	SupportsDeepThinking bool     `json:"supportsDeepThinking"`
	SupportsToolCall     bool     `json:"supportsToolCall"`
	ShareScope           string   `json:"shareScope"`
	Enabled              bool     `json:"enabled"`
}

// AdminModelUpdate 管理员更新系统模型
type AdminModelUpdate struct {
	ID                   uint     `json:"id" binding:"required"`
	Name                 string   `json:"name"`
	ModelName            string   `json:"modelName"`
	ModelTypes           []string `json:"modelTypes"`
	BaseURL              string   `json:"baseUrl"`
	APIKey               string   `json:"apiKey"`
	MaxContextTokens     uint     `json:"maxContextTokens"`
	SupportsDeepThinking bool     `json:"supportsDeepThinking"`
	SupportsToolCall     bool     `json:"supportsToolCall"`
	ShareScope           string   `json:"shareScope"`
	Enabled              *bool    `json:"enabled"`
}

// AdminModelDelete 管理员删除系统模型
type AdminModelDelete struct {
	ID uint `json:"id" binding:"required"`
}

// AdminModelList 管理员模型列表
type AdminModelList struct {
	request.PageInfo
	ScenarioType string `json:"scenarioType"`
}

// SystemDefaultModelSet 设置系统全局默认模型
type SystemDefaultModelSet struct {
	ModelType     string `json:"modelType" binding:"required"`
	LLMProviderID uint   `json:"llmProviderId" binding:"required"`
}

// SystemDefaultModelClear 清除系统全局默认模型
type SystemDefaultModelClear struct {
	ModelType string `json:"modelType" binding:"required"`
}

// SystemWebSearchConfigSet 设置系统全局默认互联网搜索配置
type SystemWebSearchConfigSet struct {
	Provider string            `json:"provider" binding:"required"` // duckduckgo|baidu
	Config   map[string]string `json:"config"`                      // 引擎配置，如 {"apiKey":"xxx"}
}
