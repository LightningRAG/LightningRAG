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

// Azure Azure OpenAI 语音转写（Whisper 等部署）
type Azure struct {
	apiKey     string
	baseURL    string
	model      string
	apiVersion string
	client     *http.Client
}

// NewAzure 创建 Azure Speech2Text；baseURL 为资源端点
func NewAzure(apiKey, baseURL, model, apiVersion string) *Azure {
	if baseURL == "" {
		baseURL = "https://YOUR_RESOURCE.openai.azure.com"
	}
	if apiVersion == "" {
		apiVersion = "2024-02-01"
	}
	if model == "" {
		model = "whisper-1"
	}
	return &Azure{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		apiVersion: apiVersion,
		client:     &http.Client{},
	}
}

func (a *Azure) transcriptionURL() string {
	return fmt.Sprintf("%s/openai/deployments/%s/audio/transcriptions?api-version=%s",
		a.baseURL, a.model, a.apiVersion)
}

func (a *Azure) Transcribe(ctx context.Context, audio interface{}) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("model", a.model)

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

	req, err := http.NewRequestWithContext(ctx, "POST", a.transcriptionURL(), &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("api-key", a.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("azure whisper api error: %s", string(b))
	}

	var result whisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Text), nil
}

func (a *Azure) ProviderName() string { return "azure" }
func (a *Azure) ModelName() string    { return a.model }

var _ interfaces.Speech2Text = (*Azure)(nil)
