package rerank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// HuggingFaceHTTP 调用本地 Text Embeddings Inference（TEI）等服务的 /rerank 端点，与上游参考 HuggingfaceRerank 行为一致
type HuggingFaceHTTP struct {
	rerankURL string
	model     string
	client    *http.Client
}

// NewHuggingFaceHTTP baseURL 可为完整 URL（如 http://127.0.0.1:8080）或 host:port；空则默认 127.0.0.1:80
// modelName 仅用于展示（对应模型名可含 ___ 后缀，此处取第一段）
func NewHuggingFaceHTTP(baseURL, modelName string) *HuggingFaceHTTP {
	base := normalizeHFRerankBase(baseURL)
	model := modelName
	if i := strings.Index(model, "___"); i >= 0 {
		model = model[:i]
	}
	if strings.TrimSpace(model) == "" {
		model = "BAAI/bge-reranker-v2-m3"
	}
	return &HuggingFaceHTTP{
		rerankURL: strings.TrimSuffix(base, "/") + "/rerank",
		model:     strings.TrimSpace(model),
		client:    &http.Client{},
	}
}

func normalizeHFRerankBase(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return "http://127.0.0.1:80"
	}
	if strings.Contains(u, "://") {
		return strings.TrimSuffix(u, "/")
	}
	return "http://" + strings.TrimPrefix(strings.TrimPrefix(u, "http://"), "https://")
}

type hfRerankReq struct {
	Query     string   `json:"query"`
	Texts     []string `json:"texts"`
	RawScores bool     `json:"raw_scores"`
	Truncate  bool     `json:"truncate"`
}

type hfRerankItem struct {
	Index int     `json:"index"`
	Score float64 `json:"score"`
}

func (h *HuggingFaceHTTP) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	scores := make([]float32, len(texts))
	const batchSize = 8
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		body, _ := json.Marshal(hfRerankReq{
			Query:     query,
			Texts:     batch,
			RawScores: false,
			Truncate:  true,
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.rerankURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := h.client.Do(req)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("huggingface rerank: HTTP %d %s", resp.StatusCode, string(b))
		}
		var items []hfRerankItem
		if err := json.Unmarshal(b, &items); err != nil {
			return nil, fmt.Errorf("huggingface rerank: %w", err)
		}
		for _, o := range items {
			idx := o.Index + i
			if idx >= 0 && idx < len(scores) {
				scores[idx] = float32(o.Score)
			}
		}
	}
	return scores, nil
}

func (h *HuggingFaceHTTP) ProviderName() string { return "huggingface" }
func (h *HuggingFaceHTTP) ModelName() string    { return h.model }

var _ interfaces.Reranker = (*HuggingFaceHTTP)(nil)
