package registry

import (
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// EmbeddingRegistry 嵌入模型注册表
type EmbeddingRegistry struct {
	mu  sync.RWMutex
	fns map[string]EmbeddingFactory
}

// EmbeddingFactory 创建 Embedder 的工厂函数
type EmbeddingFactory func(config EmbeddingConfig) (interfaces.Embedder, error)

// EmbeddingConfig 嵌入模型配置
type EmbeddingConfig struct {
	Provider   string
	ModelName  string
	BaseURL    string
	APIKey     string
	Dimensions int
	Extra      map[string]any
}

var defaultEmbeddingRegistry = &EmbeddingRegistry{fns: make(map[string]EmbeddingFactory)}

// RegisterEmbedding 注册嵌入模型提供商
func RegisterEmbedding(provider string, factory EmbeddingFactory) {
	defaultEmbeddingRegistry.mu.Lock()
	defer defaultEmbeddingRegistry.mu.Unlock()
	defaultEmbeddingRegistry.fns[provider] = factory
}

// CreateEmbedding 根据配置创建 Embedder
func CreateEmbedding(config EmbeddingConfig) (interfaces.Embedder, error) {
	defaultEmbeddingRegistry.mu.RLock()
	fn, ok := defaultEmbeddingRegistry.fns[config.Provider]
	defaultEmbeddingRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config)
}

// ListEmbeddingProviders 返回已注册的 Embedding 提供商名称列表
func ListEmbeddingProviders() []string {
	defaultEmbeddingRegistry.mu.RLock()
	defer defaultEmbeddingRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultEmbeddingRegistry.fns))
	for k := range defaultEmbeddingRegistry.fns {
		names = append(names, k)
	}
	return names
}
