package system

import (
	api "github.com/LightningRAG/LightningRAG/server/api/v1"
	"github.com/LightningRAG/LightningRAG/server/middleware"
	"github.com/gin-gonic/gin"
)

var oauthPublicApi = api.ApiGroupApp.OAuthAppGroup.OAuthPublicApi

type BaseRouter struct{}

func (s *BaseRouter) InitBaseRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	baseRouter := Router.Group("base")
	{
		baseRouter.POST("login", baseApi.Login)
		baseRouter.POST("captcha", baseApi.Captcha)
		oauthPub := baseRouter.Group("")
		oauthPub.Use(middleware.OAuthPublicRateLimit())
		oauthPub.Use(middleware.OAuthPublicNoCache())
		{
			oauthPub.GET("oauth/providers", oauthPublicApi.OAuthPublicProviders)
			oauthPub.GET("oauth/authorize/:kind", oauthPublicApi.OAuthAuthorize)
			oauthPub.GET("oauth/callback/:kind", oauthPublicApi.OAuthCallback)
			oauthPub.GET("oauth/exchange", oauthPublicApi.OAuthExchange)
		}
	}
	return baseRouter
}
