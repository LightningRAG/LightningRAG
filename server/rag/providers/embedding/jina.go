package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// JinaEmbed Jina AI 嵌入实现，参考 references 目录内 JinaMultiVecEmbed
// API: https://api.jina.ai/v1/embeddings
// 支持 jina-embeddings-v2/v3/v4，v3/v4 需指定 task (retrieval.passage / retrieval.query)
type JinaEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	dimensions int
}

// NewJinaEmbed 创建 Jina Embedder
func NewJinaEmbed(apiKey, baseURL, model string, dimensions int) *JinaEmbed {
	if baseURL == "" {
		baseURL = "https://api.jina.ai/v1/embeddings"
	}
	if model == "" {
		model = "jina-embeddings-v3"
	}
	return &JinaEmbed{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		client:     &http.Client{},
		dimensions: dimensions,
	}
}

type jinaEmbedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Task       string   `json:"task,omitempty"`
	Truncate   bool     `json:"truncate,omitempty"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type jinaEmbedData struct {
	Index      int         `json:"index"`
	Embedding  []float32   `json:"embedding,omitempty"`
	Embeddings [][]float32 `json:"embeddings,omitempty"` // v4 multivector
}

type jinaEmbedResponse struct {
	Data []jinaEmbedData `json:"data"`
}

func (e *JinaEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	batchSize := 16
	result := make([][]float32, 0, len(texts))
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		embeds, err := e.embedBatch(ctx, batch, "retrieval.passage")
		if err != nil {
			return nil, err
		}
		result = append(result, embeds...)
	}
	return result, nil
}

func (e *JinaEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeds, err := e.embedBatch(ctx, []string{text}, "retrieval.query")
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeds[0], nil
}

func (e *JinaEmbed) embedBatch(ctx context.Context, texts []string, task string) ([][]float32, error) {
	reqBody := map[string]any{
		"model": e.model,
		"input": texts,
	}
	if strings.Contains(e.model, "v3") || strings.Contains(e.model, "v4") {
		reqBody["task"] = task
		reqBody["truncate"] = true
	}
	if e.dimensions > 0 {
		reqBody["dimensions"] = e.dimensions
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jina embedding error: %s", string(b))
	}

	var result jinaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(result.Data), len(texts))
	}
	batch := make([]openaiEmbedDatum, len(result.Data))
	for i, d := range result.Data {
		var emb []float32
		if len(d.Embedding) > 0 {
			emb = d.Embedding
		} else if len(d.Embeddings) > 0 {
			vecs := d.Embeddings
			dim := len(vecs[0])
			avg := make([]float32, dim)
			for _, v := range vecs {
				for j := range v {
					avg[j] += v[j]
				}
			}
			n := float32(len(vecs))
			for j := range avg {
				avg[j] /= n
			}
			emb = avg
		} else {
			return nil, fmt.Errorf("jina embedding: empty data at index %d", i)
		}
		batch[i] = openaiEmbedDatum{Embedding: emb, Index: d.Index}
	}
	return mapOpenAIEmbeddingBatch(len(texts), batch)
}

func (e *JinaEmbed) ProviderName() string { return "jina" }
func (e *JinaEmbed) ModelName() string    { return e.model }
func (e *JinaEmbed) Dimensions() int      { return e.dimensions }
