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

// Azure Azure OpenAI 文字转语音
type Azure struct {
	apiKey     string
	baseURL    string
	model      string
	apiVersion string
	client     *http.Client
}

// NewAzure 创建 Azure TTS；baseURL 为资源端点
func NewAzure(apiKey, baseURL, model, apiVersion string) *Azure {
	if baseURL == "" {
		baseURL = "https://YOUR_RESOURCE.openai.azure.com"
	}
	if apiVersion == "" {
		apiVersion = "2024-02-01"
	}
	if model == "" {
		model = "tts-1"
	}
	return &Azure{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		apiVersion: apiVersion,
		client:     &http.Client{},
	}
}

func (a *Azure) speechURL() string {
	return fmt.Sprintf("%s/openai/deployments/%s/audio/speech?api-version=%s",
		a.baseURL, a.model, a.apiVersion)
}

func (a *Azure) Synthesize(ctx context.Context, text string, voice string) (io.ReadCloser, error) {
	if voice == "" {
		voice = "alloy"
	}
	reqBody := ttsRequest{
		Model: a.model,
		Input: text,
		Voice: voice,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", a.speechURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("azure tts api error: %s", string(b))
	}

	return resp.Body, nil
}

func (a *Azure) ProviderName() string { return "azure" }
func (a *Azure) ModelName() string    { return a.model }

var _ interfaces.TTS = (*Azure)(nil)
