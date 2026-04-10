package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// RateLimitWaitError is returned when the client must wait before retrying.
type RateLimitWaitError struct {
	Seconds int64
}

func (e RateLimitWaitError) Error() string { return "rate_limit_wait" }

var errRateLimitTooFrequent = errors.New("rate_limit_too_frequent")

type LimitConfig struct {
	// GenerationKey 根据业务生成key 下面CheckOrMark查询生成
	GenerationKey func(c *gin.Context) string
	// 检查函数,用户可修改具体逻辑,更加灵活
	CheckOrMark func(key string, expire int, limit int) error
	// Expire key 过期时间
	Expire int
	// Limit 周期时间
	Limit int
}

func (l LimitConfig) LimitWithTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := l.CheckOrMark(l.GenerationKey(c), l.Expire, l.Limit); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": response.ERROR, "msg": rateLimitMessage(c, err)})
			c.Abort()
			return
		} else {
			c.Next()
		}
	}
}

// DefaultGenerationKey 默认生成key
func DefaultGenerationKey(c *gin.Context) string {
	return "LightningRAG_Limit" + c.ClientIP()
}

func DefaultCheckOrMark(key string, expire int, limit int) (err error) {
	// 判断是否开启redis
	if global.LRAG_REDIS == nil {
		return err
	}
	if err = SetLimitWithTime(key, limit, time.Duration(expire)*time.Second); err != nil {
		global.LRAG_LOG.Error("limit", zap.Error(err))
	}
	return err
}

func DefaultLimit() gin.HandlerFunc {
	return LimitConfig{
		GenerationKey: DefaultGenerationKey,
		CheckOrMark:   DefaultCheckOrMark,
		Expire:        global.LRAG_CONFIG.System.LimitTimeIP,
		Limit:         global.LRAG_CONFIG.System.LimitCountIP,
	}.LimitWithTime()
}

// SetLimitWithTime 设置访问次数
func SetLimitWithTime(key string, limit int, expiration time.Duration) error {
	count, err := global.LRAG_REDIS.Exists(context.Background(), key).Result()
	if err != nil {
		return err
	}
	if count == 0 {
		pipe := global.LRAG_REDIS.TxPipeline()
		pipe.Incr(context.Background(), key)
		pipe.Expire(context.Background(), key, expiration)
		_, err = pipe.Exec(context.Background())
		return err
	} else {
		// 次数
		if times, err := global.LRAG_REDIS.Get(context.Background(), key).Int(); err != nil {
			return err
		} else {
			if times >= limit {
				ttl, errPTTL := global.LRAG_REDIS.PTTL(context.Background(), key).Result()
				if errPTTL != nil || ttl <= 0 {
					return errRateLimitTooFrequent
				}
				sec := int64((ttl + time.Second - 1) / time.Second)
				if sec < 1 {
					sec = 1
				}
				return RateLimitWaitError{Seconds: sec}
			} else {
				return global.LRAG_REDIS.Incr(context.Background(), key).Err()
			}
		}
	}
}

func rateLimitMessage(c *gin.Context, err error) string {
	var rw RateLimitWaitError
	if errors.As(err, &rw) {
		return i18n.Msgf(c, "rate_limit.wait_seconds", rw.Seconds)
	}
	if errors.Is(err, errRateLimitTooFrequent) {
		return i18n.Msg(c, "rate_limit.too_frequent")
	}
	return err.Error()
}
