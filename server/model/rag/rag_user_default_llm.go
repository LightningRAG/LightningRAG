package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagUserDefaultLLM 用户默认大模型配置（用户为自己设置）
type RagUserDefaultLLM struct {
	global.LRAG_MODEL
	UserID        uint   `json:"userId" gorm:"uniqueIndex:idx_user_model_type;comment:用户ID"`
	ModelType     string `json:"modelType" gorm:"size:32;uniqueIndex:idx_user_model_type;comment:模型类型 chat|embedding|rerank 等"`
	LLMProviderID uint   `json:"llmProviderId" gorm:"comment:LLM配置ID"`
	LLMSource     string `json:"llmSource" gorm:"size:16;default:user;comment:模型来源 admin|user"`
}

func (RagUserDefaultLLM) TableName() string {
	return "rag_user_default_llms"
}
