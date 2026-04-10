package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// OpenAI TTS 文字转语音实现
type OpenAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAI 创建 OpenAI TTS
func NewOpenAI(apiKey, baseURL, model string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "tts-1"
	}
	return &OpenAI{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type ttsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Voice string `json:"voice"`
}

func (t *OpenAI) Synthesize(ctx context.Context, text string, voice string) (io.ReadCloser, error) {
	if voice == "" {
		voice = "alloy"
	}
	reqBody := ttsRequest{
		Model: t.model,
		Input: text,
		Voice: voice,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("tts api error: %s", string(b))
	}

	return resp.Body, nil
}

func (t *OpenAI) ProviderName() string { return "openai" }
func (t *OpenAI) ModelName() string    { return t.model }

// ProviderNameAdapter 包装 TTS 以自定义 ProviderName
type ProviderNameAdapter struct {
	interfaces.TTS
	DisplayName string
}

func (p *ProviderNameAdapter) ProviderName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.TTS.ProviderName()
}
