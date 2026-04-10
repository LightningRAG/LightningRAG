package oauthapp

import "github.com/LightningRAG/LightningRAG/server/service"

// ApiPack 聚合第三方快捷登录相关 API，供 api/v1.ApiGroup 嵌入一层命名空间。
type ApiPack struct {
	OAuthPublicApi
	OAuthProviderApi
	OAuthSettingApi
}

var loginLogService = service.ServiceGroupApp.SystemServiceGroup.LoginLogService
