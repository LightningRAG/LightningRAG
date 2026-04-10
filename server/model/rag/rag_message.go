package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/google/uuid"
)

// RagMessage 对话消息
type RagMessage struct {
	global.LRAG_MODEL
	UUID           uuid.UUID `json:"uuid" gorm:"index;comment:消息UUID"`
	ConversationID uint      `json:"conversationId" gorm:"index;comment:会话ID"`
	Role           string    `json:"role" gorm:"size:16;comment:角色 user|assistant|system"`
	Content        string    `json:"content" gorm:"type:text;comment:消息内容"`
	TokenCount     int       `json:"tokenCount" gorm:"comment:Token 数"`
	References     string    `json:"references" gorm:"type:text;comment:引用来源 JSON 数组"`
}

func (RagMessage) TableName() string {
	return "rag_messages"
}
