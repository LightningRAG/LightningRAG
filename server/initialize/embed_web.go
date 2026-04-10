package initialize

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/webui"
	"github.com/gin-gonic/gin"
)

func registerEmbeddedWebUI(Router *gin.Engine) {
	if !global.LRAG_CONFIG.System.EmbedWebUI {
		return
	}
	if !webui.HasDist() {
		global.LRAG_LOG.Warn("已开启 embed-web-ui，但未找到 webui/webdist/index.html；请在仓库根目录执行 scripts/sync-web-dist.sh（或 make sync-web-dist）后再编译 server")
		return
	}
	sub, err := fs.Sub(webui.Dist(), "webdist")
	if err != nil {
		global.LRAG_LOG.Error("embed webui: " + err.Error())
		return
	}
	global.LRAG_LOG.Info("embed web-ui: 已注册内置静态资源与 SPA 回退")

	Router.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		reqPath := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
		if reqPath == "." {
			reqPath = ""
		}

		tryPaths := []string{}
		if reqPath != "" {
			tryPaths = append(tryPaths, reqPath)
		}
		tryPaths = append(tryPaths, "index.html")

		for _, tp := range tryPaths {
			if strings.Contains(tp, "..") {
				continue
			}
			data, err := fs.ReadFile(sub, tp)
			if err != nil {
				continue
			}
			ct := mime.TypeByExtension(path.Ext(tp))
			if ct == "" {
				ct = "application/octet-stream"
			}
			if strings.HasPrefix(ct, "text/html") {
				c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
			}
			c.Data(http.StatusOK, ct, data)
			return
		}
		c.AbortWithStatus(http.StatusNotFound)
	})
}
