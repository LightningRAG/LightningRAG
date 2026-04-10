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

// VoyageEmbed Voyage AI 嵌入实现，参考 references 目录内 VoyageEmbed
// API: https://api.voyageai.com/v1/embeddings
type VoyageEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	dimensions int
}

// NewVoyageEmbed 创建 Voyage AI Embedder
func NewVoyageEmbed(apiKey, baseURL, model string, dimensions int) *VoyageEmbed {
	if baseURL == "" {
		baseURL = "https://api.voyageai.com/v1"
	}
	if model == "" {
		model = "voyage-3"
	}
	return &VoyageEmbed{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		client:     &http.Client{},
		dimensions: dimensions,
	}
}

type voyageEmbedRequest struct {
	Input           []string `json:"input"`
	Model           string   `json:"model"`
	InputType       string   `json:"input_type,omitempty"`
	OutputDimension int      `json:"output_dimension,omitempty"`
}

type voyageEmbedData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type voyageEmbedResponse struct {
	Data []voyageEmbedData `json:"data"`
}

func (e *VoyageEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
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
		embeds, err := e.embedBatch(ctx, batch, "document")
		if err != nil {
			return nil, err
		}
		result = append(result, embeds...)
	}
	return result, nil
}

func (e *VoyageEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeds, err := e.embedBatch(ctx, []string{text}, "query")
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeds[0], nil
}

func (e *VoyageEmbed) embedBatch(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	reqBody := voyageEmbedRequest{
		Input:     texts,
		Model:     e.model,
		InputType: inputType,
	}
	if e.dimensions > 0 {
		reqBody.OutputDimension = e.dimensions
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
		return nil, fmt.Errorf("voyage embedding error: %s", string(b))
	}

	var result voyageEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(result.Data), len(texts))
	}
	embeds := make([][]float32, len(texts))
	for _, d := range result.Data {
		if d.Index >= 0 && d.Index < len(texts) {
			embeds[d.Index] = d.Embedding
		}
	}
	filled := 0
	for _, e := range embeds {
		if len(e) > 0 {
			filled++
		}
	}
	if filled != len(texts) {
		embeds = make([][]float32, len(texts))
		for i, d := range result.Data {
			embeds[i] = d.Embedding
		}
	}
	return embeds, nil
}

func (e *VoyageEmbed) ProviderName() string { return "voyageai" }
func (e *VoyageEmbed) ModelName() string    { return e.model }
func (e *VoyageEmbed) Dimensions() int      { return e.dimensions }
