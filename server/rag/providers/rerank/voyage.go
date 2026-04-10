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

const maxDocLenVoyage = 8196

// Voyage Voyage AI Rerank 实现，参考 references 目录内 VoyageRerank
// API: https://api.voyageai.com/v1/rerank
type Voyage struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewVoyage 创建 Voyage AI Reranker
func NewVoyage(apiKey, baseURL, model string) *Voyage {
	if baseURL == "" {
		baseURL = "https://api.voyageai.com/v1"
	}
	if model == "" {
		model = "rerank-2"
	}
	return &Voyage{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type voyageRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	Model     string   `json:"model"`
	TopK      int      `json:"top_k,omitempty"`
}

type voyageDataItem struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type voyageResponse struct {
	Data []voyageDataItem `json:"data"`
}

func truncateVoyage(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (v *Voyage) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, t := range texts {
		truncated[i] = truncateVoyage(t, maxDocLenVoyage)
	}
	reqBody := voyageRequest{
		Query:     query,
		Documents: truncated,
		Model:     v.model,
		TopK:      len(texts),
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/rerank", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("voyage rerank error: %s", string(b))
	}

	var result voyageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rows := make([]scoreRow, len(result.Data))
	for i, r := range result.Data {
		rows[i] = scoreRow{Index: r.Index, Score: r.RelevanceScore}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (v *Voyage) ProviderName() string { return "voyageai" }
func (v *Voyage) ModelName() string    { return v.model }
