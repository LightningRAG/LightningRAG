package middleware

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"github.com/LightningRAG/LightningRAG/server/plugin/email/utils"
	utils2 "github.com/LightningRAG/LightningRAG/server/utils"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorToEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var username string
		claims, _ := utils2.GetClaims(c)
		if claims.Username != "" {
			username = claims.Username
		} else {
			id, _ := strconv.Atoi(c.Request.Header.Get("x-user-id"))
			var u system.SysUser
			err := global.LRAG_DB.Where("id = ?", id).First(&u).Error
			if err != nil {
				username = "Unknown"
			}
			username = u.Username
		}
		body, _ := io.ReadAll(c.Request.Body)
		// 再重新写回请求体body中，ioutil.ReadAll会清空c.Request.Body中的数据
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		record := system.SysOperationRecord{
			Ip:     c.ClientIP(),
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Agent:  c.Request.UserAgent(),
			Body:   string(body),
		}
		now := time.Now()

		c.Next()

		latency := time.Since(now)
		status := c.Writer.Status()
		record.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		lang := i18n.GetLocale(c)
		str := i18n.Tf(lang, "sys.alert.http_error_body", record.Body, record.Method, record.ErrorMessage, latency.String())
		if status != 200 {
			subject := i18n.Tf(lang, "sys.alert.http_error_subject", username, record.Ip, record.Path)
			if err := utils.ErrorToEmail(subject, str); err != nil {
				global.LRAG_LOG.Error("ErrorToEmail Failed, err:", zap.Error(err))
			}
		}
	}
}
