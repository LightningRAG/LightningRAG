package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// SysOAuthProvider 第三方 OAuth 平台配置（每种 kind 至多一条）。读写与登录流程见 server/service/oauthapp。
type SysOAuthProvider struct {
	global.LRAG_MODEL
	Kind               string         `json:"kind" gorm:"uniqueIndex;size:32;comment:提供商标识 github/google/..."`
	Enabled            bool           `json:"enabled" gorm:"default:false;comment:是否启用"`
	DisplayName        string         `json:"displayName" gorm:"size:64;comment:展示名称"`
	ButtonIcon         string         `json:"buttonIcon" gorm:"column:button_icon;type:mediumtext;comment:登录按钮小图标：图片 URL 或 data:image/*;base64,..."`
	ClientID           string         `json:"clientId" gorm:"size:512;comment:客户端ID"`
	ClientSecretEnc    string         `json:"-" gorm:"column:client_secret_enc;size:1024;comment:客户端密钥密文"`
	Scopes             string         `json:"scopes" gorm:"size:512;comment:空格分隔Scope，空则用内置默认"`
	Extra              common.JSONMap `json:"extra" gorm:"type:text;comment:扩展 JSON，如 microsoft 的 tenant"`
	DefaultAuthorityID uint           `json:"defaultAuthorityId" gorm:"default:888;comment:自动注册用户的角色ID"`
}

func (SysOAuthProvider) TableName() string {
	return "sys_oauth_providers"
}
