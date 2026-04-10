package rag

import (
	api "github.com/LightningRAG/LightningRAG/server/api/v1"
)

type (
	KnowledgeBaseRouter    struct{}
	ConversationRouter     struct{}
	LLMProviderRouter      struct{}
	AgentRouter            struct{}
	ChannelConnectorRouter struct{}
	SettingsRouter         struct{}
	SystemModelRouter      struct{}
)

type RouterGroup struct {
	KnowledgeBaseRouter
	ConversationRouter
	LLMProviderRouter
	AgentRouter
	ChannelConnectorRouter
	SettingsRouter
	SystemModelRouter
}

var (
	knowledgeBaseApi    = api.ApiGroupApp.RagApiGroup.KnowledgeBaseApi
	conversationApi     = api.ApiGroupApp.RagApiGroup.ConversationApi
	llmProviderApi      = api.ApiGroupApp.RagApiGroup.LLMProviderApi
	agentApi            = api.ApiGroupApp.RagApiGroup.AgentApi
	channelConnectorApi = api.ApiGroupApp.RagApiGroup.ChannelConnectorApi
	settingsApi         = api.ApiGroupApp.RagApiGroup.SettingsApi
	systemModelApi      = api.ApiGroupApp.RagApiGroup.SystemModelApi
)
