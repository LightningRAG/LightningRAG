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

const ocrPrompt = "Extract all text from this image. Preserve the structure and layout. Return only the extracted text, nothing else."

// OpenAI OCR 实现，使用 Vision 模型提取图片中的文字
type OpenAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAI 创建 OpenAI OCR
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

type ocrContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type ocrMessage struct {
	Role    string           `json:"role"`
	Content []ocrContentPart `json:"content"`
}

type ocrRequest struct {
	Model    string       `json:"model"`
	Messages []ocrMessage `json:"messages"`
}

type ocrChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type ocrResponse struct {
	Choices []ocrChoice `json:"choices"`
}

func (o *OpenAI) imageToDataURL(data []byte) string {
	mime := "image/png"
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		mime = "image/jpeg"
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, b64)
}

func (o *OpenAI) ExtractText(ctx context.Context, data []byte, filename string) (*interfaces.OCRResult, error) {
	dataURL := o.imageToDataURL(data)
	reqBody := ocrRequest{
		Model: o.model,
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

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ocr api error: %s", string(b))
	}

	var result ocrResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty ocr response")
	}
	text := strings.TrimSpace(result.Choices[0].Message.Content)
	return &interfaces.OCRResult{
		Text:     text,
		Sections: []string{text},
	}, nil
}

func (o *OpenAI) ProviderName() string { return "openai" }
func (o *OpenAI) ModelName() string    { return o.model }
