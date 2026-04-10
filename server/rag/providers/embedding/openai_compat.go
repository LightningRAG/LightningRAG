package embedding

import (
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// EmbedderNameAdapter 包装 Embedder 以自定义 ProviderName，用于同一实现多厂商注册
type EmbedderNameAdapter struct {
	interfaces.Embedder
	DisplayName string
}

func (e *EmbedderNameAdapter) ProviderName() string {
	if e.DisplayName != "" {
		return e.DisplayName
	}
	return e.Embedder.ProviderName()
}

// NewOpenAICompatEmbed 创建 OpenAI 兼容的 Embedder（可指定显示名称）
// 用于 DeepSeek、Xinference、LocalAI、SiliconFlow 等 OpenAI 兼容 API
func NewOpenAICompatEmbed(displayName, apiKey, baseURL, model string, dimensions int) *EmbedderNameAdapter {
	impl := NewOpenAIEmbed(apiKey, baseURL, model, dimensions)
	return &EmbedderNameAdapter{
		Embedder:    impl,
		DisplayName: displayName,
	}
}
