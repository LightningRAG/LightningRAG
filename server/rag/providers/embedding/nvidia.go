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

// NvidiaEmbed NVIDIA NIM 嵌入实现，参考 references 目录内 NvidiaEmbed
// API: https://integrate.api.nvidia.com/v1/embeddings 或 ai.api.nvidia.com
type NvidiaEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	dimensions int
}

// NewNvidiaEmbed 创建 NVIDIA Embedder
func NewNvidiaEmbed(apiKey, baseURL, model string, dimensions int) *NvidiaEmbed {
	if baseURL == "" {
		baseURL = "https://integrate.api.nvidia.com/v1/embeddings"
	}
	if model == "" {
		model = "nvidia/nv-embed-qa-4"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	if strings.Contains(model, "embed-qa-4") || model == "nvidia/embed-qa-4" {
		baseURL = "https://ai.api.nvidia.com/v1/retrieval/nvidia/embeddings"
		model = "NV-Embed-QA"
	} else if strings.Contains(model, "arctic-embed") {
		baseURL = "https://ai.api.nvidia.com/v1/retrieval/snowflake/arctic-embed-l/embeddings"
	}
	return &NvidiaEmbed{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		client:     &http.Client{},
		dimensions: dimensions,
	}
}

type nvidiaEmbedRequest struct {
	Input          []string `json:"input"`
	InputType      string   `json:"input_type"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format"`
	Truncate       string   `json:"truncate"`
}

type nvidiaEmbedData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type nvidiaEmbedResponse struct {
	Data []nvidiaEmbedData `json:"data"`
}

func (e *NvidiaEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
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
		embeds, err := e.embedBatch(ctx, batch, "passage")
		if err != nil {
			return nil, err
		}
		result = append(result, embeds...)
	}
	return result, nil
}

func (e *NvidiaEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeds, err := e.embedBatch(ctx, []string{text}, "query")
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeds[0], nil
}

func (e *NvidiaEmbed) embedBatch(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	reqBody := nvidiaEmbedRequest{
		Input:          texts,
		InputType:      inputType,
		Model:          e.model,
		EncodingFormat: "float",
		Truncate:       "END",
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
		return nil, fmt.Errorf("nvidia embedding error: %s", string(b))
	}

	var result nvidiaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	batch := make([]openaiEmbedDatum, len(result.Data))
	for i, d := range result.Data {
		batch[i] = openaiEmbedDatum{Embedding: d.Embedding, Index: d.Index}
	}
	return mapOpenAIEmbeddingBatch(len(texts), batch)
}

func (e *NvidiaEmbed) ProviderName() string { return "nvidia" }
func (e *NvidiaEmbed) ModelName() string    { return e.model }
func (e *NvidiaEmbed) Dimensions() int      { return e.dimensions }
