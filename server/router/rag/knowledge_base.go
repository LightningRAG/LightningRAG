package rag

import (
	"github.com/gin-gonic/gin"
)

func (r *KnowledgeBaseRouter) InitKnowledgeBaseRouter(Router *gin.RouterGroup) {
	kbRouter := Router.Group("rag/knowledgeBase")
	{
		kbRouter.POST("create", knowledgeBaseApi.Create)
		kbRouter.POST("list", knowledgeBaseApi.List)
		kbRouter.POST("get", knowledgeBaseApi.Get)
		kbRouter.POST("update", knowledgeBaseApi.Update)
		kbRouter.POST("delete", knowledgeBaseApi.Delete)
		kbRouter.POST("listDocuments", knowledgeBaseApi.ListDocuments)
		kbRouter.POST("uploadDocument", knowledgeBaseApi.UploadDocument)
		kbRouter.POST("getDocument", knowledgeBaseApi.GetDocument)
		kbRouter.POST("deleteDocument", knowledgeBaseApi.DeleteDocument)
		kbRouter.POST("batchDeleteDocuments", knowledgeBaseApi.BatchDeleteDocuments)
		kbRouter.POST("batchReindexDocuments", knowledgeBaseApi.BatchReindexDocuments)
		kbRouter.POST("batchCancelDocumentIndexing", knowledgeBaseApi.BatchCancelDocumentIndexing)
		kbRouter.POST("batchSetDocumentRetrieval", knowledgeBaseApi.BatchSetDocumentRetrieval)
		kbRouter.POST("batchSetDocumentPriority", knowledgeBaseApi.BatchSetDocumentPriority)
		kbRouter.POST("retryDocument", knowledgeBaseApi.RetryDocument)
		kbRouter.GET("downloadDocument", knowledgeBaseApi.DownloadDocument)
		kbRouter.POST("listChunks", knowledgeBaseApi.ListChunks)
		kbRouter.POST("retrieve", knowledgeBaseApi.Retrieve)
		kbRouter.POST("knowledgeGraph", knowledgeBaseApi.KnowledgeGraph)
		kbRouter.POST("updateChunk", knowledgeBaseApi.UpdateChunk)
		kbRouter.POST("share", knowledgeBaseApi.Share)
		kbRouter.POST("transfer", knowledgeBaseApi.Transfer)
		kbRouter.POST("listEmbeddingProviders", knowledgeBaseApi.ListEmbeddingProviders)
		kbRouter.POST("listVectorStoreConfigs", knowledgeBaseApi.ListVectorStoreConfigs)
		kbRouter.POST("listFileStorageConfigs", knowledgeBaseApi.ListFileStorageConfigs)
	}
}
