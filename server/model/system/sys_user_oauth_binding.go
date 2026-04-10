package system

import "github.com/LightningRAG/LightningRAG/server/global"

// SysUserOAuthBinding 系统用户与第三方账号绑定
type SysUserOAuthBinding struct {
	global.LRAG_MODEL
	UserID       uint   `json:"userId" gorm:"index;comment:系统用户ID"`
	ProviderKind string `json:"providerKind" gorm:"size:32;uniqueIndex:oauth_user_bind;comment:平台 kind"`
	Subject      string `json:"subject" gorm:"size:256;uniqueIndex:oauth_user_bind;comment:平台用户唯一标识"`
}

func (SysUserOAuthBinding) TableName() string {
	return "sys_user_oauth_bindings"
}
