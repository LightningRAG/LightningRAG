package tts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DashScopeTTS 通义 Qwen TTS（multimodal-generation/generation），与阿里云文档 curl 示例一致
type DashScopeTTS struct {
	apiKey     string
	genURL     string
	model      string
	httpClient *http.Client
	extra      map[string]any
}

// NewDashScopeTTS apiRoot 如 https://dashscope.aliyuncs.com/api/v1；留空按区域默认
func NewDashScopeTTS(apiKey, apiRoot, model string, extra map[string]any) (*DashScopeTTS, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("dashscope tts: 需要 API Key")
	}
	region := ""
	if extra != nil {
		if v, ok := extra["dashscope_region"].(string); ok {
			region = strings.ToLower(strings.TrimSpace(v))
		}
	}
	if apiRoot != "" && strings.Contains(apiRoot, "compatible-mode") {
		apiRoot = ""
	}
	if apiRoot == "" {
		if region == "intl" || region == "singapore" {
			apiRoot = "https://dashscope-intl.aliyuncs.com/api/v1"
		} else {
			apiRoot = "https://dashscope.aliyuncs.com/api/v1"
		}
	}
	apiRoot = strings.TrimSuffix(apiRoot, "/")
	if model == "" {
		model = "qwen3-tts-flash"
	}
	return &DashScopeTTS{
		apiKey:     apiKey,
		genURL:     apiRoot + "/services/aigc/multimodal-generation/generation",
		model:      model,
		extra:      extra,
		httpClient: &http.Client{},
	}, nil
}

type dsTTSRequest struct {
	Model string `json:"model"`
	Input struct {
		Text         string `json:"text"`
		Voice        string `json:"voice"`
		LanguageType string `json:"language_type,omitempty"`
	} `json:"input"`
}

type dsTTSResponse struct {
	StatusCode int `json:"status_code"`
	Output     struct {
		Audio struct {
			URL  string `json:"url"`
			Data string `json:"data"`
		} `json:"audio"`
	} `json:"output"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func strFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func (d *DashScopeTTS) Synthesize(ctx context.Context, text string, voice string) (io.ReadCloser, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("dashscope tts: 文本为空")
	}
	v := strings.TrimSpace(voice)
	if v == "" {
		v = "Cherry"
	}
	lang := strFromMap(d.extra, "language_type")
	if lang == "" {
		lang = "Chinese"
	}
	var req dsTTSRequest
	req.Model = d.model
	req.Input.Text = text
	req.Input.Voice = v
	req.Input.LanguageType = lang
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.genURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dashscope tts: HTTP %d %s", resp.StatusCode, string(respBody))
	}
	var out dsTTSResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	if out.StatusCode != 200 {
		return nil, fmt.Errorf("dashscope tts: status=%d code=%s msg=%s", out.StatusCode, out.Code, out.Message)
	}
	audioURL := strings.TrimSpace(out.Output.Audio.URL)
	if audioURL != "" {
		getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
		if err != nil {
			return nil, err
		}
		getResp, err := d.httpClient.Do(getReq)
		if err != nil {
			return nil, err
		}
		if getResp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(getResp.Body)
			getResp.Body.Close()
			return nil, fmt.Errorf("dashscope tts: 下载音频 %d %s", getResp.StatusCode, string(b))
		}
		return getResp.Body, nil
	}
	data := strings.TrimSpace(out.Output.Audio.Data)
	if data == "" {
		return nil, fmt.Errorf("dashscope tts: 响应无音频 url 与 data")
	}
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("dashscope tts: 解码 base64: %w", err)
	}
	return io.NopCloser(bytes.NewReader(raw)), nil
}

func (d *DashScopeTTS) ProviderName() string { return "dashscope" }
func (d *DashScopeTTS) ModelName() string    { return d.model }
