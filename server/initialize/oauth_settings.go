package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/service/oauthapp"
	"go.uber.org/zap"
)

// LoadOAuthGlobalFromDB 在库表就绪后加载 OAuth 全局设置到进程内存（含 EnsureSingleton）
func LoadOAuthGlobalFromDB() {
	if global.LRAG_DB == nil {
		return
	}
	if err := oauthapp.SysOAuthSettingServiceApp.EnsureSingleton(); err != nil {
		global.LRAG_LOG.Warn("oauth global: ensure singleton", zap.Error(err))
	}
	oauthapp.SysOAuthSettingServiceApp.ReloadRuntime()
}
