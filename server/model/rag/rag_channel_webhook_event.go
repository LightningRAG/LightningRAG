package rag

import "github.com/LightningRAG/LightningRAG/server/global"

// RagChannelWebhookEvent Webhook 事件幂等（connector_id + event_key 唯一）
type RagChannelWebhookEvent struct {
	global.LRAG_MODEL
	ConnectorID uint   `json:"connectorId" gorm:"uniqueIndex:idx_webhook_event;index;comment:连接器ID"`
	EventKey    string `json:"eventKey" gorm:"size:191;uniqueIndex:idx_webhook_event;comment:平台事件唯一键"`
}

func (RagChannelWebhookEvent) TableName() string {
	return "rag_channel_webhook_events"
}
