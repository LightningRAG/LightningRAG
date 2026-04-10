package rag

import "github.com/LightningRAG/LightningRAG/server/global"

// RagChannelConnector 第三方对话渠道与 Agent 的绑定（Webhook 鉴权走 WebhookSecret，不经 JWT）
type RagChannelConnector struct {
	global.LRAG_MODEL
	OwnerID       uint   `json:"ownerId" gorm:"index;comment:创建者用户ID"`
	AuthorityID   uint   `json:"authorityId" gorm:"index;comment:创建时的角色ID，用于默认模型等解析"`
	Name          string `json:"name" gorm:"size:128;comment:连接器名称"`
	Channel       string `json:"channel" gorm:"size:32;index;comment:渠道 mock|feishu|discord|dingtalk|wechat_mp|wecom|slack|telegram|teams|whatsapp|line"`
	AgentID       uint   `json:"agentId" gorm:"index;comment:绑定的 Agent ID"`
	WebhookSecret string `json:"-" gorm:"size:256;comment:Webhook 共享密钥"`
	Enabled       bool   `json:"enabled" gorm:"default:true;comment:是否启用"`
	Extra         string `json:"extra" gorm:"type:text;comment:渠道扩展 JSON：飞书 app_id/secret/encrypt_key，可选 feishu_api_base 或 lark_api_base（国际 Lark）；微信/钉钉见文档；wecom_*；slack_*；telegram_*；teams_*；whatsapp_*；line_*"`
}

func (RagChannelConnector) TableName() string {
	return "rag_channel_connectors"
}
