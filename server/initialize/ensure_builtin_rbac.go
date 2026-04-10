package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	ragservice "github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/service/system"
)

func init() {
	system.RegisterPostInitDBCallback(func() {
		if global.LRAG_DB == nil {
			return
		}
		// 与 main/core 在「启动时 DB 已连接」路径下等价的补全（向导 InitDB 前 LRAG_DB 常为 nil，这些步骤曾被跳过）
		EnsureBuiltinRBACDataInDB()
		LoadOAuthGlobalFromDB()
		system.LoadAll()
		TryDeferredPluginInstall()
		ragservice.ResumeIncompleteDocumentJobs()
		_ = system.CasbinServiceApp.FreshCasbin()
	})
}

// EnsureBuiltinRBACDataInDB 将 RAG、OAuth 等扩展的 sys_api、Casbin 规则、菜单及菜单 i18n 写入数据库。
//
// 典型场景：Docker 首次启动时 config 中数据库名为空，global.LRAG_DB 为 nil，Routers() 阶段会跳过
// EnsureRagCasbin 等写入；InitDB 向导写入的种子数据不含 RAG Casbin（见 source/system/casbin.go 注释），
// 故在 InitDB 成功后必须再执行本函数，否则角色 888 等用户访问 /rag/* 会全部被 Casbin 拒绝。
func EnsureBuiltinRBACDataInDB() {
	if global.LRAG_DB == nil {
		return
	}
	EnsureRagApis()
	EnsureRagCasbin()
	EnsureRagMenus()
	EnsureOAuthApis()
	EnsureOAuthCasbin()
	EnsureOAuthQuickLoginMenu()
	SyncMenuTitleKeys()
	SyncMenuEnglishDefaultTitles()
}
