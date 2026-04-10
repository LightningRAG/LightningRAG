package rag

type ServiceGroup struct {
	KnowledgeBaseService
	ConversationService
	LLMProviderService
	AgentService
	ChannelConnectorService
	SystemModelService
}

// 服务实现为 struct，嵌入到 ServiceGroup
type (
	KnowledgeBaseService struct{}
	ConversationService  struct{}
	LLMProviderService   struct{}
)

// llmProviderService 供 ConversationService 等调用默认模型解析
var llmProviderService = &LLMProviderService{}

// systemModelService 供 LLMProviderService 等调用系统默认配置
var systemModelService = &SystemModelService{}
