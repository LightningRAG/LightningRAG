package oauthapp

import (
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	"gorm.io/gorm"
)

type SysOAuthSettingService struct{}

var SysOAuthSettingServiceApp = new(SysOAuthSettingService)

// OAuthCallbackPathPattern 返回 OAuth 回调路径模板（含 system.router-prefix，不含域名），供 IdP 与前端拼接公网 URL。
func OAuthCallbackPathPattern() string {
	p := strings.TrimSpace(global.LRAG_CONFIG.System.RouterPrefix)
	p = strings.TrimSuffix(p, "/")
	suffix := "/base/oauth/callback/{kind}"
	if p == "" {
		return suffix
	}
	return p + suffix
}

// EffectiveOAuthFrontendRedirect 来自数据库全局设置；未配置时使用内置默认（开发用）。
func EffectiveOAuthFrontendRedirect() string {
	fe := strings.TrimSpace(global.OAuthFrontendRedirectFromDB())
	if fe == "" {
		fe = "http://127.0.0.1:8080/#/login"
	}
	return fe
}

// ReloadRuntime 从库刷新到进程内存（启动与更新后调用）
func (s *SysOAuthSettingService) ReloadRuntime() {
	if global.LRAG_DB == nil {
		global.SetOAuthRuntime("", "")
		return
	}
	var row system.SysOAuthSetting
	if err := global.LRAG_DB.First(&row, system.SysOAuthSettingSingletonID).Error; err != nil {
		global.SetOAuthRuntime("", "")
		return
	}
	global.SetOAuthRuntime(row.FrontendRedirect, row.SecretKey)
}

// EnsureSingleton 迁移后保证存在 id=1 记录
func (s *SysOAuthSettingService) EnsureSingleton() error {
	if global.LRAG_DB == nil {
		return nil
	}
	var n int64
	global.LRAG_DB.Model(&system.SysOAuthSetting{}).Where("id = ?", system.SysOAuthSettingSingletonID).Count(&n)
	if n > 0 {
		return nil
	}
	return global.LRAG_DB.Create(&system.SysOAuthSetting{
		ID: system.SysOAuthSettingSingletonID,
	}).Error
}

func (s *SysOAuthSettingService) GetAdmin() (systemRes.SysOAuthSettingAdmin, error) {
	if global.LRAG_DB == nil {
		return systemRes.SysOAuthSettingAdmin{
			CallbackPathPattern: OAuthCallbackPathPattern(),
		}, nil
	}
	if err := s.EnsureSingleton(); err != nil {
		return systemRes.SysOAuthSettingAdmin{}, err
	}
	var row system.SysOAuthSetting
	if err := global.LRAG_DB.First(&row, system.SysOAuthSettingSingletonID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return systemRes.SysOAuthSettingAdmin{
				CallbackPathPattern: OAuthCallbackPathPattern(),
			}, nil
		}
		return systemRes.SysOAuthSettingAdmin{}, err
	}
	return systemRes.SysOAuthSettingAdmin{
		FrontendRedirect:    strings.TrimSpace(row.FrontendRedirect),
		SecretKeyConfigured: strings.TrimSpace(row.SecretKey) != "",
		CallbackPathPattern: OAuthCallbackPathPattern(),
	}, nil
}

func (s *SysOAuthSettingService) Update(req systemReq.SysOAuthSettingUpdate) error {
	if global.LRAG_DB == nil {
		return i18n.NewError("svc.oauth.db_not_initialized")
	}
	if err := s.EnsureSingleton(); err != nil {
		return err
	}
	var row system.SysOAuthSetting
	if err := global.LRAG_DB.First(&row, system.SysOAuthSettingSingletonID).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{
		"frontend_redirect": strings.TrimSpace(req.FrontendRedirect),
	}
	if strings.TrimSpace(req.SecretKey) != "" {
		updates["secret_key"] = strings.TrimSpace(req.SecretKey)
	}
	if err := global.LRAG_DB.Model(&row).Updates(updates).Error; err != nil {
		return err
	}
	s.ReloadRuntime()
	return nil
}
