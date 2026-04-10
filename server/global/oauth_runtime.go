package global

import (
	"strings"
	"sync"
)

// OAuth 全局运行时配置（来自数据库 sys_oauth_settings，启动时与后台更新时刷新）
var (
	oauthRuntimeMu            sync.RWMutex
	oauthRuntimeFrontend      string
	oauthRuntimeSecretKey     string
)

// SetOAuthRuntime 由 service/oauthapp.SysOAuthSettingService 在加载/更新后写入；secret 为空表示未在库中配置主密钥。
func SetOAuthRuntime(frontendRedirect, secretKey string) {
	oauthRuntimeMu.Lock()
	defer oauthRuntimeMu.Unlock()
	oauthRuntimeFrontend = strings.TrimSpace(frontendRedirect)
	oauthRuntimeSecretKey = strings.TrimSpace(secretKey)
}

// OAuthFrontendRedirectFromDB 仅返回库中配置的前端回跳地址。
func OAuthFrontendRedirectFromDB() string {
	oauthRuntimeMu.RLock()
	defer oauthRuntimeMu.RUnlock()
	return oauthRuntimeFrontend
}

// OAuthSecretKeyFromDB 仅返回库中配置的加密主密钥材料。
func OAuthSecretKeyFromDB() string {
	oauthRuntimeMu.RLock()
	defer oauthRuntimeMu.RUnlock()
	return oauthRuntimeSecretKey
}
