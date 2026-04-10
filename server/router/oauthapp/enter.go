package oauthapp

import api "github.com/LightningRAG/LightningRAG/server/api/v1"

var (
	oauthProviderApi = api.ApiGroupApp.OAuthAppGroup.OAuthProviderApi
	oauthSettingApi  = api.ApiGroupApp.OAuthAppGroup.OAuthSettingApi
)

type RouterGroup struct {
	OAuthProviderRouter
	OAuthSettingRouter
}
