package middleware

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// CasbinHandler 拦截器
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		waitUse, _ := utils.GetClaims(c)
		//获取请求的PATH
		path := c.Request.URL.Path
		obj := strings.TrimPrefix(path, global.LRAG_CONFIG.System.RouterPrefix)
		// 规范化路径：确保以 / 开头，且处理 /api 前缀（代理可能未重写路径）
		obj = normalizePathForCasbin(obj)
		// 获取请求方法
		act := c.Request.Method
		// 获取用户的角色
		sub := strconv.Itoa(int(waitUse.AuthorityId))
		e := utils.GetCasbin() // 判断策略中是否存在
		success, _ := e.Enforce(sub, obj, act)
		if !success {
			// 若 obj 以 /api/ 开头，尝试去掉 /api 再匹配（兼容不同代理配置）
			if strings.HasPrefix(obj, "/api/") {
				objAlt := "/" + strings.TrimPrefix(obj, "/api/")
				success, _ = e.Enforce(sub, objAlt, act)
			}
		}
		if !success {
			response.FailWithDetailed(gin.H{}, i18n.Msg(c, "casbin.forbidden"), c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// normalizePathForCasbin 规范化路径，确保与 sys_api 中存储的格式一致
func normalizePathForCasbin(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}
