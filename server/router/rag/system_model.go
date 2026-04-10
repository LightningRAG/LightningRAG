package rag

import (
	"github.com/gin-gonic/gin"
)

func (r *SystemModelRouter) InitSystemModelRouter(Router *gin.RouterGroup) {
	smRouter := Router.Group("rag/systemModel")
	{
		smRouter.POST("listAdminModels", systemModelApi.ListAdminModels)
		smRouter.POST("createAdminModel", systemModelApi.CreateAdminModel)
		smRouter.POST("updateAdminModel", systemModelApi.UpdateAdminModel)
		smRouter.POST("deleteAdminModel", systemModelApi.DeleteAdminModel)
		smRouter.POST("getSystemDefaults", systemModelApi.GetSystemDefaults)
		smRouter.POST("setSystemDefault", systemModelApi.SetSystemDefault)
		smRouter.POST("clearSystemDefault", systemModelApi.ClearSystemDefault)
		smRouter.POST("listSystemWebSearchProviders", systemModelApi.ListSystemWebSearchProviders)
		smRouter.POST("getSystemWebSearchConfig", systemModelApi.GetSystemWebSearchConfig)
		smRouter.POST("setSystemWebSearchConfig", systemModelApi.SetSystemWebSearchConfig)
		smRouter.POST("clearSystemWebSearchConfig", systemModelApi.ClearSystemWebSearchConfig)
		// 全局共享知识库
		smRouter.POST("listGlobalKnowledgeBases", systemModelApi.ListGlobalKnowledgeBases)
		smRouter.POST("setGlobalKnowledgeBase", systemModelApi.SetGlobalKnowledgeBase)
		smRouter.POST("removeGlobalKnowledgeBase", systemModelApi.RemoveGlobalKnowledgeBase)
		smRouter.POST("listAllKnowledgeBases", systemModelApi.ListAllKnowledgeBases)
	}
}
