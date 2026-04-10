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

const maxDocLen = 8196

// Jina Jina Rerank 实现，参考 references 目录内 JinaRerank
type Jina struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewJina 创建 Jina Reranker
func NewJina(apiKey, baseURL, model string) *Jina {
	if baseURL == "" {
		baseURL = "https://api.jina.ai/v1/rerank"
	}
	return &Jina{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type jinaRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type jinaResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type jinaResponse struct {
	Results []jinaResult `json:"results"`
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (j *Jina) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, t := range texts {
		truncated[i] = truncate(t, maxDocLen)
	}
	reqBody := jinaRequest{
		Model:     j.model,
		Query:     query,
		Documents: truncated,
		TopN:      len(texts),
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", j.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.apiKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jina rerank error: %s", string(b))
	}

	var result jinaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rows := make([]scoreRow, len(result.Results))
	for i, r := range result.Results {
		rows[i] = scoreRow{Index: r.Index, Score: r.RelevanceScore}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (j *Jina) ProviderName() string { return "jina" }
func (j *Jina) ModelName() string    { return j.model }
