// Package tools 网页搜索引擎接口与注册
package tools

import (
	"context"
	"sync"
)

// WebSearchProvider 网页搜索引擎接口，DuckDuckGo、百度等实现此接口
type WebSearchProvider interface {
	// ID 唯一标识，如 duckduckgo、baidu
	ID() string
	// DisplayName 显示名称
	DisplayName() string
	// ConfigSchema 返回该引擎需要的配置项（用于前端展示），如 [{"key":"apiKey","label":"API Key","required":true}]
	ConfigSchema() []WebSearchConfigField
	// Search 执行搜索，config 为用户配置的 key-value
	Search(ctx context.Context, query string, maxResults int, config map[string]string) (string, error)
}

// WebSearchConfigField 搜索引擎配置项定义
type WebSearchConfigField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"` // 是否密码类型，前端用 password 输入
	Placeholder string `json:"placeholder,omitempty"`
}

// webSearchRegistry 搜索引擎注册表
var webSearchRegistry = struct {
	mu        sync.RWMutex
	providers map[string]WebSearchProvider
}{
	providers: make(map[string]WebSearchProvider),
}

// RegisterWebSearchProvider 注册网页搜索引擎
func RegisterWebSearchProvider(p WebSearchProvider) {
	webSearchRegistry.mu.Lock()
	defer webSearchRegistry.mu.Unlock()
	webSearchRegistry.providers[p.ID()] = p
}

// GetWebSearchProvider 按 ID 获取搜索引擎
func GetWebSearchProvider(id string) WebSearchProvider {
	webSearchRegistry.mu.RLock()
	defer webSearchRegistry.mu.RUnlock()
	return webSearchRegistry.providers[id]
}

// ListWebSearchProviders 返回所有已注册的搜索引擎
func ListWebSearchProviders() []WebSearchProvider {
	webSearchRegistry.mu.RLock()
	defer webSearchRegistry.mu.RUnlock()
	out := make([]WebSearchProvider, 0, len(webSearchRegistry.providers))
	for _, p := range webSearchRegistry.providers {
		out = append(out, p)
	}
	return out
}
