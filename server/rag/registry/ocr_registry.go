package registry

import (
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// OCRFactory 创建 OCR 的工厂函数
type OCRFactory func(config OCRConfig) (interfaces.OCR, error)

// OCRConfig OCR 配置
type OCRConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
	Extra     map[string]any
}

// OCRRegistry OCR 注册表
type OCRRegistry struct {
	mu  sync.RWMutex
	fns map[string]OCRFactory
}

var defaultOCRRegistry = &OCRRegistry{fns: make(map[string]OCRFactory)}

// RegisterOCR 注册 OCR 提供商
func RegisterOCR(provider string, factory OCRFactory) {
	defaultOCRRegistry.mu.Lock()
	defer defaultOCRRegistry.mu.Unlock()
	defaultOCRRegistry.fns[provider] = factory
}

// CreateOCR 根据配置创建 OCR
func CreateOCR(config OCRConfig) (interfaces.OCR, error) {
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	defaultOCRRegistry.mu.RLock()
	fn, ok := defaultOCRRegistry.fns[provider]
	defaultOCRRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config)
}

// ListOCRProviders 返回已注册的 OCR 提供商名称列表
func ListOCRProviders() []string {
	defaultOCRRegistry.mu.RLock()
	defer defaultOCRRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultOCRRegistry.fns))
	for k := range defaultOCRRegistry.fns {
		names = append(names, k)
	}
	return names
}
