package rag

import "github.com/LightningRAG/LightningRAG/server/global"

// RagChannelSession 外部会话线程与内部 rag_conversations 的映射（实现多轮）
type RagChannelSession struct {
	global.LRAG_MODEL
	ConnectorID    uint   `json:"connectorId" gorm:"uniqueIndex:idx_connector_thread;index;comment:连接器ID"`
	ThreadKey      string `json:"threadKey" gorm:"size:512;uniqueIndex:idx_connector_thread;comment:渠道侧会话键"`
	ConversationID uint   `json:"conversationId" gorm:"index;comment:内部对话ID"`
}

func (RagChannelSession) TableName() string {
	return "rag_channel_sessions"
}
