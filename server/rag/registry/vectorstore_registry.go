package registry

import (
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// VectorStoreConfig 向量存储配置，从 RagVectorStoreConfig 解析
type VectorStoreConfig struct {
	Provider string         // postgresql | elasticsearch | 后续可扩展 pgvector, milvus 等
	Config   map[string]any // 连接配置 JSON，如 address, username, password
}

// VectorStoreFactory 创建 VectorStore 的工厂函数
type VectorStoreFactory func(config VectorStoreConfig, embedder interfaces.Embedder, namespace string, vectorDims int) (interfaces.VectorStore, error)

// VectorStoreRegistry 向量存储注册表
type VectorStoreRegistry struct {
	mu  sync.RWMutex
	fns map[string]VectorStoreFactory
}

var defaultVectorStoreRegistry = &VectorStoreRegistry{fns: make(map[string]VectorStoreFactory)}

// RegisterVectorStore 注册向量存储提供商
func RegisterVectorStore(provider string, factory VectorStoreFactory) {
	defaultVectorStoreRegistry.mu.Lock()
	defer defaultVectorStoreRegistry.mu.Unlock()
	defaultVectorStoreRegistry.fns[strings.ToLower(provider)] = factory
}

// CreateVectorStore 根据配置创建 VectorStore
func CreateVectorStore(config VectorStoreConfig, embedder interfaces.Embedder, namespace string, vectorDims int) (interfaces.VectorStore, error) {
	defaultVectorStoreRegistry.mu.RLock()
	fn, ok := defaultVectorStoreRegistry.fns[strings.ToLower(config.Provider)]
	defaultVectorStoreRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config, embedder, namespace, vectorDims)
}

// ListVectorStoreProviders 返回已注册的向量存储提供商名称列表
func ListVectorStoreProviders() []string {
	defaultVectorStoreRegistry.mu.RLock()
	defer defaultVectorStoreRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultVectorStoreRegistry.fns))
	for k := range defaultVectorStoreRegistry.fns {
		names = append(names, k)
	}
	return names
}
