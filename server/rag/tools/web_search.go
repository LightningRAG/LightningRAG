package tools

import (
	"context"
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
)

func init() {
	Register(&WebSearchTool{})
}

// WebSearchTool 网页搜索引擎工具，根据用户配置调用 DuckDuckGo 或百度
type WebSearchTool struct{}

func (w *WebSearchTool) Name() string {
	return "web_search"
}

func (w *WebSearchTool) Description() string {
	return "Search the web for information. Use when the user asks for real-time facts, news, recent events, or topics unlikely to be in the knowledge base."
}

func (w *WebSearchTool) Parameters() *ParameterSchema {
	return &ParameterSchema{
		Type: "object",
		Properties: map[string]PropertySchema{
			"query": {
				Type:        "string",
				Description: "Search keywords or question",
			},
			"max_results": {
				Type:        "number",
				Description: "Number of results to return; default 5, max 20",
			},
		},
		Required: []string{"query"},
	}
}

func (w *WebSearchTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	query, _ := params["query"].(string)
	if query == "" {
		if q, ok := params["q"].(string); ok {
			query = q
		}
	}
	if query == "" {
		return "", fmt.Errorf("missing required parameter query")
	}

	maxResults := 5
	if n, ok := params["max_results"].(float64); ok && n > 0 && n <= 20 {
		maxResults = int(n)
	}
	if maxResults > 20 {
		maxResults = 20
	}

	uid := UserIDFromContext(ctx)
	provider, config := getWebSearchConfig(ctx, uid)
	if provider == "" {
		provider = "duckduckgo"
	}

	p := GetWebSearchProvider(provider)
	if p == nil {
		return "", fmt.Errorf("unsupported search engine %q; configure DuckDuckGo or Baidu under model settings → web search", provider)
	}

	return p.Search(ctx, query, maxResults, config)
}

// getWebSearchConfig 获取用户互联网搜索配置
// 回退链：用户自定义（UseSystemDefault=false） → 系统全局默认 → DuckDuckGo
func getWebSearchConfig(ctx context.Context, uid uint) (provider string, config map[string]string) {
	if uid == 0 {
		return resolveSystemWebSearchDefault(ctx)
	}
	var cfg rag.RagUserWebSearchConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ?", uid).First(&cfg).Error; err != nil {
		return resolveSystemWebSearchDefault(ctx)
	}
	if cfg.UseSystemDefault {
		return resolveSystemWebSearchDefault(ctx)
	}
	if cfg.Provider == "" {
		return "duckduckgo", nil
	}
	configMap := make(map[string]string)
	if cfg.Config != nil {
		for k, v := range cfg.Config {
			if s, ok := v.(string); ok {
				configMap[k] = s
			}
		}
	}
	return cfg.Provider, configMap
}

// resolveSystemWebSearchDefault 从系统默认互联网搜索配置获取，无配置则返回 DuckDuckGo
func resolveSystemWebSearchDefault(ctx context.Context) (string, map[string]string) {
	var cfg rag.RagSystemDefaultWebSearchConfig
	if err := global.LRAG_DB.WithContext(ctx).First(&cfg).Error; err != nil || cfg.Provider == "" {
		return "duckduckgo", nil
	}
	configMap := make(map[string]string)
	if cfg.Config != nil {
		for k, v := range cfg.Config {
			if s, ok := v.(string); ok {
				configMap[k] = s
			}
		}
	}
	return cfg.Provider, configMap
}
