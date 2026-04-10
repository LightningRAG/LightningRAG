package request

// ChannelConnectorCreate 创建渠道连接器
type ChannelConnectorCreate struct {
	Name          string `json:"name" binding:"required"`
	Channel       string `json:"channel" binding:"required"` // mock, feishu, discord, dingtalk, wechat_mp, wecom, slack, telegram, teams, whatsapp, line
	AgentID       uint   `json:"agentId" binding:"required"`
	WebhookSecret string `json:"webhookSecret"` // 可选，空则服务端生成
	Enabled       *bool  `json:"enabled"`
	Extra         string `json:"extra"` // JSON 字符串
}

// ChannelConnectorUpdate 更新
type ChannelConnectorUpdate struct {
	ID            uint   `json:"id" binding:"required"`
	Name          string `json:"name"`
	AgentID       uint   `json:"agentId"`
	WebhookSecret string `json:"webhookSecret"` // 非空则轮换密钥
	Enabled       *bool  `json:"enabled"`
	Extra         string `json:"extra"`
}

// ChannelConnectorList 列表
type ChannelConnectorList struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Channel  string `json:"channel" form:"channel"`
	Name     string `json:"name" form:"name"`
}

// ChannelConnectorGet / Delete
type ChannelConnectorGet struct {
	ID uint `json:"id" binding:"required"`
}

type ChannelConnectorDelete struct {
	ID uint `json:"id" binding:"required"`
}

// ChannelOutboundList 出站重试队列列表
type ChannelOutboundList struct {
	Page        int  `json:"page" form:"page"`
	PageSize    int  `json:"pageSize" form:"pageSize"`
	ConnectorID uint `json:"connectorId" form:"connectorId"`
}

// ChannelOutboundDelete 删除一条出站重试任务
type ChannelOutboundDelete struct {
	ID uint `json:"id" binding:"required"`
}
