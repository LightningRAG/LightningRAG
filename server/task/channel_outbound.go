package task

import (
	"context"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/service"
	"go.uber.org/zap"
)

// ProcessRagChannelOutbound 重试失败的第三方渠道出站消息（指数退避，见 RagChannelOutbound）
func ProcessRagChannelOutbound() {
	if global.LRAG_DB == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// limit 0：批次大小由 rag.channel-outbound-batch-size 决定
	n, err := service.ServiceGroupApp.RagServiceGroup.ChannelConnectorService.ProcessChannelOutboundQueue(ctx, 0)
	if err != nil {
		global.LRAG_LOG.Warn("rag channel outbound queue", zap.Error(err))
		return
	}
	if n > 0 {
		global.LRAG_LOG.Debug("rag channel outbound batch", zap.Int("rows", n))
	}
}
