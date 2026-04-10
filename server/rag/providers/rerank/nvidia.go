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

const maxDocLenNvidia = 8196

// Nvidia NVIDIA NIM Rerank 实现，参考 references 目录内 NvidiaRerank
// API: https://ai.api.nvidia.com/v1/retrieval/nvidia/.../reranking
type Nvidia struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewNvidia 创建 NVIDIA Reranker
// 路径约定：nv-rerankqa-mistral-4b-v3 使用 /nv-rerankqa-mistral-4b-v3/reranking
func NewNvidia(apiKey, baseURL, model string) *Nvidia {
	if baseURL == "" {
		baseURL = "https://ai.api.nvidia.com/v1/retrieval/nvidia"
	}
	if model == "" {
		model = "nv-rerankqa-mistral-4b-v3"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	if strings.Contains(model, "nv-rerankqa-mistral-4b-v3") {
		baseURL = baseURL + "/nv-rerankqa-mistral-4b-v3/reranking"
	} else {
		baseURL = baseURL + "/reranking"
		if model == "rerank-qa-mistral-4b" || model == "nvidia/rerank-qa-mistral-4b" {
			model = "nv-rerankqa-mistral-4b:1"
		}
	}
	return &Nvidia{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

type nvidiaRequest struct {
	Model    string          `json:"model"`
	Query    nvidiaQuery     `json:"query"`
	Passages []nvidiaPassage `json:"passages"`
	Truncate string          `json:"truncate"`
	TopN     int             `json:"top_n"`
}

type nvidiaQuery struct {
	Text string `json:"text"`
}

type nvidiaPassage struct {
	Text string `json:"text"`
}

type nvidiaRanking struct {
	Index int     `json:"index"`
	Logit float64 `json:"logit"`
}

type nvidiaResponse struct {
	Rankings []nvidiaRanking `json:"rankings"`
}

func truncateNvidia(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (n *Nvidia) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, t := range texts {
		truncated[i] = truncateNvidia(t, maxDocLenNvidia)
	}
	passages := make([]nvidiaPassage, len(truncated))
	for i, t := range truncated {
		passages[i] = nvidiaPassage{Text: t}
	}
	reqBody := nvidiaRequest{
		Model:    n.model,
		Query:    nvidiaQuery{Text: query},
		Passages: passages,
		Truncate: "END",
		TopN:     len(texts),
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", n.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nvidia rerank error: %s", string(b))
	}

	var result nvidiaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rows := make([]scoreRow, len(result.Rankings))
	for i, r := range result.Rankings {
		rows[i] = scoreRow{Index: r.Index, Score: r.Logit}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (n *Nvidia) ProviderName() string { return "nvidia" }
func (n *Nvidia) ModelName() string    { return n.model }
