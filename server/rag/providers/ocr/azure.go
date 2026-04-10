package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// Azure Azure OpenAI 视觉模型做 OCR（与 CV/Azure 相同：部署名在 URL，api-key 头）
type Azure struct {
	apiKey     string
	baseURL    string
	deployment string
	apiVersion string
	client     *http.Client
}

// NewAzure baseURL 为资源端点；deployment 为部署名
func NewAzure(apiKey, baseURL, deployment, apiVersion string) *Azure {
	if baseURL == "" {
		baseURL = "https://YOUR_RESOURCE.openai.azure.com"
	}
	if apiVersion == "" {
		apiVersion = "2024-02-01"
	}
	if deployment == "" {
		deployment = "gpt-4o"
	}
	return &Azure{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		deployment: deployment,
		apiVersion: apiVersion,
		client:     &http.Client{},
	}
}

func (a *Azure) chatURL() string {
	return fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		a.baseURL, a.deployment, a.apiVersion)
}

func azureOCRImageDataURL(data []byte) string {
	mime := "image/png"
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		mime = "image/jpeg"
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, b64)
}

func (a *Azure) ExtractText(ctx context.Context, data []byte, filename string) (*interfaces.OCRResult, error) {
	_ = filename
	dataURL := azureOCRImageDataURL(data)
	reqBody := ocrRequest{
		Model: a.deployment,
		Messages: []ocrMessage{
			{
				Role: "user",
				Content: []ocrContentPart{
					{Type: "text", Text: ocrPrompt},
					{
						Type: "image_url",
						ImageURL: &struct {
							URL string `json:"url"`
						}{URL: dataURL},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.chatURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.apiKey)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure ocr: %s", string(b))
	}
	var result ocrResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("azure ocr: 空响应")
	}
	text := strings.TrimSpace(result.Choices[0].Message.Content)
	return &interfaces.OCRResult{Text: text, Sections: []string{text}}, nil
}

func (a *Azure) ProviderName() string { return "azure" }
func (a *Azure) ModelName() string    { return a.deployment }
