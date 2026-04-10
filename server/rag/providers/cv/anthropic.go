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

// Anthropic Claude 视觉（Messages API 多模态 content，与上游参考 AnthropicCV 行为一致）
type Anthropic struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
	client    *http.Client
}

// NewAnthropic baseURL 为空时使用 https://api.anthropic.com/v1
func NewAnthropic(apiKey, baseURL, model string) *Anthropic {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	maxTok := 8192
	ml := strings.ToLower(model)
	if strings.Contains(ml, "haiku") || strings.Contains(ml, "opus") {
		maxTok = 4096
	}
	return &Anthropic{
		apiKey:    apiKey,
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		model:     model,
		maxTokens: maxTok,
		client:    &http.Client{},
	}
}

type anthropicVisionMsg struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicVisionRequest struct {
	Model     string               `json:"model"`
	MaxTokens int                  `json:"max_tokens"`
	Messages  []anthropicVisionMsg `json:"messages"`
}

type anthropicVisionImageBlock struct {
	Type   string `json:"type"`
	Source struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source"`
}

type anthropicVisionTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func anthropicImageMediaType(image []byte) string {
	if len(image) >= 2 && image[0] == 0xFF && image[1] == 0xD8 {
		return "image/jpeg"
	}
	return "image/png"
}

func (a *Anthropic) describe(ctx context.Context, image []byte, prompt string) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(image)
	imgBlock := anthropicVisionImageBlock{Type: "image"}
	imgBlock.Source.Type = "base64"
	imgBlock.Source.MediaType = anthropicImageMediaType(image)
	imgBlock.Source.Data = b64

	content := []any{
		anthropicVisionTextBlock{Type: "text", Text: prompt},
		imgBlock,
	}

	reqBody := anthropicVisionRequest{
		Model:     a.model,
		MaxTokens: a.maxTokens,
		Messages: []anthropicVisionMsg{
			{Role: "user", Content: content},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic vision: %s", string(b))
	}
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	var text strings.Builder
	for _, c := range result.Content {
		if c.Type == "text" {
			text.WriteString(c.Text)
		}
	}
	out := strings.TrimSpace(text.String())
	if out == "" {
		return "", fmt.Errorf("anthropic vision: 空内容")
	}
	return out, nil
}

func (a *Anthropic) Describe(ctx context.Context, image []byte) (string, error) {
	return a.describe(ctx, image, defaultDescribePrompt)
}

func (a *Anthropic) DescribeWithPrompt(ctx context.Context, image []byte, prompt string) (string, error) {
	if prompt == "" {
		prompt = defaultDescribePrompt
	}
	return a.describe(ctx, image, prompt)
}

func (a *Anthropic) ProviderName() string { return "anthropic" }
func (a *Anthropic) ModelName() string    { return a.model }

var _ interfaces.CV = (*Anthropic)(nil)
