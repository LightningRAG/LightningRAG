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

// OpenAIEmbed OpenAI 嵌入模型实现
type OpenAIEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	dimensions int
}

// NewOpenAIEmbed 创建 OpenAI Embedder
func NewOpenAIEmbed(apiKey, baseURL, model string, dimensions int) *OpenAIEmbed {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIEmbed{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		client:     &http.Client{},
		dimensions: dimensions,
	}
}

type openaiEmbedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type openaiEmbedDatum struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type openaiEmbedResponse struct {
	Data []openaiEmbedDatum `json:"data"`
}

// mapOpenAIEmbeddingBatch 将 OpenAI 兼容的 data[] 对齐到与输入文本同序的向量。
// 规范要求用 index 对应 input 下标；部分网关会打乱 data 顺序或省略 index（均为 0），
// 按 index 无法填满时再按响应数组顺序回退，避免向量与错误文本绑定（检索结果重复/错乱）。
func mapOpenAIEmbeddingBatch(textsLen int, data []openaiEmbedDatum) ([][]float32, error) {
	if len(data) != textsLen {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(data), textsLen)
	}
	out := make([][]float32, textsLen)
	for _, d := range data {
		if d.Index >= 0 && d.Index < textsLen {
			out[d.Index] = d.Embedding
		}
	}
	filled := 0
	for _, e := range out {
		if len(e) > 0 {
			filled++
		}
	}
	if filled == textsLen {
		return out, nil
	}
	out = make([][]float32, textsLen)
	for i := range data {
		out[i] = data[i].Embedding
	}
	return out, nil
}

func (e *OpenAIEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	// OpenAI 限制 batch size 最多 2048，通常 16 较安全
	batchSize := 16
	result := make([][]float32, 0, len(texts))
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		embeds, err := e.embedBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
		result = append(result, embeds...)
	}
	return result, nil
}

func (e *OpenAIEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeds, err := e.embedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeds[0], nil
}

func (e *OpenAIEmbed) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := openaiEmbedRequest{
		Model: e.model,
		Input: texts,
	}
	if e.dimensions > 0 {
		reqBody.Dimensions = e.dimensions
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewReader(body))
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
		return nil, fmt.Errorf("openai embedding error: %s", string(b))
	}

	var result openaiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return mapOpenAIEmbeddingBatch(len(texts), result.Data)
}

func (e *OpenAIEmbed) ProviderName() string { return "openai" }
func (e *OpenAIEmbed) ModelName() string    { return e.model }
func (e *OpenAIEmbed) Dimensions() int      { return e.dimensions }
