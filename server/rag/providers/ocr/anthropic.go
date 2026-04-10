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

// Anthropic Claude 视觉 OCR（Messages API，与 CV/Anthropic 一致，提示词使用 ocrPrompt）
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

type ocrAnthropicMsg struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ocrAnthropicReq struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	Messages  []ocrAnthropicMsg `json:"messages"`
}

type ocrAnthropicImg struct {
	Type   string `json:"type"`
	Source struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source"`
}

type ocrAnthropicTxt struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func ocrAnthropicMime(image []byte) string {
	if len(image) >= 2 && image[0] == 0xFF && image[1] == 0xD8 {
		return "image/jpeg"
	}
	return "image/png"
}

func (a *Anthropic) ExtractText(ctx context.Context, data []byte, filename string) (*interfaces.OCRResult, error) {
	_ = filename
	b64 := base64.StdEncoding.EncodeToString(data)
	img := ocrAnthropicImg{Type: "image"}
	img.Source.Type = "base64"
	img.Source.MediaType = ocrAnthropicMime(data)
	img.Source.Data = b64
	content := []any{
		ocrAnthropicTxt{Type: "text", Text: ocrPrompt},
		img,
	}
	body, _ := json.Marshal(ocrAnthropicReq{
		Model:     a.model,
		MaxTokens: a.maxTokens,
		Messages:  []ocrAnthropicMsg{{Role: "user", Content: content}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic ocr: %s", string(b))
	}
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var sb strings.Builder
	for _, c := range result.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return nil, fmt.Errorf("anthropic ocr: 空内容")
	}
	return &interfaces.OCRResult{Text: text, Sections: []string{text}}, nil
}

func (a *Anthropic) ProviderName() string { return "anthropic" }
func (a *Anthropic) ModelName() string    { return a.model }
