package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/google/uuid"
)

// RagConversation 对话会话
type RagConversation struct {
	global.LRAG_MODEL
	UUID             uuid.UUID `json:"uuid" gorm:"index;comment:会话UUID"`
	UserID           uint      `json:"userId" gorm:"index;comment:用户ID"`
	Title            string    `json:"title" gorm:"size:256;default:新对话;comment:对话标题"`
	LLMProviderID    uint      `json:"llmProviderId" gorm:"comment:使用的LLM配置ID"`
	LLMSource        string    `json:"llmSource" gorm:"size:16;default:user;comment:模型来源 admin|user"`
	SourceType       string    `json:"sourceType" gorm:"size:32;comment:来源类型 knowledge_base|files"`
	SourceIDs        string    `json:"sourceIds" gorm:"type:text;comment:来源ID JSON 如知识库ID列表或文件ID列表"`
	EnabledToolNames string    `json:"enabledToolNames" gorm:"type:text;comment:启用的工具 JSON 数组如 [\"web_search\"]，空则不使用任何工具"`
}

func (RagConversation) TableName() string {
	return "rag_conversations"
}
