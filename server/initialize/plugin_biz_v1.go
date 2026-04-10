package initialize

import (
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/plugin/email"
	"github.com/LightningRAG/LightningRAG/server/utils/plugin"
	"github.com/gin-gonic/gin"
)

func PluginInit(group *gin.RouterGroup, Plugin ...plugin.Plugin) {
	for i := range Plugin {
		fmt.Println(Plugin[i].RouterPath(), "注册开始!")
		PluginGroup := group.Group(Plugin[i].RouterPath())
		Plugin[i].Register(PluginGroup)
		fmt.Println(Plugin[i].RouterPath(), "注册成功!")
	}
}

func bizPluginV1(group ...*gin.RouterGroup) {
	private := group[0]
	public := group[1]
	//  添加跟角色挂钩权限的插件 示例 本地示例模式于在线仓库模式注意上方的import 可以自行切换 效果相同
	PluginInit(private, email.CreateEmailPlug(
		global.LRAG_CONFIG.Email.To,
		global.LRAG_CONFIG.Email.From,
		global.LRAG_CONFIG.Email.Host,
		global.LRAG_CONFIG.Email.Secret,
		global.LRAG_CONFIG.Email.Nickname,
		global.LRAG_CONFIG.Email.Port,
		global.LRAG_CONFIG.Email.IsSSL,
		global.LRAG_CONFIG.Email.IsLoginAuth,
	))
	holder(public, private)
}
