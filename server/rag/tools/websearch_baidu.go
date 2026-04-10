package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baiduWebSearchURL = "https://qianfan.baidubce.com/v2/ai_search/web_search"

func init() {
	RegisterWebSearchProvider(&BaiduProvider{})
}

// BaiduProvider 百度千帆网页搜索，需配置 API Key
type BaiduProvider struct{}

func (b *BaiduProvider) ID() string {
	return "baidu"
}

func (b *BaiduProvider) DisplayName() string {
	return "Baidu"
}

func (b *BaiduProvider) ConfigSchema() []WebSearchConfigField {
	return []WebSearchConfigField{
		{Key: "apiKey", Label: "API Key", Required: true, Secret: true, Placeholder: "Qianfan API Key"},
	}
}

func (b *BaiduProvider) Search(ctx context.Context, query string, maxResults int, config map[string]string) (string, error) {
	apiKey := strings.TrimSpace(config["apiKey"])
	if apiKey == "" {
		return "", fmt.Errorf("Baidu search requires an API key; set it under model settings → web search")
	}
	reqBody := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": truncateQuery(query, 72)},
		},
		"search_source": "baidu_search_v2",
		"resource_type_filter": []map[string]any{
			{"type": "web", "top_k": maxResults},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baiduWebSearchURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Appbuilder-Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}
	var result struct {
		Code       int    `json:"code"`
		Message    string `json:"message"`
		References []struct {
			ID      int    `json:"id"`
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"snippet"`
			Content string `json:"content"`
		} `json:"references"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("parse response failed: %w", err)
	}
	if result.Code != 0 && result.Code != 200 {
		return "", fmt.Errorf("Baidu search API error: %s (code=%d)", result.Message, result.Code)
	}
	var sb strings.Builder
	for i, r := range result.References {
		desc := r.Snippet
		if desc == "" {
			desc = r.Content
		}
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n   %s\n\n", i+1, r.Title, r.URL, desc))
	}
	if sb.Len() == 0 {
		return "No relevant results found.", nil
	}
	return strings.TrimSpace(sb.String()) + "\n\n(Results powered by Baidu search.)", nil
}

// truncateQuery 百度 API 限制 content 长度 72 字符（汉字占 2 字符）
func truncateQuery(s string, maxBytes int) string {
	runes := []rune(s)
	n := 0
	for i, r := range runes {
		if r >= 0x4e00 && r <= 0x9fff {
			n += 2
		} else {
			n++
		}
		if n > maxBytes {
			return string(runes[:i])
		}
	}
	return s
}
