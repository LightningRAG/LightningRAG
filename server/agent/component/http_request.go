package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("HTTPRequest", NewHTTPRequest)
}

// HTTPRequest HTTP 请求组件，调用远程 API
type HTTPRequest struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewHTTPRequest 创建 HTTPRequest 组件
func NewHTTPRequest(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &HTTPRequest{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (h *HTTPRequest) ComponentName() string {
	return "HTTPRequest"
}

// Invoke 执行 HTTP 请求
func (h *HTTPRequest) Invoke(inputs map[string]any) error {
	h.mu.Lock()
	h.err = ""
	h.mu.Unlock()

	urlStr := h.canvas.ResolveString(getStrParam(h.params, "url"))
	if urlStr == "" {
		urlStr = getStrParam(h.params, "url")
	}
	if urlStr == "" {
		h.mu.Lock()
		h.err = "HTTPRequest url 为空"
		h.mu.Unlock()
		return fmt.Errorf("HTTPRequest url 为空")
	}

	method := strings.ToUpper(getStrParam(h.params, "method"))
	if method == "" {
		method = "GET"
	}
	timeoutSec := getIntParam(h.params, "timeout", 60)
	cleanHTML := getBoolParam(h.params, "clean_html")

	headers := h.getHeaders()
	paramsMap := h.getParams()

	// GET: 参数拼到 URL
	if method == "GET" && len(paramsMap) > 0 {
		u, err := url.Parse(urlStr)
		if err == nil {
			q := u.Query()
			for k, v := range paramsMap {
				q.Set(k, v)
			}
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}
	}

	var body io.Reader
	if method == "POST" || method == "PUT" {
		bodyBytes, _ := json.Marshal(paramsMap)
		body = bytes.NewReader(bodyBytes)
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "application/json"
		}
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		h.mu.Lock()
		h.err = err.Error()
		h.mu.Unlock()
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	if proxyURL := getStrParam(h.params, "proxy"); proxyURL != "" {
		pu, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(pu)}
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		h.mu.Lock()
		h.err = err.Error()
		h.mu.Unlock()
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		h.mu.Lock()
		h.err = err.Error()
		h.mu.Unlock()
		return err
	}

	result := string(b)
	if cleanHTML {
		result = stripHTML(result)
	}

	h.mu.Lock()
	h.output["result"] = result
	h.output["status_code"] = resp.StatusCode
	h.mu.Unlock()
	return nil
}

func (h *HTTPRequest) getHeaders() map[string]string {
	out := make(map[string]string)
	v, ok := h.params["headers"]
	if !ok || v == nil {
		return out
	}
	m, ok := v.(map[string]any)
	if !ok {
		return out
	}
	for k, val := range m {
		var s string
		switch v := val.(type) {
		case string:
			s = v
		default:
			s = fmt.Sprint(v)
		}
		out[k] = h.canvas.ResolveString(s)
	}
	return out
}

func (h *HTTPRequest) getParams() map[string]string {
	out := make(map[string]string)
	v, ok := h.params["params"]
	if !ok || v == nil {
		return out
	}
	m, ok := v.(map[string]any)
	if !ok {
		return out
	}
	for k, val := range m {
		var s string
		switch v := val.(type) {
		case string:
			s = v
		default:
			s = fmt.Sprint(v)
		}
		out[k] = h.canvas.ResolveString(s)
	}
	return out
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	return htmlTagRe.ReplaceAllString(s, "")
}

// Output 获取输出
func (h *HTTPRequest) Output(key string) any {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.output[key]
}

// OutputAll 获取所有输出
func (h *HTTPRequest) OutputAll() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range h.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (h *HTTPRequest) SetOutput(key string, value any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.output[key] = value
}

// Error 返回错误
func (h *HTTPRequest) Error() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.err
}

// Reset 重置
func (h *HTTPRequest) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.output = make(map[string]any)
	h.err = ""
}
