package oauthapp

import (
	"github.com/LightningRAG/LightningRAG/server/middleware"
	"github.com/gin-gonic/gin"
)

type OAuthProviderRouter struct{}

func (r *OAuthProviderRouter) InitOAuthProviderRouter(Router *gin.RouterGroup) {
	g := Router.Group("sysOAuthProvider").Use(middleware.OperationRecord())
	gWithout := Router.Group("sysOAuthProvider")
	{
		g.POST("createOAuthProvider", oauthProviderApi.Create)
		g.DELETE("deleteOAuthProvider", oauthProviderApi.Delete)
		g.DELETE("deleteOAuthProviderByIds", oauthProviderApi.DeleteByIds)
		g.PUT("updateOAuthProvider", oauthProviderApi.Update)
	}
	{
		gWithout.GET("findOAuthProvider", oauthProviderApi.Find)
		gWithout.GET("getOAuthProviderList", oauthProviderApi.List)
		gWithout.GET("getRegisteredOAuthKinds", oauthProviderApi.RegisteredKinds)
	}
}
