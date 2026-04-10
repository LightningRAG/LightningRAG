package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OpenChannelWebhookRateLimit 按「连接器 ID + 客户端 IP」限制每分钟请求数（滑动窗口由 SetLimitWithTime 的 TTL 决定）。
// 仅当 rag.channel-webhook-ip-limit-per-minute > 0 且已配置 Redis 时生效；Redis 报错时放行，避免误伤第三方回调。
func OpenChannelWebhookRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := global.LRAG_CONFIG.Rag.ChannelWebhookIPLimitPerMinute
		if limit <= 0 || global.LRAG_REDIS == nil {
			c.Next()
			return
		}
		cid := c.Param("connectorId")
		if cid == "" {
			c.Next()
			return
		}
		key := "LragChanWh:" + cid + ":" + c.ClientIP()
		err := SetLimitWithTime(key, limit, time.Minute)
		if err == nil {
			c.Next()
			return
		}
		var rw RateLimitWaitError
		if errors.As(err, &rw) {
			c.Header("Retry-After", strconv.FormatInt(rw.Seconds, 10))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		if errors.Is(err, errRateLimitTooFrequent) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		global.LRAG_LOG.Warn("channel webhook rate limit check skipped", zap.Error(err))
		c.Next()
	}
}
