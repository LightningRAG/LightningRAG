package response

import "time"

// OAuthExchangeData 一次性换票响应：与 LoginResponse 字段平级，并可选前端路由回跳
type OAuthExchangeData struct {
	LoginResponse
	Redirect string `json:"redirect,omitempty"`
}

// SysOAuthProviderAdmin 管理端列表/详情（不含密钥明文）
type SysOAuthProviderAdmin struct {
	ID                 uint                   `json:"ID"`
	CreatedAt          time.Time              `json:"CreatedAt"`
	UpdatedAt          time.Time              `json:"UpdatedAt"`
	Kind               string                 `json:"kind"`
	Enabled            bool                   `json:"enabled"`
	DisplayName        string                 `json:"displayName"`
	ButtonIcon         string                 `json:"buttonIcon"`
	ButtonIconPreview  string                 `json:"buttonIconPreview"`
	ClientID           string                 `json:"clientId"`
	ClientSecretSet    bool                   `json:"clientSecretSet"`
	Scopes             string                 `json:"scopes"`
	Extra              map[string]interface{} `json:"extra"`
	DefaultAuthorityID uint                   `json:"defaultAuthorityId"`
}

// OAuthPublicProvider 登录页展示的已启用平台
type OAuthPublicProvider struct {
	Kind        string `json:"kind"`
	DisplayName string `json:"displayName"`
	ButtonIcon  string `json:"buttonIcon"`
}

// OAuthRegisteredKind 代码中已注册的平台（用于后台下拉）
type OAuthRegisteredKind struct {
	Kind              string `json:"kind"`
	DisplayName       string `json:"displayName"`
	DefaultButtonIcon string `json:"defaultButtonIcon"`
}
