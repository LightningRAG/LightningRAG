package response

// SysOAuthSettingAdmin 管理端展示（不含密钥明文）
type SysOAuthSettingAdmin struct {
	FrontendRedirect    string `json:"frontendRedirect"`
	SecretKeyConfigured bool   `json:"secretKeyConfigured"`
	CallbackPathPattern string `json:"callbackPathPattern"` // 含 router-prefix，不含域名，如 /api/base/oauth/callback/{kind}
}
