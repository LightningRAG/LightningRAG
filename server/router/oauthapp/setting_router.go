package oauthapp

import (
	"github.com/LightningRAG/LightningRAG/server/middleware"
	"github.com/gin-gonic/gin"
)

type OAuthSettingRouter struct{}

func (r *OAuthSettingRouter) InitOAuthSettingRouter(Router *gin.RouterGroup) {
	g := Router.Group("sysOAuthSetting").Use(middleware.OperationRecord())
	gWithout := Router.Group("sysOAuthSetting")
	{
		g.PUT("updateOAuthSetting", oauthSettingApi.Update)
	}
	{
		gWithout.GET("getOAuthSetting", oauthSettingApi.Get)
	}
}
