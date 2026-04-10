package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagLLMProvider 大模型提供商配置（管理员配置）
type RagLLMProvider struct {
	global.LRAG_MODEL
	Name                 string         `json:"name" gorm:"size:64;comment:提供商名称 openai|ollama|anthropic 等"`
	ModelName            string         `json:"modelName" gorm:"size:128;comment:模型名称"`
	ModelTypes           []string       `json:"modelTypes" gorm:"serializer:json;type:text;comment:适用场景 chat|rerank 等"`
	BaseURL              string         `json:"baseUrl" gorm:"size:256;comment:API Base URL"`
	APIKey               string         `json:"-" gorm:"size:512;comment:API Key(加密存储)"`
	Config               common.JSONMap `json:"config" gorm:"type:text;comment:额外配置 JSON"`
	MaxContextTokens     uint           `json:"maxContextTokens" gorm:"default:0;comment:最大上下文token数，0表示不限制"`
	SupportsDeepThinking bool           `json:"supportsDeepThinking" gorm:"default:false;comment:是否支持深度思考"`
	SupportsToolCall     bool           `json:"supportsToolCall" gorm:"default:true;comment:是否支持工具调用"`
	ShareScope           string         `json:"shareScope" gorm:"size:32;default:private;comment:共享范围 private|role|org|all"`
	ShareTarget          string         `json:"shareTarget" gorm:"type:text;comment:共享目标 JSON 如角色ID列表"`
	Enabled              bool           `json:"enabled" gorm:"default:true;comment:是否启用"`
}

func (RagLLMProvider) TableName() string {
	return "rag_llm_providers"
}
