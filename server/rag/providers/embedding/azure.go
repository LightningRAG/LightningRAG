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

// AzureEmbed Azure OpenAI 文本嵌入（部署名在 URL 路径中）
type AzureEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	apiVersion string
	dimensions int
	client     *http.Client
}

// NewAzureEmbed 创建 Azure OpenAI Embedder；baseURL 为资源端点（如 https://xxx.openai.azure.com）
func NewAzureEmbed(apiKey, baseURL, model, apiVersion string, dimensions int) *AzureEmbed {
	if baseURL == "" {
		baseURL = "https://YOUR_RESOURCE.openai.azure.com"
	}
	if apiVersion == "" {
		apiVersion = "2024-02-01"
	}
	return &AzureEmbed{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		apiVersion: apiVersion,
		dimensions: dimensions,
		client:     &http.Client{},
	}
}

func (e *AzureEmbed) embeddingsURL() string {
	return fmt.Sprintf("%s/openai/deployments/%s/embeddings?api-version=%s",
		e.baseURL, e.model, e.apiVersion)
}

func (e *AzureEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
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
		embeds, err := e.embedBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
		result = append(result, embeds...)
	}
	return result, nil
}

func (e *AzureEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeds, err := e.embedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeds[0], nil
}

func (e *AzureEmbed) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := openaiEmbedRequest{
		Model: e.model,
		Input: texts,
	}
	if e.dimensions > 0 {
		reqBody.Dimensions = e.dimensions
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", e.embeddingsURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure embedding error: %s", string(b))
	}

	var result openaiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return mapOpenAIEmbeddingBatch(len(texts), result.Data)
}

func (e *AzureEmbed) ProviderName() string { return "azure" }
func (e *AzureEmbed) ModelName() string    { return e.model }
func (e *AzureEmbed) Dimensions() int      { return e.dimensions }
