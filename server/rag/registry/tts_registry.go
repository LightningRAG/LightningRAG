package registry

import (
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// TTSFactory 创建 TTS 的工厂函数
type TTSFactory func(config TTSConfig) (interfaces.TTS, error)

// TTSConfig 文字转语音配置
type TTSConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
	Extra     map[string]any
}

// TTSRegistry 文字转语音注册表
type TTSRegistry struct {
	mu  sync.RWMutex
	fns map[string]TTSFactory
}

var defaultTTSRegistry = &TTSRegistry{fns: make(map[string]TTSFactory)}

// RegisterTTS 注册文字转语音提供商
func RegisterTTS(provider string, factory TTSFactory) {
	defaultTTSRegistry.mu.Lock()
	defer defaultTTSRegistry.mu.Unlock()
	defaultTTSRegistry.fns[provider] = factory
}

// CreateTTS 根据配置创建 TTS
func CreateTTS(config TTSConfig) (interfaces.TTS, error) {
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	defaultTTSRegistry.mu.RLock()
	fn, ok := defaultTTSRegistry.fns[provider]
	defaultTTSRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return fn(config)
}

// ListTTSProviders 返回已注册的 TTS 提供商名称列表
func ListTTSProviders() []string {
	defaultTTSRegistry.mu.RLock()
	defer defaultTTSRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultTTSRegistry.fns))
	for k := range defaultTTSRegistry.fns {
		names = append(names, k)
	}
	return names
}
