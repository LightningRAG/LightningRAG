package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kuhahalong/ddgsearch"
)

func init() {
	RegisterWebSearchProvider(&DuckDuckGoProvider{})
}

// DuckDuckGoProvider DuckDuckGo 搜索引擎，无需 API Key
type DuckDuckGoProvider struct{}

func (d *DuckDuckGoProvider) ID() string {
	return "duckduckgo"
}

func (d *DuckDuckGoProvider) DisplayName() string {
	return "DuckDuckGo"
}

func (d *DuckDuckGoProvider) ConfigSchema() []WebSearchConfigField {
	return nil // 无需配置
}

func (d *DuckDuckGoProvider) Search(ctx context.Context, query string, maxResults int, config map[string]string) (string, error) {
	cfg := &ddgsearch.Config{
		Timeout:    15 * time.Second,
		MaxRetries: 2,
		Cache:      false,
	}
	client, err := ddgsearch.New(cfg)
	if err != nil {
		// 初始化失败直接作为工具错误抛出
		return "", fmt.Errorf("init DuckDuckGo search client failed: %w", err)
	}
	searchParams := &ddgsearch.SearchParams{
		Query:      query,
		Region:     ddgsearch.RegionUS,
		SafeSearch: ddgsearch.SafeSearchModerate,
		MaxResults: maxResults,
	}
	resp, err := client.Search(ctx, searchParams)
	if err != nil {
		// 对常见网络问题做更友好的说明，并将原因作为正常结果返回给上层，
		// 便于大模型直接转述给用户，而不是只看到“工具执行失败”。
		errMsg := err.Error()
		// 网络超时或无法连接（本地/服务器网络无法访问 DuckDuckGo，常见于被防火墙屏蔽）
		if errors.Is(err, context.DeadlineExceeded) ||
			strings.Contains(strings.ToLower(errMsg), "timeout") ||
			strings.Contains(strings.ToLower(errMsg), "context deadline exceeded") {
			return fmt.Sprintf(
				"DuckDuckGo search timed out; this server may be unable to reach DuckDuckGo (firewall or proxy).\n"+
					"Suggestions:\n"+
					"1. Verify https://duckduckgo.com is reachable from the server.\n"+
					"2. Or switch web search to Baidu in model settings → web search and set its API key.\n\n"+
					"Underlying error: %s", errMsg,
			), nil
		}

		// 其他错误也以可读形式返回，让用户能看到具体原因
		return fmt.Sprintf(
			"DuckDuckGo search failed; web search could not be completed.\n"+
				"Retry later, or switch to another engine (e.g. Baidu) under model settings → web search.\n\n"+
				"Underlying error: %s", errMsg,
		), nil
	}
	var sb strings.Builder
	results := resp.Results
	if results == nil {
		results = []ddgsearch.SearchResult{}
	}
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n   %s\n\n", i+1, r.Title, r.URL, r.Description))
	}
	if sb.Len() == 0 {
		return "No relevant results found.", nil
	}
	return strings.TrimSpace(sb.String()), nil
}
