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

// OAuthPublicNoCache 禁止缓存 OAuth 公开端点响应，避免浏览器对 GET /authorize 返回 304 并复用旧 302，导致无法跳转 Gitee 或误用过期 Location。
func OAuthPublicNoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Next()
	}
}

// OAuthPublicRateLimit 限制公开 OAuth 端点（authorize / callback / exchange / providers）按客户端 IP 的请求频率。
// 处理器见 api/v1/oauthapp.OAuthPublicApi。
// 仅当 rag.oauth-ip-limit-per-minute > 0 且已配置 Redis 时生效；Redis 报错时放行。
func OAuthPublicRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := global.LRAG_CONFIG.Rag.OAuthIPLimitPerMinute
		if limit <= 0 || global.LRAG_REDIS == nil {
			c.Next()
			return
		}
		key := "LragOAuthIP:" + c.ClientIP()
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
		global.LRAG_LOG.Warn("oauth public rate limit check skipped", zap.Error(err))
		c.Next()
	}
}
