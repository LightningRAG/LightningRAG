package rag

import (
	"github.com/gin-gonic/gin"
)

func (r *LLMProviderRouter) InitLLMProviderRouter(Router *gin.RouterGroup) {
	llmRouter := Router.Group("rag/llm")
	{
		llmRouter.POST("listProviders", llmProviderApi.ListProviders)
		llmRouter.POST("listAvailableProviders", llmProviderApi.ListAvailableProviders)
		llmRouter.POST("listUserModels", llmProviderApi.ListUserModels)
		llmRouter.POST("addUserModel", llmProviderApi.AddUserModel)
		llmRouter.POST("updateUserModel", llmProviderApi.UpdateUserModel)
		llmRouter.POST("deleteUserModel", llmProviderApi.DeleteUserModel)
		llmRouter.POST("setAuthorityDefaultLLM", llmProviderApi.SetAuthorityDefaultLLM)
		llmRouter.POST("getAuthorityDefaultLLMs", llmProviderApi.GetAuthorityDefaultLLMs)
		llmRouter.POST("clearAuthorityDefaultLLM", llmProviderApi.ClearAuthorityDefaultLLM)
		llmRouter.POST("setUserDefaultLLM", llmProviderApi.SetUserDefaultLLM)
		llmRouter.POST("getUserDefaultLLMs", llmProviderApi.GetUserDefaultLLMs)
		llmRouter.POST("clearUserDefaultLLM", llmProviderApi.ClearUserDefaultLLM)
		llmRouter.POST("listWebSearchProviders", llmProviderApi.ListWebSearchProviders)
		llmRouter.POST("getWebSearchConfig", llmProviderApi.GetWebSearchConfig)
		llmRouter.POST("setWebSearchConfig", llmProviderApi.SetWebSearchConfig)
	}
}
