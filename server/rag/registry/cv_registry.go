package registry

import (
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// CVFactory 创建 CV 的工厂函数
type CVFactory func(config CVConfig) (interfaces.CV, error)

// CVConfig 计算机视觉配置
type CVConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
	Extra     map[string]any
}

// CVRegistry 计算机视觉注册表
type CVRegistry struct {
	mu  sync.RWMutex
	fns map[string]CVFactory
}

var defaultCVRegistry = &CVRegistry{fns: make(map[string]CVFactory)}

// RegisterCV 注册计算机视觉提供商
func RegisterCV(provider string, factory CVFactory) {
	defaultCVRegistry.mu.Lock()
	defer defaultCVRegistry.mu.Unlock()
	defaultCVRegistry.fns[provider] = factory
}

// CreateCV 根据配置创建 CV
func CreateCV(config CVConfig) (interfaces.CV, error) {
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	defaultCVRegistry.mu.RLock()
	fn, ok := defaultCVRegistry.fns[provider]
	defaultCVRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config)
}

// ListCVProviders 返回已注册的 CV 提供商名称列表
func ListCVProviders() []string {
	defaultCVRegistry.mu.RLock()
	defer defaultCVRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultCVRegistry.fns))
	for k := range defaultCVRegistry.fns {
		names = append(names, k)
	}
	return names
}
