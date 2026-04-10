package rag

import "github.com/LightningRAG/LightningRAG/server/service"

type ApiGroup struct {
	KnowledgeBaseApi
	ConversationApi
	LLMProviderApi
	AgentApi
	ChannelConnectorApi
	SettingsApi
	SystemModelApi
}

var (
	knowledgeBaseService    = service.ServiceGroupApp.RagServiceGroup.KnowledgeBaseService
	conversationService     = service.ServiceGroupApp.RagServiceGroup.ConversationService
	llmProviderService      = service.ServiceGroupApp.RagServiceGroup.LLMProviderService
	agentService            = service.ServiceGroupApp.RagServiceGroup.AgentService
	channelConnectorService = service.ServiceGroupApp.RagServiceGroup.ChannelConnectorService
	systemModelService      = service.ServiceGroupApp.RagServiceGroup.SystemModelService
)
