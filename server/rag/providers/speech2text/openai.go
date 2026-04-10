package speech2text

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// OpenAI Whisper 语音转文字实现
type OpenAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAI 创建 OpenAI Speech2Text
func NewOpenAI(apiKey, baseURL, model string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "whisper-1"
	}
	return &OpenAI{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type whisperResponse struct {
	Text string `json:"text"`
}

func (s *OpenAI) Transcribe(ctx context.Context, audio interface{}) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// model field
	_ = w.WriteField("model", s.model)

	// file field
	var reader io.Reader
	var filename string
	switch v := audio.(type) {
	case string:
		f, err := os.Open(v)
		if err != nil {
			return "", fmt.Errorf("open audio file: %w", err)
		}
		defer f.Close()
		reader = f
		filename = filepath.Base(v)
	case []byte:
		reader = bytes.NewReader(v)
		filename = "audio.wav"
	default:
		return "", fmt.Errorf("audio must be file path (string) or bytes ([]byte)")
	}

	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, reader); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/audio/transcriptions", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper api error: %s", string(b))
	}

	var result whisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Text), nil
}

func (s *OpenAI) ProviderName() string { return "openai" }
func (s *OpenAI) ModelName() string    { return s.model }

// ProviderNameAdapter 包装 Speech2Text 以自定义 ProviderName
type ProviderNameAdapter struct {
	interfaces.Speech2Text
	DisplayName string
}

func (p *ProviderNameAdapter) ProviderName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Speech2Text.ProviderName()
}
