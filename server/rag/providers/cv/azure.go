package cv

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

// Azure Azure OpenAI 多模态 chat/completions（与上游参考 AzureGptV4 一致：部署名在 URL，api-key 头）
type Azure struct {
	apiKey     string
	baseURL    string
	deployment string
	apiVersion string
	client     *http.Client
}

// NewAzure baseURL 为资源端点；deployment 为部署名（与模型名一致）
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

// ParseAzureAPIKey 若 key 为 JSON（含 api_key 字段）则解析出真实密钥
func ParseAzureAPIKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" || key[0] != '{' {
		return key
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(key), &m); err != nil {
		return key
	}
	if v, ok := m["api_key"].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return key
}

func azureImageDataURL(data []byte) string {
	mime := "image/png"
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		mime = "image/jpeg"
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, b64)
}

func (a *Azure) describe(ctx context.Context, image []byte, prompt string) (string, error) {
	dataURL := azureImageDataURL(image)
	reqBody := visionRequest{
		Model: a.deployment,
		Messages: []visionMessage{
			{
				Role: "user",
				Content: []visionContentPart{
					{Type: "text", Text: prompt},
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
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.apiKey)
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("azure vision: %s", string(b))
	}
	var result visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("azure vision: 空响应")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func (a *Azure) Describe(ctx context.Context, image []byte) (string, error) {
	return a.describe(ctx, image, defaultDescribePrompt)
}

func (a *Azure) DescribeWithPrompt(ctx context.Context, image []byte, prompt string) (string, error) {
	if prompt == "" {
		prompt = defaultDescribePrompt
	}
	return a.describe(ctx, image, prompt)
}

func (a *Azure) ProviderName() string { return "azure" }
func (a *Azure) ModelName() string    { return a.deployment }

var _ interfaces.CV = (*Azure)(nil)
