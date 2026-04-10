package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// LLMRegistry LLM 提供商注册表
type LLMRegistry struct {
	mu  sync.RWMutex
	fns map[string]LLMFactory
}

// LLMFactory 创建 LLM 的工厂函数
type LLMFactory func(config LLMConfig) (interfaces.LLM, error)

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider  string
	ModelName string
	BaseURL   string
	APIKey    string
	Extra     map[string]any
}

var defaultLLMRegistry = &LLMRegistry{fns: make(map[string]LLMFactory)}

// RegisterLLM 注册 LLM 提供商
func RegisterLLM(provider string, factory LLMFactory) {
	defaultLLMRegistry.mu.Lock()
	defer defaultLLMRegistry.mu.Unlock()
	defaultLLMRegistry.fns[provider] = factory
}

// CreateLLM 根据配置创建 LLM
func CreateLLM(config LLMConfig) (interfaces.LLM, error) {
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	defaultLLMRegistry.mu.RLock()
	fn, ok := defaultLLMRegistry.fns[provider]
	defaultLLMRegistry.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("未知的 LLM 提供商: %q，请检查 provider 或 llm_id 格式（provider@model）", provider)
	}
	return fn(config)
}

// MustCreateLLM 创建 LLM，失败则 panic
func MustCreateLLM(ctx context.Context, config LLMConfig) interfaces.LLM {
	llm, err := CreateLLM(config)
	if err != nil {
		panic(err)
	}
	return llm
}

// ListLLMProviders 返回已注册的 LLM 提供商名称列表
func ListLLMProviders() []string {
	defaultLLMRegistry.mu.RLock()
	defer defaultLLMRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultLLMRegistry.fns))
	for k := range defaultLLMRegistry.fns {
		names = append(names, k)
	}
	return names
}
