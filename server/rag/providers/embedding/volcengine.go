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

// VolcEngineEmbed 火山引擎嵌入实现，参考 references 目录内 VolcEngineEmbed
// API: https://ark.cn-beijing.volces.com/api/v3/embeddings/multimodal
// key 可为 JSON: {"ark_api_key":"xxx"}
type VolcEngineEmbed struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	dimensions int
}

// NewVolcEngineEmbed 创建火山引擎 Embedder
func NewVolcEngineEmbed(apiKey, baseURL, model string, dimensions int) *VolcEngineEmbed {
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if model == "" {
		model = "ep-20241104103920-xxxxx" // 需用户配置实际 endpoint
	}
	return &VolcEngineEmbed{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		client:     &http.Client{},
		dimensions: dimensions,
	}
}

type volcInputItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type volcEmbedRequest struct {
	Model string          `json:"model"`
	Input []volcInputItem `json:"input"`
}

type volcEmbedData struct {
	Embedding []float32 `json:"embedding"`
}

type volcEmbedResponse struct {
	Data json.RawMessage `json:"data"` // 可能是 [] 或 {}
}

func (e *VolcEngineEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
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

func (e *VolcEngineEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.embedOne(ctx, text)
}

func (e *VolcEngineEmbed) embedOne(ctx context.Context, text string) ([]float32, error) {
	input := []volcInputItem{{Type: "text", Text: text}}
	reqBody := volcEmbedRequest{Model: e.model, Input: input}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings/multimodal", bytes.NewReader(body))
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
		return nil, fmt.Errorf("volcengine embedding error: %s", string(b))
	}

	var result volcEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var dataList []volcEmbedData
	if err := json.Unmarshal(result.Data, &dataList); err != nil {
		var single volcEmbedData
		if err2 := json.Unmarshal(result.Data, &single); err2 != nil {
			return nil, fmt.Errorf("volcengine: invalid data format")
		}
		return single.Embedding, nil
	}
	if len(dataList) == 0 {
		return nil, fmt.Errorf("volcengine: empty embedding")
	}
	return dataList[0].Embedding, nil
}

func (e *VolcEngineEmbed) ProviderName() string { return "volcengine" }
func (e *VolcEngineEmbed) ModelName() string    { return e.model }
func (e *VolcEngineEmbed) Dimensions() int      { return e.dimensions }
