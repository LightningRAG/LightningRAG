package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	ragservice "github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/service/system"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"go.uber.org/zap"
)

// Reload 优雅地重新加载系统配置
func Reload() error {
	global.LRAG_LOG.Info("正在重新加载系统配置...")

	// 重新加载配置文件
	if err := global.LRAG_VP.ReadInConfig(); err != nil {
		global.LRAG_LOG.Error("重新读取配置文件失败!", zap.Error(err))
		return err
	}

	// 重新初始化数据库连接
	if global.LRAG_DB != nil {
		db, _ := global.LRAG_DB.DB()
		err := db.Close()
		if err != nil {
			global.LRAG_LOG.Error("关闭原数据库连接失败!", zap.Error(err))
			return err
		}
	}

	// 重新建立数据库连接
	global.LRAG_DB = Gorm()

	// 重新初始化其他配置
	OtherInit()
	DBList()

	if global.LRAG_DB != nil {
		RegisterTables()
		EnsureBuiltinRBACDataInDB()
		LoadOAuthGlobalFromDB()
		utils.ResetCasbinEnforcer()
		_ = system.CasbinServiceApp.FreshCasbin()
		ragservice.ResumeIncompleteDocumentJobs()
	}

	// 重新初始化定时任务
	Timer()

	global.LRAG_LOG.Info("系统配置重新加载完成")
	return nil
}
