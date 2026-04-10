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

// CohereEmbed Cohere 嵌入实现，参考 references 目录内 CoHereEmbed
// API: https://api.cohere.ai/v1/embed
type CohereEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	dimensions int
}

// NewCohereEmbed 创建 Cohere Embedder
func NewCohereEmbed(apiKey, baseURL, model string, dimensions int) *CohereEmbed {
	if baseURL == "" {
		baseURL = "https://api.cohere.ai"
	}
	if model == "" {
		model = "embed-english-v3.0"
	}
	return &CohereEmbed{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		client:     &http.Client{},
		dimensions: dimensions,
	}
}

type cohereEmbedRequest struct {
	Texts          []string `json:"texts"`
	Model          string   `json:"model"`
	InputType      string   `json:"input_type"`
	EmbeddingTypes []string `json:"embedding_types,omitempty"`
}

func (e *CohereEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
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
		embeds, err := e.embedBatch(ctx, batch, "search_document")
		if err != nil {
			return nil, err
		}
		result = append(result, embeds...)
	}
	return result, nil
}

func (e *CohereEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeds, err := e.embedBatch(ctx, []string{text}, "search_query")
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeds[0], nil
}

func (e *CohereEmbed) embedBatch(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	reqBody := cohereEmbedRequest{
		Texts:          texts,
		Model:          e.model,
		InputType:      inputType,
		EmbeddingTypes: []string{"float"},
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/v1/embed", bytes.NewReader(body))
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
		return nil, fmt.Errorf("cohere embedding error: %s", string(b))
	}

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("cohere embedding: empty response")
	}
	return result.Embeddings, nil
}

func (e *CohereEmbed) ProviderName() string { return "cohere" }
func (e *CohereEmbed) ModelName() string    { return e.model }
func (e *CohereEmbed) Dimensions() int      { return e.dimensions }
