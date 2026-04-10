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

// OllamaEmbed Ollama 嵌入模型实现，支持本地部署
type OllamaEmbed struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbed 创建 Ollama Embedder
func NewOllamaEmbed(baseURL, model string) *OllamaEmbed {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaEmbed{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (e *OllamaEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	result := make([][]float32, 0, len(texts))
	for _, t := range texts {
		emb, err := e.embedOne(ctx, t)
		if err != nil {
			return nil, err
		}
		result = append(result, emb)
	}
	return result, nil
}

func (e *OllamaEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.embedOne(ctx, text)
}

func (e *OllamaEmbed) embedOne(ctx context.Context, text string) ([]float32, error) {
	reqBody := ollamaEmbedRequest{
		Model:  e.model,
		Prompt: text,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embedding error: %s", string(b))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}

func (e *OllamaEmbed) ProviderName() string { return "ollama" }
func (e *OllamaEmbed) ModelName() string    { return e.model }
func (e *OllamaEmbed) Dimensions() int      { return 0 }
