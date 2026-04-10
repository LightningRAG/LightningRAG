package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/middleware"
	"github.com/gin-gonic/gin"
)

var (
	deferredPluginEngine    *gin.Engine
	pluginInstallWasSkipped bool
)

// RegisterDeferredPluginEngine 由 Routers 注册当前 Gin 引擎，供 InitDB 完成后补装插件路由。
func RegisterDeferredPluginEngine(e *gin.Engine) {
	deferredPluginEngine = e
}

func InstallPlugin(PrivateGroup *gin.RouterGroup, PublicRouter *gin.RouterGroup, engine *gin.Engine) {
	if global.LRAG_DB == nil {
		pluginInstallWasSkipped = true
		global.LRAG_LOG.Info("项目暂未初始化，已延迟插件路由安装；完成数据库初始化后将自动补装（无需重启）")
		return
	}
	pluginInstallWasSkipped = false
	bizPluginV1(PrivateGroup, PublicRouter)
	bizPluginV2(engine)
}

// TryDeferredPluginInstall 在 InitDB 成功后调用：若首次 Routers 因无数据库跳过了 InstallPlugin，则在此挂载插件路由。
func TryDeferredPluginInstall() {
	if !pluginInstallWasSkipped || global.LRAG_DB == nil || deferredPluginEngine == nil {
		return
	}
	prefix := global.LRAG_CONFIG.System.RouterPrefix
	publicGroup := deferredPluginEngine.Group(prefix)
	privateGroup := deferredPluginEngine.Group(prefix)
	privateGroup.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	bizPluginV1(privateGroup, publicGroup)
	bizPluginV2(deferredPluginEngine)
	pluginInstallWasSkipped = false
	global.LRAG_LOG.Info("InitDB 后回调：插件路由补装完成")
}
