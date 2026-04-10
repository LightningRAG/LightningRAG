package task

import (
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"go.uber.org/zap"
)

// PruneRagChannelWebhookEvents 删除过旧的 Webhook 幂等记录，避免表无限增长（每日任务）。
func PruneRagChannelWebhookEvents() {
	if global.LRAG_DB == nil {
		return
	}
	days := global.LRAG_CONFIG.Rag.ChannelWebhookEventRetentionDays
	if days < 0 {
		return
	}
	if days == 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	res := global.LRAG_DB.Unscoped().
		Where("created_at < ?", cutoff).
		Delete(&rag.RagChannelWebhookEvent{})
	if res.Error != nil {
		global.LRAG_LOG.Warn("rag channel webhook events prune", zap.Error(res.Error))
		return
	}
	if res.RowsAffected > 0 {
		global.LRAG_LOG.Info("rag channel webhook events pruned",
			zap.Int64("rows", res.RowsAffected),
			zap.Int("retention_days", days))
	}
}
