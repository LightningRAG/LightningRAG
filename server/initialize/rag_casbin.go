package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	adapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
)

// ragCasbinRules 需要确保存在的 RAG 相关 Casbin 权限规则
// 与 source/system/casbin.go 和 rag_api.go 保持一致
var ragCasbinRules = []adapter.CasbinRule{
	// RAG 知识库
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/create", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/list", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/get", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/update", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/delete", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/listDocuments", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/uploadDocument", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/getDocument", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/deleteDocument", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/batchDeleteDocuments", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/batchReindexDocuments", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/batchCancelDocumentIndexing", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/batchSetDocumentRetrieval", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/batchSetDocumentPriority", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/retryDocument", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/downloadDocument", V2: "GET"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/listChunks", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/retrieve", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/knowledgeGraph", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/updateChunk", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/share", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/transfer", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/listEmbeddingProviders", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/listVectorStoreConfigs", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/knowledgeBase/listFileStorageConfigs", V2: "POST"},
	// RAG 设置
	{Ptype: "p", V0: "888", V1: "/rag/settings/vectorStore/create", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/vectorStore/update", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/vectorStore/delete", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/vectorStore/get", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/vectorStore/list", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/fileStorage/create", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/fileStorage/update", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/fileStorage/delete", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/fileStorage/get", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/settings/fileStorage/list", V2: "POST"},
	// RAG 对话
	{Ptype: "p", V0: "888", V1: "/rag/conversation/create", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/chat", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/chatStream", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/queryData", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/list", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/get", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/update", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/listMessages", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/listTools", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/conversation/delete", V2: "POST"},
	// RAG 模型
	{Ptype: "p", V0: "888", V1: "/rag/llm/listProviders", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/listAvailableProviders", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/listUserModels", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/addUserModel", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/updateUserModel", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/deleteUserModel", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/setAuthorityDefaultLLM", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/getAuthorityDefaultLLMs", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/setUserDefaultLLM", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/getUserDefaultLLMs", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/clearAuthorityDefaultLLM", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/clearUserDefaultLLM", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/listWebSearchProviders", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/getWebSearchConfig", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/llm/setWebSearchConfig", V2: "POST"},
	// RAG Agent
	{Ptype: "p", V0: "888", V1: "/rag/agent/run", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/runStream", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/templates", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/loadTemplate", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/create", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/list", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/get", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/update", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/delete", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/agent/createFromTemplate", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/create", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/update", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/list", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/channelTypes", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/get", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/delete", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/outbound/list", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/outbound/delete", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/channelConnector/outbound/runOnce", V2: "POST"},
	// RAG 系统模型
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/listAdminModels", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/createAdminModel", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/updateAdminModel", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/deleteAdminModel", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/getSystemDefaults", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/setSystemDefault", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/clearSystemDefault", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/listSystemWebSearchProviders", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/getSystemWebSearchConfig", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/setSystemWebSearchConfig", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/clearSystemWebSearchConfig", V2: "POST"},
	// 全局共享知识库
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/listGlobalKnowledgeBases", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/setGlobalKnowledgeBase", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/removeGlobalKnowledgeBase", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/rag/systemModel/listAllKnowledgeBases", V2: "POST"},
}

// EnsureRagCasbin 确保 RAG 相关 Casbin 权限规则已注册到 casbin_rule 表。
// 同时清理历史遗留的重复规则。
func EnsureRagCasbin() {
	if global.LRAG_DB == nil {
		return
	}
	deduplicateCasbinRules()
	for _, rule := range ragCasbinRules {
		ruleCopy := rule
		if err := global.LRAG_DB.
			Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).
			FirstOrCreate(&ruleCopy).Error; err != nil {
			global.LRAG_LOG.Error("EnsureRagCasbin: create rule failed", zap.String("path", rule.V1), zap.Error(err))
		}
	}
}

// deduplicateCasbinRules 清理 casbin_rule 中 (ptype,v0,v1,v2) 重复的记录，每组只保留 ID 最小的一条
func deduplicateCasbinRules() {
	type dupRow struct {
		Ptype string `gorm:"column:ptype"`
		V0    string `gorm:"column:v0"`
		V1    string `gorm:"column:v1"`
		V2    string `gorm:"column:v2"`
		Cnt   int    `gorm:"column:cnt"`
		MinID uint   `gorm:"column:min_id"`
	}
	var dups []dupRow
	if err := global.LRAG_DB.Raw(
		"SELECT ptype, v0, v1, v2, COUNT(*) as cnt, MIN(id) as min_id FROM casbin_rule GROUP BY ptype, v0, v1, v2 HAVING COUNT(*) > 1",
	).Scan(&dups).Error; err != nil {
		global.LRAG_LOG.Debug("deduplicateCasbinRules: query failed", zap.Error(err))
		return
	}
	for _, d := range dups {
		if err := global.LRAG_DB.Exec(
			"DELETE FROM casbin_rule WHERE ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND id != ?",
			d.Ptype, d.V0, d.V1, d.V2, d.MinID,
		).Error; err != nil {
			global.LRAG_LOG.Debug("deduplicateCasbinRules: delete duplicates failed", zap.String("v1", d.V1), zap.Error(err))
		} else if d.Cnt > 1 {
			global.LRAG_LOG.Info("deduplicateCasbinRules: removed duplicates", zap.String("v0", d.V0), zap.String("v1", d.V1), zap.Int("removed", d.Cnt-1))
		}
	}
}
