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

// Tongyi 通义 Rerank 实现，参考 references 目录内 QWenRerank
// API: https://dashscope.aliyuncs.com/api/v1/services/rerank/text-rerank/text-rerank
type Tongyi struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewTongyi 创建通义 Reranker
func NewTongyi(apiKey, baseURL, model string) *Tongyi {
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/api/v1/services/rerank/text-rerank"
	}
	if model == "" {
		model = "gte-rerank"
	}
	return &Tongyi{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type tongyiRerankInput struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
}

type tongyiRerankParams struct {
	TopN            int  `json:"top_n"`
	ReturnDocuments bool `json:"return_documents"`
}

type tongyiRerankRequest struct {
	Model      string             `json:"model"`
	Input      tongyiRerankInput  `json:"input"`
	Parameters tongyiRerankParams `json:"parameters"`
}

type tongyiRerankItem struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type tongyiRerankResponse struct {
	Status int `json:"status"`
	Data   struct {
		Rerank []tongyiRerankItem `json:"rerank"`
	} `json:"data"`
}

func truncateTongyi(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (t *Tongyi) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, s := range texts {
		truncated[i] = truncateTongyi(s, 4000)
	}
	reqBody := tongyiRerankRequest{
		Model: t.model,
		Input: tongyiRerankInput{
			Query:     query,
			Documents: truncated,
		},
		Parameters: tongyiRerankParams{
			TopN:            len(texts),
			ReturnDocuments: false,
		},
	}
	body, _ := json.Marshal(reqBody)

	url := t.baseURL
	if !strings.HasSuffix(url, "/text-rerank") {
		url = url + "/text-rerank"
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tongyi rerank error: %s", string(b))
	}

	var result tongyiRerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != 0 {
		return nil, fmt.Errorf("tongyi rerank status: %d", result.Status)
	}

	rows := make([]scoreRow, len(result.Data.Rerank))
	for i, r := range result.Data.Rerank {
		rows[i] = scoreRow{Index: r.Index, Score: r.RelevanceScore}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (t *Tongyi) ProviderName() string { return "tongyi" }
func (t *Tongyi) ModelName() string    { return t.model }
