package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const localeContextKey = "lrag_locale"

// Middleware parses Accept-Language (and optional X-Locale for overrides) into the Gin context.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := DefaultLocale
		if x := strings.TrimSpace(c.GetHeader("X-Locale")); x != "" {
			lower := strings.ToLower(x)
			if code := normalizeLocaleTag(lower); code != "" {
				lang = code
			} else if IsSupportedLocale(x) {
				lang = x
			}
		} else if h := c.GetHeader("Accept-Language"); h != "" {
			lang = ParseAcceptLanguage(h)
		}
		if !IsSupportedLocale(lang) {
			lang = DefaultLocale
		}
		c.Set(localeContextKey, lang)
		c.Next()
	}
}

// GetLocale returns the resolved locale for the request, or DefaultLocale if missing.
func GetLocale(c *gin.Context) string {
	if c == nil {
		return DefaultLocale
	}
	if v, ok := c.Get(localeContextKey); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s
		}
	}
	return DefaultLocale
}

// Msg is T bound to the request locale.
func Msg(c *gin.Context, key string) string {
	return T(GetLocale(c), key)
}

// Msgf is Tf bound to the request locale.
func Msgf(c *gin.Context, key string, args ...any) string {
	return Tf(GetLocale(c), key, args...)
}
