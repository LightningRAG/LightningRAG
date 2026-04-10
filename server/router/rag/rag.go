package rag

import (
	"github.com/gin-gonic/gin"
)

// InitRagRouter 初始化 RAG 相关路由
func (r *RouterGroup) InitRagRouter(Router *gin.RouterGroup) {
	r.InitKnowledgeBaseRouter(Router)
	r.InitSettingsRouter(Router)
	r.InitConversationRouter(Router)
	r.InitLLMProviderRouter(Router)
	r.InitAgentRouter(Router)
	r.InitChannelConnectorRouter(Router)
	r.InitSystemModelRouter(Router)
}
