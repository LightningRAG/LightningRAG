package middleware

import (
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/gin-gonic/gin"
)

// Locale resolves request language from X-Locale or Accept-Language (aligned with the web app).
func Locale() gin.HandlerFunc {
	return i18n.Middleware()
}
