package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagGlobalKnowledgeBase 系统全局共享知识库配置
// 管理员可在系统配置中将指定知识库设为全局共享，用户对话时自动使用
type RagGlobalKnowledgeBase struct {
	global.LRAG_MODEL
	KnowledgeBaseID uint   `json:"knowledgeBaseId" gorm:"uniqueIndex;comment:知识库ID"`
	Description     string `json:"description" gorm:"size:256;comment:备注说明"`
}

func (RagGlobalKnowledgeBase) TableName() string {
	return "rag_global_knowledge_bases"
}
