package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagKnowledgeBaseShare 知识库分享
type RagKnowledgeBaseShare struct {
	global.LRAG_MODEL
	KnowledgeBaseID uint   `json:"knowledgeBaseId" gorm:"index;comment:知识库ID"`
	ShareType       string `json:"shareType" gorm:"size:32;comment:分享类型 share|transfer"`
	TargetType      string `json:"targetType" gorm:"size:32;comment:目标类型 user|role|org"`
	TargetID        uint   `json:"targetId" gorm:"comment:目标ID"`
	Permission      string `json:"permission" gorm:"size:32;default:read;comment:权限 read|write|admin"`
}

func (RagKnowledgeBaseShare) TableName() string {
	return "rag_knowledge_base_shares"
}
