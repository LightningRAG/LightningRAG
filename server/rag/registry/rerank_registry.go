package registry

import (
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// RerankFactory 创建 Reranker 的工厂函数
type RerankFactory func(config RerankConfig) (interfaces.Reranker, error)

// RerankConfig 重排序模型配置
type RerankConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
	Extra     map[string]any
}

// RerankRegistry 重排序模型注册表
type RerankRegistry struct {
	mu  sync.RWMutex
	fns map[string]RerankFactory
}

var defaultRerankRegistry = &RerankRegistry{fns: make(map[string]RerankFactory)}

// RegisterRerank 注册重排序模型提供商
func RegisterRerank(provider string, factory RerankFactory) {
	defaultRerankRegistry.mu.Lock()
	defer defaultRerankRegistry.mu.Unlock()
	defaultRerankRegistry.fns[provider] = factory
}

// CreateRerank 根据配置创建 Reranker
func CreateRerank(config RerankConfig) (interfaces.Reranker, error) {
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	defaultRerankRegistry.mu.RLock()
	fn, ok := defaultRerankRegistry.fns[provider]
	defaultRerankRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config)
}

// ListRerankProviders 返回已注册的 Rerank 提供商名称列表
func ListRerankProviders() []string {
	defaultRerankRegistry.mu.RLock()
	defer defaultRerankRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultRerankRegistry.fns))
	for k := range defaultRerankRegistry.fns {
		names = append(names, k)
	}
	return names
}
