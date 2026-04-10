package registry

import (
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// Speech2TextFactory 创建 Speech2Text 的工厂函数
type Speech2TextFactory func(config Speech2TextConfig) (interfaces.Speech2Text, error)

// Speech2TextConfig 语音转文字配置
type Speech2TextConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
	Extra     map[string]any
}

// Speech2TextRegistry 语音转文字注册表
type Speech2TextRegistry struct {
	mu  sync.RWMutex
	fns map[string]Speech2TextFactory
}

var defaultSpeech2TextRegistry = &Speech2TextRegistry{fns: make(map[string]Speech2TextFactory)}

// RegisterSpeech2Text 注册语音转文字提供商
func RegisterSpeech2Text(provider string, factory Speech2TextFactory) {
	defaultSpeech2TextRegistry.mu.Lock()
	defer defaultSpeech2TextRegistry.mu.Unlock()
	defaultSpeech2TextRegistry.fns[provider] = factory
}

// CreateSpeech2Text 根据配置创建 Speech2Text
func CreateSpeech2Text(config Speech2TextConfig) (interfaces.Speech2Text, error) {
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	defaultSpeech2TextRegistry.mu.RLock()
	fn, ok := defaultSpeech2TextRegistry.fns[provider]
	defaultSpeech2TextRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config)
}

// ListSpeech2TextProviders 返回已注册的 Speech2Text 提供商名称列表
func ListSpeech2TextProviders() []string {
	defaultSpeech2TextRegistry.mu.RLock()
	defer defaultSpeech2TextRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultSpeech2TextRegistry.fns))
	for k := range defaultSpeech2TextRegistry.fns {
		names = append(names, k)
	}
	return names
}
