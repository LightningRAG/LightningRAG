package request

// SysOAuthSettingUpdate 更新全局 OAuth 设置（secretKey 为空表示不修改原密钥）
type SysOAuthSettingUpdate struct {
	FrontendRedirect string `json:"frontendRedirect"`
	SecretKey        string `json:"secretKey"`
}
