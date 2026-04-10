package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagSystemDefaultModel 系统全局默认模型配置（管理员为系统全局设置）
// 回退链：知识库配置 → 用户默认 → 角色默认 → 系统全局默认
type RagSystemDefaultModel struct {
	global.LRAG_MODEL
	ModelType     string `json:"modelType" gorm:"size:32;uniqueIndex;comment:模型类型 chat|embedding|rerank|ocr|cv|speech2text|tts"`
	LLMProviderID uint   `json:"llmProviderId" gorm:"comment:管理员模型 ID(RagLLMProvider)"`
}

func (RagSystemDefaultModel) TableName() string {
	return "rag_system_default_models"
}
