package rag

import (
	"github.com/gin-gonic/gin"
)

func (r *SettingsRouter) InitSettingsRouter(Router *gin.RouterGroup) {
	settingsRouter := Router.Group("rag/settings")
	{
		// 向量存储配置
		settingsRouter.POST("vectorStore/create", settingsApi.CreateVectorStoreConfig)
		settingsRouter.POST("vectorStore/update", settingsApi.UpdateVectorStoreConfig)
		settingsRouter.POST("vectorStore/delete", settingsApi.DeleteVectorStoreConfig)
		settingsRouter.POST("vectorStore/get", settingsApi.GetVectorStoreConfig)
		settingsRouter.POST("vectorStore/list", settingsApi.ListVectorStoreConfigsFull)
		// 文件存储配置
		settingsRouter.POST("fileStorage/create", settingsApi.CreateFileStorageConfig)
		settingsRouter.POST("fileStorage/update", settingsApi.UpdateFileStorageConfig)
		settingsRouter.POST("fileStorage/delete", settingsApi.DeleteFileStorageConfig)
		settingsRouter.POST("fileStorage/get", settingsApi.GetFileStorageConfig)
		settingsRouter.POST("fileStorage/list", settingsApi.ListFileStorageConfigsFull)
	}
}
