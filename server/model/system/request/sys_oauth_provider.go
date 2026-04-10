package request

import "github.com/LightningRAG/LightningRAG/server/model/common"

// SysOAuthProviderCreate 新建 OAuth 配置
type SysOAuthProviderCreate struct {
	Kind               string         `json:"kind" binding:"required"`
	Enabled            bool           `json:"enabled"`
	DisplayName        string         `json:"displayName"`
	ButtonIcon         string         `json:"buttonIcon"`
	ClientID           string         `json:"clientId" binding:"required"`
	ClientSecret       string         `json:"clientSecret" binding:"required"`
	Scopes             string         `json:"scopes"`
	Extra              common.JSONMap `json:"extra"`
	DefaultAuthorityID uint           `json:"defaultAuthorityId"`
}

// SysOAuthProviderUpdate 更新（clientSecret 为空表示不修改原密钥）
type SysOAuthProviderUpdate struct {
	ID                 uint           `json:"ID" binding:"required"`
	Enabled            *bool          `json:"enabled"`
	DisplayName        string         `json:"displayName"`
	ButtonIcon         *string        `json:"buttonIcon,omitempty"`
	ClientID           string         `json:"clientId"`
	ClientSecret       string         `json:"clientSecret"`
	Scopes             string         `json:"scopes"`
	Extra              common.JSONMap `json:"extra"`
	DefaultAuthorityID *uint          `json:"defaultAuthorityId"`
}

// SysOAuthProviderSearch 列表查询
type SysOAuthProviderSearch struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Kind     string `json:"kind" form:"kind"`
	Enabled  *bool  `json:"enabled" form:"enabled"`
}
