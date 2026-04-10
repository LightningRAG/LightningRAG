package rag

import (
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagChannelOutbound 第三方渠道出站消息重试队列（SendReply 失败后入队，定时任务指数退避重试）
type RagChannelOutbound struct {
	global.LRAG_MODEL
	ConnectorID uint       `json:"connectorId" gorm:"index;not null;comment:rag_channel_connectors.id"`
	Channel     string     `json:"channel" gorm:"size:32;index;comment:渠道类型"`
	Payload     string     `json:"payload" gorm:"type:text;not null;comment:JSON：text + opaque（ReplyRef）"`
	Attempts    int        `json:"attempts" gorm:"default:0;comment:已执行的重试次数（含本轮）"`
	NextRetryAt time.Time  `json:"nextRetryAt" gorm:"index;comment:下次可重试时间"`
	LeaseUntil  *time.Time `json:"leaseUntil,omitempty" gorm:"index;comment:多实例出站认领租约，到期前其他 worker 不抢占"`
	LastErr     string     `json:"lastErr" gorm:"size:1024;comment:最近一次失败原因"`
}

func (RagChannelOutbound) TableName() string {
	return "rag_channel_outbounds"
}
