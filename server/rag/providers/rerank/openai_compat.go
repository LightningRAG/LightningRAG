package rerank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxDocLenCompat = 500

// OpenAICompat OpenAI 兼容 API 的 Rerank 实现（Jina 格式）
// 支持 LocalAI、Xinference、VLLM 等提供 /rerank 端点的服务
type OpenAICompat struct {
	apiKey   string
	baseURL  string
	model    string
	client   *http.Client
	provider string
}

// NewOpenAICompat 创建 OpenAI 兼容 Reranker
func NewOpenAICompat(provider, apiKey, baseURL, model string) *OpenAICompat {
	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/rerank") {
		baseURL = baseURL + "/rerank"
	}
	return &OpenAICompat{
		apiKey:   apiKey,
		baseURL:  baseURL,
		model:    model,
		client:   &http.Client{},
		provider: provider,
	}
}

type openaiCompatRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type openaiCompatResult struct {
	Index          int     `json:"index"`
	DocumentIndex  int     `json:"document_index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type openaiCompatResponse struct {
	Results []openaiCompatResult `json:"results"`
}

func truncateCompat(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (o *OpenAICompat) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, t := range texts {
		truncated[i] = truncateCompat(t, maxDocLenCompat)
	}
	reqBody := openaiCompatRequest{
		Model:     o.model,
		Query:     query,
		Documents: truncated,
		TopN:      len(texts),
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rerank api error: %s", string(b))
	}

	var result openaiCompatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rows := make([]scoreRow, len(result.Results))
	for i, r := range result.Results {
		rows[i] = scoreRow{Index: r.Index, DocIdx: r.DocumentIndex, Score: r.RelevanceScore}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (o *OpenAICompat) ProviderName() string {
	if o.provider != "" {
		return o.provider
	}
	return "openai_compat"
}
func (o *OpenAICompat) ModelName() string { return o.model }
