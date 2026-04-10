package ocr

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// MinerUOCR 调用 MinerU HTTP API（与上游参考 MinerUParser._run_mineru_api 一致：POST {apiserver}/file_parse，返回 ZIP）
type MinerUOCR struct {
	apiServer   string
	serverURL   string
	backend     string
	parseMethod string
	langList    string
	client      *http.Client
	model       string
}

// NewMinerUOCR
// BaseURL：MinerU API 根地址（对应 mineru_apiserver），如 http://127.0.0.1:8000
// APIKey：JSON，可含 mineru_server_url、mineru_backend、mineru_apiserver 等（与上游参考字段一致）；支持嵌套 api_key
// ModelName：非空时作为 lang_list（如 ch、en）
// Extra：可覆盖 mineru_backend、mineru_parse_method、lang_list
func NewMinerUOCR(apiKey, baseURL, model string, extra map[string]any) (*MinerUOCR, error) {
	cfg, err := parseMinerUKeyJSON(apiKey)
	if err != nil && apiKey != "" && !strings.HasPrefix(strings.TrimSpace(apiKey), "{") {
		return nil, fmt.Errorf("mineru: API Key 须为 JSON 或留空仅用 BaseURL")
	}
	api := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if cfg != nil {
		if v := strFromMinerMap(cfg, "mineru_apiserver", "MINERU_APISERVER"); v != "" {
			api = strings.TrimSuffix(v, "/")
		}
	}
	if api == "" {
		return nil, fmt.Errorf("mineru: 请配置 BaseURL（MinerU API 地址）或 JSON 中的 mineru_apiserver")
	}
	srvURL := ""
	backend := "pipeline"
	parseMethod := "auto"
	lang := strings.TrimSpace(model)
	if cfg != nil {
		if v := strFromMinerMap(cfg, "mineru_server_url", "MINERU_SERVER_URL"); v != "" {
			srvURL = strings.TrimSuffix(v, "/")
		}
		if v := strFromMinerMap(cfg, "mineru_backend", "MINERU_BACKEND"); v != "" {
			backend = v
		}
	}
	if extra != nil {
		if v, ok := extra["mineru_server_url"].(string); ok && strings.TrimSpace(v) != "" {
			srvURL = strings.TrimSuffix(strings.TrimSpace(v), "/")
		}
		if v, ok := extra["mineru_backend"].(string); ok && v != "" {
			backend = v
		}
		if v, ok := extra["mineru_parse_method"].(string); ok && v != "" {
			parseMethod = v
		}
		if v, ok := extra["lang_list"].(string); ok && v != "" {
			lang = v
		}
	}
	return &MinerUOCR{
		apiServer:   api,
		serverURL:   srvURL,
		backend:     backend,
		parseMethod: parseMethod,
		langList:    lang,
		model:       model,
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}, nil
}

func parseMinerUKeyJSON(apiKey string) (map[string]any, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, nil
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(apiKey), &raw); err != nil {
		return nil, err
	}
	if nested, ok := raw["api_key"].(map[string]any); ok {
		return nested, nil
	}
	return raw, nil
}

func strFromMinerMap(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			if s := strings.TrimSpace(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func (m *MinerUOCR) ExtractText(ctx context.Context, data []byte, filename string) (*interfaces.OCRResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("mineru: 空文件")
	}
	stem := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	stem = strings.TrimSpace(stem)
	if stem == "" {
		stem = "document"
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".pdf"
	}
	uploadName := stem + ext

	var body bytes.Buffer
	mpw := multipart.NewWriter(&body)
	part, err := mpw.CreateFormFile("files", uploadName)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}
	fields := map[string]string{
		"output_dir":          "./output",
		"backend":             m.backend,
		"parse_method":        m.parseMethod,
		"formula_enable":      "true",
		"table_enable":        "true",
		"return_md":           "true",
		"return_middle_json":  "true",
		"return_model_output": "true",
		"return_content_list": "true",
		"return_images":       "true",
		"response_format_zip": "true",
		"start_page_id":       "0",
		"end_page_id":         "99999",
	}
	if m.langList != "" {
		fields["lang_list"] = m.langList
	}
	if m.serverURL != "" {
		fields["server_url"] = m.serverURL
	}
	for k, v := range fields {
		if err := mpw.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := mpw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.apiServer+"/file_parse", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mpw.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mineru: HTTP %d %s", resp.StatusCode, truncateBytes(zipBytes, 500))
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(ct), "zip") && !isZipMagic(zipBytes) {
		return nil, fmt.Errorf("mineru: 期望 ZIP 响应，Content-Type=%s body=%s", ct, truncateBytes(zipBytes, 300))
	}
	blocks, err := parseMinerUContentListFromZip(zipBytes, stem)
	if err != nil {
		return nil, err
	}
	text, sections := mineruBlocksToText(blocks)
	return &interfaces.OCRResult{Text: text, Sections: sections}, nil
}

func isZipMagic(b []byte) bool {
	return len(b) >= 4 && b[0] == 0x50 && b[1] == 0x4b && (b[2] == 0x03 || b[2] == 0x05 || b[2] == 0x07)
}

func truncateBytes(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

func parseMinerUContentListFromZip(zipBytes []byte, stem string) ([]map[string]any, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("mineru: 解压 ZIP: %w", err)
	}
	safeStem := sanitizeMinerUFilename(stem)
	names := []string{
		stem + "_content_list.json",
		safeStem + "_content_list.json",
	}
	var raw []byte
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		base := filepath.Base(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.Contains(base, "..") {
			continue
		}
		for _, want := range names {
			if base == want {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				raw, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return nil, err
				}
				break
			}
		}
		if raw != nil {
			break
		}
	}
	if raw == nil {
		// 任取第一个 *_content_list.json
		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}
			base := filepath.Base(strings.ReplaceAll(f.Name, "\\", "/"))
			if strings.Contains(base, "..") {
				continue
			}
			if strings.HasSuffix(base, "_content_list.json") {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				raw, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("mineru: ZIP 内未找到 content_list.json")
	}
	var blocks []map[string]any
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil, fmt.Errorf("mineru: 解析 JSON: %w", err)
	}
	return blocks, nil
}

func sanitizeMinerUFilename(name string) string {
	s := name
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "..", "")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" || strings.HasPrefix(out, ".") {
		return "unnamed"
	}
	return out
}

func mineruBlocksToText(blocks []map[string]any) (string, []string) {
	var sections []string
	var full strings.Builder
	for _, b := range blocks {
		typ, _ := b["type"].(string)
		var seg string
		switch typ {
		case "text", "equation":
			if t, ok := b["text"].(string); ok {
				seg = t
			}
		case "table":
			if t, ok := b["table_body"].(string); ok {
				seg = t
			}
			seg += joinMineruStrSlice(b["table_caption"])
			seg += joinMineruStrSlice(b["table_footnote"])
			if strings.TrimSpace(seg) == "" {
				seg = "FAILED TO PARSE TABLE"
			}
		case "image":
			seg = joinMineruStrSlice(b["image_caption"]) + joinMineruStrSlice(b["image_footnote"])
		case "code":
			if t, ok := b["code_body"].(string); ok {
				seg = t
			}
			seg += joinMineruStrSlice(b["code_caption"])
		case "list":
			if arr, ok := b["list_items"].([]any); ok {
				var lines []string
				for _, it := range arr {
					if s, ok := it.(string); ok {
						lines = append(lines, s)
					}
				}
				seg = strings.Join(lines, "\n")
			}
		case "discarded":
			continue
		default:
			if t, ok := b["text"].(string); ok {
				seg = t
			}
		}
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		sections = append(sections, seg)
		if full.Len() > 0 {
			full.WriteByte('\n')
		}
		full.WriteString(seg)
	}
	return full.String(), sections
}

func joinMineruStrSlice(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, x := range arr {
		if s, ok := x.(string); ok {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n")
}

func (m *MinerUOCR) ProviderName() string { return "mineru" }
func (m *MinerUOCR) ModelName() string    { return m.model }
