package system

import "time"

const SysOAuthSettingSingletonID uint = 1

// SysOAuthSetting OAuth 全局设置（单行，主键固定为 1；无软删，避免误删导致配置丢失）。业务见 server/service/oauthapp。
type SysOAuthSetting struct {
	ID               uint      `json:"ID" gorm:"primaryKey;comment:固定为1"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
	FrontendRedirect string    `json:"frontendRedirect" gorm:"size:512;column:frontend_redirect;comment:OAuth完成后浏览器跳转，须含hash"`
	SecretKey        string    `json:"-" gorm:"size:1024;column:secret_key;comment:加密存储各平台client_secret的主密钥材料"`
}

func (SysOAuthSetting) TableName() string {
	return "sys_oauth_settings"
}
