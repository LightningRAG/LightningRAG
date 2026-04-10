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

// Cohere Rerank 实现，参考 references 目录内 CoHereRerank
type Cohere struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewCohere 创建 Cohere Reranker
func NewCohere(apiKey, baseURL, model string) *Cohere {
	if baseURL == "" {
		baseURL = "https://api.cohere.ai"
	}
	if model == "" {
		model = "rerank-v3.5"
	}
	return &Cohere{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type cohereRerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type cohereRerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type cohereRerankResponse struct {
	Results []cohereRerankResult `json:"results"`
}

func truncateCohere(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (c *Cohere) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, t := range texts {
		truncated[i] = truncateCohere(t, 8196)
	}
	reqBody := cohereRerankRequest{
		Model:     c.model,
		Query:     query,
		Documents: truncated,
		TopN:      len(texts),
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v2/rerank", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cohere rerank error: %s", string(b))
	}

	var result cohereRerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rows := make([]scoreRow, len(result.Results))
	for i, r := range result.Results {
		rows[i] = scoreRow{Index: r.Index, Score: r.RelevanceScore}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (c *Cohere) ProviderName() string { return "cohere" }
func (c *Cohere) ModelName() string    { return c.model }
