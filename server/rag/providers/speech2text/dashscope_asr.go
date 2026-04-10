package speech2text

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DashScopeASR 通义 Qwen ASR（OpenAI 兼容 chat/completions + input_audio），参考阿里云文档
type DashScopeASR struct {
	apiKey     string
	chatURL    string
	model      string
	httpClient *http.Client
}

// NewDashScopeASR chatBase 为 compatible-mode 根地址，如 https://dashscope.aliyuncs.com/compatible-mode/v1
func NewDashScopeASR(apiKey, chatBase, model string, extra map[string]any) (*DashScopeASR, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("dashscope asr: 需要 API Key")
	}
	region := ""
	if extra != nil {
		if v, ok := extra["dashscope_region"].(string); ok {
			region = strings.ToLower(strings.TrimSpace(v))
		}
	}
	if chatBase != "" && !strings.Contains(chatBase, "compatible-mode") {
		chatBase = ""
	}
	if chatBase == "" {
		if region == "intl" || region == "singapore" {
			chatBase = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
		} else {
			chatBase = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		}
	}
	chatBase = strings.TrimSuffix(chatBase, "/")
	if model == "" {
		model = "qwen3-asr-flash"
	}
	return &DashScopeASR{
		apiKey:     apiKey,
		chatURL:    chatBase + "/chat/completions",
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

type dsASRRequest struct {
	Model    string         `json:"model"`
	Messages []dsASRMessage `json:"messages"`
	Stream   bool           `json:"stream"`
	ASROpts  map[string]any `json:"asr_options,omitempty"`
}

type dsASRMessage struct {
	Role    string      `json:"role"`
	Content []dsASRPart `json:"content"`
}

type dsASRPart struct {
	Type       string         `json:"type"`
	InputAudio map[string]any `json:"input_audio"`
}

type dsASRResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (d *DashScopeASR) Transcribe(ctx context.Context, audio interface{}) (string, error) {
	raw, mime, err := readAudioBytes(audio)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(raw)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b64)

	enableITN := false
	req := dsASRRequest{
		Model: d.model,
		Messages: []dsASRMessage{
			{
				Role: "user",
				Content: []dsASRPart{
					{
						Type: "input_audio",
						InputAudio: map[string]any{
							"data": dataURL,
						},
					},
				},
			},
		},
		Stream: false,
		ASROpts: map[string]any{
			"enable_itn": enableITN,
		},
	}
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.chatURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dashscope asr: HTTP %d %s", resp.StatusCode, string(respBody))
	}
	var out dsASRResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", err
	}
	if out.Error.Message != "" {
		return "", fmt.Errorf("dashscope asr: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("dashscope asr: 空 choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func readAudioBytes(audio interface{}) ([]byte, string, error) {
	switch v := audio.(type) {
	case string:
		b, err := os.ReadFile(v)
		if err != nil {
			return nil, "", err
		}
		return b, mimeForAudioPath(v), nil
	case []byte:
		return v, "audio/wav", nil
	default:
		return nil, "", fmt.Errorf("audio 需为文件路径或 []byte")
	}
}

func mimeForAudioPath(p string) string {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".mp3", ".mpeg":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".m4a":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	case ".ogg":
		return "audio/ogg"
	default:
		return "audio/wav"
	}
}

func (d *DashScopeASR) ProviderName() string { return "dashscope" }
func (d *DashScopeASR) ModelName() string    { return d.model }
