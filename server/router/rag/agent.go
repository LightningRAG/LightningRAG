package rag

import (
	"github.com/gin-gonic/gin"
)

func (r *AgentRouter) InitAgentRouter(Router *gin.RouterGroup) {
	agentRouter := Router.Group("rag/agent")
	{
		agentRouter.POST("run", agentApi.Run)
		agentRouter.POST("runStream", agentApi.RunStream)
		agentRouter.POST("templates", agentApi.ListTemplates)
		agentRouter.POST("loadTemplate", agentApi.LoadTemplate)
		agentRouter.POST("create", agentApi.Create)
		agentRouter.POST("list", agentApi.List)
		agentRouter.POST("get", agentApi.Get)
		agentRouter.POST("update", agentApi.Update)
		agentRouter.POST("delete", agentApi.Delete)
		agentRouter.POST("createFromTemplate", agentApi.CreateFromTemplate)
	}
}
