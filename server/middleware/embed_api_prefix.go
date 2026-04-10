package middleware

import (
	"net/http"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/webui"
)

// WrapStripLeadingAPIPrefixBeforeRouting 在 Gin 路由匹配之前去掉 /api 前缀，使 /api/base/xxx 命中已注册的 /base/xxx。
//
// 注意：仅靠 Engine.Use 的中间件无法修复此类 404——handleHTTPRequest 先用原始 Path 查路由树，
// 未命中即走 NoRoute，此时再改写 Path 不会重新匹配。因此必须在 http.Handler 层克隆请求并改 Path。
func WrapStripLeadingAPIPrefixBeforeRouting(h http.Handler) http.Handler {
	if !global.LRAG_CONFIG.System.EmbedWebUI || !webui.HasDist() || global.LRAG_CONFIG.System.RouterPrefix != "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if len(p) >= 4 && p[:4] == "/api" && (len(p) == 4 || p[4] == '/') {
			nr := r.Clone(r.Context())
			np := strings.TrimPrefix(p, "/api")
			if np == "" {
				np = "/"
			}
			nr.URL.Path = np
			nr.URL.RawPath = ""
			h.ServeHTTP(w, nr)
			return
		}
		h.ServeHTTP(w, r)
	})
}
