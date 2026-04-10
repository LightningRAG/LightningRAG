package rag

import "github.com/gin-gonic/gin"

// InitChannelConnectorRouter 管理端：连接器 CRUD（JWT + Casbin）
func (r *ChannelConnectorRouter) InitChannelConnectorRouter(Router *gin.RouterGroup) {
	g := Router.Group("rag/channelConnector")
	{
		g.POST("create", channelConnectorApi.CreateChannelConnector)
		g.POST("update", channelConnectorApi.UpdateChannelConnector)
		g.POST("list", channelConnectorApi.ListChannelConnectors)
		g.POST("channelTypes", channelConnectorApi.ListChannelConnectorTypes)
		g.POST("get", channelConnectorApi.GetChannelConnector)
		g.POST("delete", channelConnectorApi.DeleteChannelConnector)
		g.POST("outbound/list", channelConnectorApi.ListChannelOutboundQueue)
		g.POST("outbound/delete", channelConnectorApi.DeleteChannelOutboundRow)
		g.POST("outbound/runOnce", channelConnectorApi.RunChannelOutboundOnce)
	}
}

// InitOpenChannelWebhookRouter 公开 Webhook（仅 X-Webhook-Secret）
func (r *ChannelConnectorRouter) InitOpenChannelWebhookRouter(Router *gin.RouterGroup) {
	Router.GET("open/channel/webhook/:connectorId", channelConnectorApi.OpenChannelWebhookGet)
	Router.POST("open/channel/webhook/:connectorId", channelConnectorApi.OpenChannelWebhook)
}
