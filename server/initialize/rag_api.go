package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	"go.uber.org/zap"
)

// ragApis 需要确保存在的 RAG 相关 API（与 source/system/api.go 保持一致）
var ragApis = []system.SysApi{
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/create", Description: "Create knowledge base"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/list", Description: "List knowledge bases"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/get", Description: "Get knowledge base"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/update", Description: "Update knowledge base"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/delete", Description: "Delete knowledge base"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/listDocuments", Description: "List documents"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/uploadDocument", Description: "Upload document"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/getDocument", Description: "Get document"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/deleteDocument", Description: "Delete document"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/batchDeleteDocuments", Description: "Batch delete documents"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/batchReindexDocuments", Description: "Batch reindex documents"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/batchCancelDocumentIndexing", Description: "Batch cancel indexing"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/batchSetDocumentRetrieval", Description: "Batch enable or disable document retrieval"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/batchSetDocumentPriority", Description: "Batch set document retrieval priority 0-1"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/retryDocument", Description: "Retry document parsing"},
	{ApiGroup: "RAG knowledge base", Method: "GET", Path: "/rag/knowledgeBase/downloadDocument", Description: "Download document"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/listChunks", Description: "List document chunks"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/retrieve", Description: "Retrieve chunks"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/knowledgeGraph", Description: "Knowledge graph visualization subset"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/updateChunk", Description: "Update chunk"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/share", Description: "Share knowledge base"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/transfer", Description: "Transfer knowledge base"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/listEmbeddingProviders", Description: "List embedding providers"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/listVectorStoreConfigs", Description: "List vector store configs"},
	{ApiGroup: "RAG knowledge base", Method: "POST", Path: "/rag/knowledgeBase/listFileStorageConfigs", Description: "List file storage configs"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/vectorStore/create", Description: "Create vector store config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/vectorStore/update", Description: "Update vector store config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/vectorStore/delete", Description: "Delete vector store config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/vectorStore/get", Description: "Get vector store config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/vectorStore/list", Description: "List vector store configs"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/fileStorage/create", Description: "Create file storage config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/fileStorage/update", Description: "Update file storage config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/fileStorage/delete", Description: "Delete file storage config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/fileStorage/get", Description: "Get file storage config"},
	{ApiGroup: "RAG settings", Method: "POST", Path: "/rag/settings/fileStorage/list", Description: "List file storage configs"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/create", Description: "Create conversation"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/chat", Description: "Chat"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/chatStream", Description: "Chat stream"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/queryData", Description: "Structured retrieval without LLM"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/list", Description: "List conversations"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/get", Description: "Get conversation"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/update", Description: "Update conversation"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/listMessages", Description: "List conversation messages"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/listTools", Description: "List available tools"},
	{ApiGroup: "RAG conversation", Method: "POST", Path: "/rag/conversation/delete", Description: "Delete conversation"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/listProviders", Description: "List providers"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/listAvailableProviders", Description: "List providers by scenario"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/listUserModels", Description: "List user models"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/addUserModel", Description: "Add user model"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/updateUserModel", Description: "Update user model"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/deleteUserModel", Description: "Delete user model"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/setAuthorityDefaultLLM", Description: "Set authority default LLM"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/getAuthorityDefaultLLMs", Description: "Get authority default LLMs"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/setUserDefaultLLM", Description: "Set user default LLM"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/getUserDefaultLLMs", Description: "Get user default LLMs"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/clearAuthorityDefaultLLM", Description: "Clear authority default LLM"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/clearUserDefaultLLM", Description: "Clear user default LLM"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/listWebSearchProviders", Description: "List web search providers"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/getWebSearchConfig", Description: "Get web search config"},
	{ApiGroup: "RAG LLM", Method: "POST", Path: "/rag/llm/setWebSearchConfig", Description: "Set web search config"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/run", Description: "Run agent"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/runStream", Description: "Run agent stream"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/templates", Description: "List agent templates"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/loadTemplate", Description: "Load agent template"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/create", Description: "Create agent"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/list", Description: "List agents"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/get", Description: "Get agent"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/update", Description: "Update agent"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/delete", Description: "Delete agent"},
	{ApiGroup: "RAG Agent", Method: "POST", Path: "/rag/agent/createFromTemplate", Description: "Create agent from template"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/create", Description: "Create third-party channel connector"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/update", Description: "Update channel connector"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/list", Description: "List channel connectors"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/channelTypes", Description: "List registered channel adapter kinds"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/get", Description: "Get channel connector"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/delete", Description: "Delete channel connector"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/outbound/list", Description: "List channel outbound retry queue"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/outbound/delete", Description: "Delete channel outbound queue row"},
	{ApiGroup: "RAG channel", Method: "POST", Path: "/rag/channelConnector/outbound/runOnce", Description: "Run one channel outbound retry batch"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/listAdminModels", Description: "List admin models"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/createAdminModel", Description: "Create admin model"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/updateAdminModel", Description: "Update admin model"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/deleteAdminModel", Description: "Delete admin model"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/getSystemDefaults", Description: "Get system default models"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/setSystemDefault", Description: "Set system default model"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/clearSystemDefault", Description: "Clear system default model"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/listSystemWebSearchProviders", Description: "List system web search providers"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/getSystemWebSearchConfig", Description: "Get system web search config"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/setSystemWebSearchConfig", Description: "Set system web search config"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/clearSystemWebSearchConfig", Description: "Clear system web search config"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/listGlobalKnowledgeBases", Description: "List global knowledge bases"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/setGlobalKnowledgeBase", Description: "Set global knowledge base"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/removeGlobalKnowledgeBase", Description: "Remove global knowledge base"},
	{ApiGroup: "RAG system model", Method: "POST", Path: "/rag/systemModel/listAllKnowledgeBases", Description: "List all knowledge bases"},
}

// EnsureRagApis 确保 RAG 相关 API 已注册到 sys_api 表，供权限配置使用。
// 同时清理历史遗留的重复记录（同一 path+method 保留 ID 最小的一条）。
func EnsureRagApis() {
	if global.LRAG_DB == nil {
		return
	}
	deduplicateSysApis()
	for _, api := range ragApis {
		apiCopy := api
		if err := global.LRAG_DB.Where("path = ? AND method = ?", apiCopy.Path, apiCopy.Method).
			FirstOrCreate(&apiCopy).Error; err != nil {
			global.LRAG_LOG.Warn("EnsureRagApis: FirstOrCreate failed",
				zap.String("path", apiCopy.Path),
				zap.String("method", apiCopy.Method),
				zap.Error(err))
		}
	}
}

// deduplicateSysApis 清理 sys_apis 中 path+method 重复的记录，每组只保留 ID 最小的一条
func deduplicateSysApis() {
	type dupRow struct {
		Path   string `gorm:"column:path"`
		Method string `gorm:"column:method"`
		Cnt    int    `gorm:"column:cnt"`
		MinID  uint   `gorm:"column:min_id"`
	}
	var dups []dupRow
	if err := global.LRAG_DB.Raw(
		"SELECT path, method, COUNT(*) as cnt, MIN(id) as min_id FROM sys_apis WHERE deleted_at IS NULL GROUP BY path, method HAVING COUNT(*) > 1",
	).Scan(&dups).Error; err != nil {
		global.LRAG_LOG.Debug("deduplicateSysApis: query failed", zap.Error(err))
		return
	}
	for _, d := range dups {
		if err := global.LRAG_DB.Exec(
			"DELETE FROM sys_apis WHERE path = ? AND method = ? AND deleted_at IS NULL AND id != ?",
			d.Path, d.Method, d.MinID,
		).Error; err != nil {
			global.LRAG_LOG.Debug("deduplicateSysApis: delete duplicates failed", zap.String("path", d.Path), zap.Error(err))
		} else {
			global.LRAG_LOG.Info("deduplicateSysApis: removed duplicates", zap.String("path", d.Path), zap.String("method", d.Method), zap.Int("removed", d.Cnt-1))
		}
	}
}
