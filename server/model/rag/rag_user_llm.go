package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagUserLLM 用户自定义大模型（用户添加的 API Key）
type RagUserLLM struct {
	global.LRAG_MODEL
	UserID               uint           `json:"userId" gorm:"index;comment:用户ID"`
	Provider             string         `json:"provider" gorm:"size:64;comment:提供商"`
	ModelName            string         `json:"modelName" gorm:"size:128;comment:模型名"`
	ModelTypes           []string       `json:"modelTypes" gorm:"serializer:json;type:text;comment:适用场景 chat|rerank 等"`
	BaseURL              string         `json:"baseUrl" gorm:"size:256;comment:Base URL"`
	APIKey               string         `json:"-" gorm:"size:512;comment:API Key"`
	Config               common.JSONMap `json:"config" gorm:"type:text;comment:配置"`
	MaxContextTokens     uint           `json:"maxContextTokens" gorm:"default:0;comment:最大上下文token数，0表示不限制"`
	SupportsDeepThinking bool           `json:"supportsDeepThinking" gorm:"default:false;comment:是否支持深度思考"`
	SupportsToolCall     bool           `json:"supportsToolCall" gorm:"default:true;comment:是否支持工具调用"`
	Enabled              bool           `json:"enabled" gorm:"default:true"`
}

func (RagUserLLM) TableName() string {
	return "rag_user_llms"
}
