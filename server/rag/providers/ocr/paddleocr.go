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
	"time"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// PaddleOCRHTTP 调用 PaddleOCR 云服务 API（与上游参考 PaddleOCRParser._send_request 一致：JSON 含 file Base64、fileType）
type PaddleOCRHTTP struct {
	apiURL      string
	accessToken string
	fileType    int
	timeout     time.Duration
	model       string
}

// NewPaddleOCRHTTP
// BaseURL：完整 API 地址（如 https://xxx.com/layout-parsing）
// APIKey：可选 JSON，含 paddleocr_api_url、access_token；或仅 access_token 字符串
// Extra：paddle_file_type（int，默认 0=PDF）、request_timeout_sec
func NewPaddleOCRHTTP(apiKey, baseURL, model string, extra map[string]any) (*PaddleOCRHTTP, error) {
	api := strings.TrimSpace(baseURL)
	token := ""
	if strings.TrimSpace(apiKey) != "" {
		if strings.HasPrefix(strings.TrimSpace(apiKey), "{") {
			var m map[string]any
			if err := json.Unmarshal([]byte(apiKey), &m); err != nil {
				return nil, fmt.Errorf("paddleocr: 解析 API Key JSON: %w", err)
			}
			if nested, ok := m["api_key"].(map[string]any); ok {
				m = nested
			}
			if v, ok := m["paddleocr_api_url"].(string); ok && strings.TrimSpace(v) != "" {
				api = strings.TrimSpace(v)
			}
			if v, ok := m["api_url"].(string); ok && strings.TrimSpace(v) != "" {
				api = strings.TrimSpace(v)
			}
			if v, ok := m["access_token"].(string); ok {
				token = strings.TrimSpace(v)
			}
		} else {
			token = strings.TrimSpace(apiKey)
		}
	}
	if api == "" {
		return nil, fmt.Errorf("paddleocr: 请配置 BaseURL（API 完整 URL）或 JSON 中的 paddleocr_api_url")
	}
	ft := 0
	timeout := 600 * time.Second
	if extra != nil {
		if v, ok := extra["paddle_file_type"].(float64); ok {
			ft = int(v)
		}
		if v, ok := extra["paddle_file_type"].(int); ok {
			ft = v
		}
		if v, ok := extra["request_timeout_sec"].(float64); ok && v > 0 {
			timeout = time.Duration(v) * time.Second
		}
	}
	return &PaddleOCRHTTP{
		apiURL:      strings.TrimSuffix(api, "/"),
		accessToken: token,
		fileType:    ft,
		timeout:     timeout,
		model:       model,
	}, nil
}

func (p *PaddleOCRHTTP) ExtractText(ctx context.Context, data []byte, filename string) (*interfaces.OCRResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("paddleocr: 空文件")
	}
	payload := map[string]any{
		"file":              base64.StdEncoding.EncodeToString(data),
		"fileType":          p.fileType,
		"prettifyMarkdown":  true,
		"showFormulaNumber": true,
		"visualize":         false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Platform", "lightningrag")
	if p.accessToken != "" {
		req.Header.Set("Authorization", "token "+p.accessToken)
	}
	client := &http.Client{Timeout: p.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paddleocr: HTTP %d %s", resp.StatusCode, truncateBytes(respBody, 400))
	}
	var top map[string]any
	if err := json.Unmarshal(respBody, &top); err != nil {
		return nil, fmt.Errorf("paddleocr: 响应非 JSON: %w", err)
	}
	if code := jsonNumToInt64(top["errorCode"]); code != 0 {
		msg, _ := top["errorMsg"].(string)
		return nil, fmt.Errorf("paddleocr: errorCode=%d %s", code, msg)
	}
	res, ok := top["result"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("paddleocr: 响应缺少 result")
	}
	text, sections := paddleLayoutToText(res)
	return &interfaces.OCRResult{Text: text, Sections: sections}, nil
}

func jsonNumToInt64(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	default:
		return -1
	}
}

func paddleLayoutToText(result map[string]any) (string, []string) {
	layouts, _ := result["layoutParsingResults"].([]any)
	var sections []string
	var full strings.Builder
	for _, page := range layouts {
		pm, ok := page.(map[string]any)
		if !ok {
			continue
		}
		pruned, _ := pm["prunedResult"].(map[string]any)
		if pruned == nil {
			continue
		}
		list, _ := pruned["parsing_res_list"].([]any)
		for _, blk := range list {
			bm, ok := blk.(map[string]any)
			if !ok {
				continue
			}
			content, _ := bm["block_content"].(string)
			content = strings.TrimSpace(content)
			if content == "" {
				continue
			}
			sections = append(sections, content)
			if full.Len() > 0 {
				full.WriteByte('\n')
			}
			full.WriteString(content)
		}
	}
	return full.String(), sections
}

func (p *PaddleOCRHTTP) ProviderName() string { return "paddleocr" }
func (p *PaddleOCRHTTP) ModelName() string {
	if p.model != "" {
		return p.model
	}
	return "PaddleOCR-VL"
}
