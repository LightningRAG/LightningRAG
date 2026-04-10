package rag

import (
	_ "github.com/LightningRAG/LightningRAG/server/mcp"       // 确保 MCP 工具注册到对话（conversation_bridge）
	_ "github.com/LightningRAG/LightningRAG/server/rag/tools" // 确保工具包（含 web_search 等）在路由初始化时加载
	"github.com/gin-gonic/gin"
)

func (r *ConversationRouter) InitConversationRouter(Router *gin.RouterGroup) {
	convRouter := Router.Group("rag/conversation")
	{
		convRouter.POST("create", conversationApi.Create)
		convRouter.POST("chat", conversationApi.Chat)
		convRouter.POST("chatStream", conversationApi.ChatStream)
		convRouter.POST("queryData", conversationApi.QueryData)
		convRouter.POST("list", conversationApi.List)
		convRouter.POST("get", conversationApi.Get)
		convRouter.POST("update", conversationApi.Update)
		convRouter.POST("delete", conversationApi.Delete)
		convRouter.POST("listMessages", conversationApi.ListMessages)
		convRouter.POST("listTools", conversationApi.ListTools)
	}
}
