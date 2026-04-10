package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagAuthorityDefaultLLM 角色默认大模型配置（管理员为角色设置）
type RagAuthorityDefaultLLM struct {
	global.LRAG_MODEL
	AuthorityId   uint   `json:"authorityId" gorm:"uniqueIndex:idx_authority_model_type;comment:角色ID"`
	ModelType     string `json:"modelType" gorm:"size:32;uniqueIndex:idx_authority_model_type;comment:模型类型 chat|embedding|rerank 等"`
	LLMProviderID uint   `json:"llmProviderId" gorm:"comment:LLM配置ID"`
	LLMSource     string `json:"llmSource" gorm:"size:16;default:admin;comment:模型来源 admin|user"`
}

func (RagAuthorityDefaultLLM) TableName() string {
	return "rag_authority_default_llms"
}
