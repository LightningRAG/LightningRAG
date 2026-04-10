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

const defaultDescribePrompt = "Describe the image in detail: time, place, people, events, and any visible data or text. Extract structured facts when present. Answer in the same language as the dominant text in the image, or in English if there is no text."

// OpenAI Vision 计算机视觉实现
type OpenAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAI 创建 OpenAI CV
func NewOpenAI(apiKey, baseURL, model string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAI{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type visionContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type visionMessage struct {
	Role    string              `json:"role"`
	Content []visionContentPart `json:"content"`
}

type visionRequest struct {
	Model    string          `json:"model"`
	Messages []visionMessage `json:"messages"`
}

type visionChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type visionResponse struct {
	Choices []visionChoice `json:"choices"`
}

func (c *OpenAI) imageToDataURL(data []byte) string {
	mime := "image/png"
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		mime = "image/jpeg"
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, b64)
}

func (c *OpenAI) describe(ctx context.Context, image []byte, prompt string) (string, error) {
	dataURL := c.imageToDataURL(image)
	reqBody := visionRequest{
		Model: c.model,
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

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vision api error: %s", string(b))
	}

	var result visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty vision response")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func (c *OpenAI) Describe(ctx context.Context, image []byte) (string, error) {
	return c.describe(ctx, image, defaultDescribePrompt)
}

func (c *OpenAI) DescribeWithPrompt(ctx context.Context, image []byte, prompt string) (string, error) {
	if prompt == "" {
		prompt = defaultDescribePrompt
	}
	return c.describe(ctx, image, prompt)
}

func (c *OpenAI) ProviderName() string { return "openai" }
func (c *OpenAI) ModelName() string    { return c.model }

// ProviderNameAdapter 包装 CV 以自定义 ProviderName
type ProviderNameAdapter struct {
	interfaces.CV
	DisplayName string
}

func (p *ProviderNameAdapter) ProviderName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.CV.ProviderName()
}
