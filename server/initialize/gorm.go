package initialize

import (
	"errors"
	"os"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/example"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/system"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	switch global.LRAG_CONFIG.System.DbType {
	case "mysql":
		global.LRAG_ACTIVE_DBNAME = &global.LRAG_CONFIG.Mysql.Dbname
		return GormMysql()
	case "pgsql":
		global.LRAG_ACTIVE_DBNAME = &global.LRAG_CONFIG.Pgsql.Dbname
		return GormPgSql()
	case "oracle":
		global.LRAG_ACTIVE_DBNAME = &global.LRAG_CONFIG.Oracle.Dbname
		return GormOracle()
	case "mssql":
		global.LRAG_ACTIVE_DBNAME = &global.LRAG_CONFIG.Mssql.Dbname
		return GormMssql()
	case "sqlite":
		global.LRAG_ACTIVE_DBNAME = &global.LRAG_CONFIG.Sqlite.Dbname
		return GormSqlite()
	default:
		global.LRAG_ACTIVE_DBNAME = &global.LRAG_CONFIG.Mysql.Dbname
		return GormMysql()
	}
}

// AutoMigrateAllSchema 与 RegisterTables 的主迁移一致（不含 disable-auto-migrate 判断）。
// 供 InitDB（首次在页面初始化数据库）在仅执行 SubInitializer 建表时补齐 RAG 等表，避免必须重启进程才出现 rag_channel_outbounds 等结构。
func AutoMigrateAllSchema(db *gorm.DB) error {
	if db == nil {
		return errors.New("AutoMigrateAllSchema: db is nil")
	}
	return db.AutoMigrate(

		system.SysApi{},
		system.SysIgnoreApi{},
		system.SysUser{},
		system.SysBaseMenu{},
		system.JwtBlacklist{},
		system.SysAuthority{},
		system.SysDictionary{},
		system.SysOperationRecord{},
		system.SysAutoCodeHistory{},
		system.SysDictionaryDetail{},
		system.SysBaseMenuParameter{},
		system.SysBaseMenuBtn{},
		system.SysAuthorityBtn{},
		system.SysAutoCodePackage{},
		system.SysExportTemplate{},
		system.Condition{},
		system.JoinTemplate{},
		system.SysParams{},
		system.SysVersion{},
		system.SysError{},
		system.SysApiToken{},
		system.SysLoginLog{},
		system.SysOAuthProvider{},
		system.SysUserOAuthBinding{},
		system.SysOAuthSetting{},

		example.ExaFile{},
		example.ExaCustomer{},
		example.ExaFileChunk{},
		example.ExaFileUploadAndDownload{},
		example.ExaAttachmentCategory{},

		rag.RagKnowledgeBase{},
		rag.RagDocument{},
		rag.RagChunk{},
		rag.RagKgEntity{},
		rag.RagKgRelationship{},
		rag.RagKgEntityChunk{},
		rag.RagKgRelationshipChunk{},
		rag.RagLLMProvider{},
		rag.RagEmbeddingProvider{},
		rag.RagVectorStoreConfig{},
		rag.RagFileStorageConfig{},
		rag.RagUserLLM{},
		rag.RagKnowledgeBaseShare{},
		rag.RagConversation{},
		rag.RagMessage{},
		rag.RagAuthorityDefaultLLM{},
		rag.RagUserDefaultLLM{},
		rag.RagUserWebSearchConfig{},
		rag.RagAgent{},
		rag.RagSystemDefaultModel{},
		rag.RagSystemDefaultWebSearchConfig{},
		rag.RagGlobalKnowledgeBase{},
		rag.RagChannelConnector{},
		rag.RagChannelSession{},
		rag.RagChannelWebhookEvent{},
		rag.RagChannelOutbound{},
	)
}

func RegisterTables() {
	if global.LRAG_CONFIG.System.DisableAutoMigrate {
		global.LRAG_LOG.Info("auto-migrate is disabled, skipping table registration")
		return
	}

	db := global.LRAG_DB
	err := AutoMigrateAllSchema(db)
	if err != nil {
		global.LRAG_LOG.Error("register table failed", zap.Error(err))
		os.Exit(0)
	}

	err = bizModel()

	if err != nil {
		global.LRAG_LOG.Error("register biz_table failed", zap.Error(err))
		os.Exit(0)
	}
	global.LRAG_LOG.Info("register table success")
}
