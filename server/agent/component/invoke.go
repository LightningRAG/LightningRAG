package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("Invoke", NewInvoke)
}

// Invoke 对齐上游编排 Invoke：通过 variables（key + ref/value）拼装 GET 查询或 POST/PUT 体
type Invoke struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewInvoke 创建 Invoke
func NewInvoke(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Invoke{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (iv *Invoke) ComponentName() string { return "Invoke" }

func (iv *Invoke) Invoke(_ map[string]any) error {
	iv.mu.Lock()
	iv.err = ""
	iv.mu.Unlock()

	urlStr := iv.canvas.ResolveString(getStrParam(iv.params, "url"))
	if urlStr == "" {
		urlStr = getStrParam(iv.params, "url")
	}
	if urlStr == "" {
		iv.fail("Invoke url 为空")
		return fmt.Errorf("Invoke url 为空")
	}

	method := strings.ToUpper(getStrParam(iv.params, "method"))
	if method == "" {
		method = "GET"
	}
	datatype := strings.ToLower(getStrParam(iv.params, "datatype"))
	if datatype == "" {
		datatype = "json"
	}
	timeoutSec := getIntParam(iv.params, "timeout", 60)

	args, err := iv.buildArgs()
	if err != nil {
		iv.fail(err.Error())
		return err
	}

	headers := iv.parseHeaders()

	var body io.Reader
	if method == "GET" {
		if len(args) > 0 {
			u, err := url.Parse(urlStr)
			if err != nil {
				iv.fail(err.Error())
				return err
			}
			q := u.Query()
			for k, v := range args {
				q.Set(k, fmt.Sprint(v))
			}
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}
	} else {
		switch datatype {
		case "json":
			b, err := json.Marshal(args)
			if err != nil {
				iv.fail(err.Error())
				return err
			}
			body = bytes.NewReader(b)
			if _, ok := headers["Content-Type"]; !ok {
				headers["Content-Type"] = "application/json"
			}
		case "formdata", "form", "multipart":
			buf := &bytes.Buffer{}
			w := multipart.NewWriter(buf)
			for k, v := range args {
				_ = w.WriteField(k, fmt.Sprint(v))
			}
			if err := w.Close(); err != nil {
				iv.fail(err.Error())
				return err
			}
			body = buf
			headers["Content-Type"] = w.FormDataContentType()
		default:
			form := url.Values{}
			for k, v := range args {
				form.Set(k, fmt.Sprint(v))
			}
			body = strings.NewReader(form.Encode())
			if _, ok := headers["Content-Type"]; !ok {
				headers["Content-Type"] = "application/x-www-form-urlencoded"
			}
		}
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		iv.fail(err.Error())
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		iv.fail(err.Error())
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		iv.fail(err.Error())
		return err
	}

	text := string(respBody)
	iv.mu.Lock()
	iv.output["content"] = text
	iv.output["status_code"] = resp.StatusCode
	iv.output["formalized_content"] = text
	if len(text) > 8000 {
		iv.output["formalized_content"] = text[:8000] + "..."
	}
	iv.mu.Unlock()
	return nil
}

func (iv *Invoke) buildArgs() (map[string]any, error) {
	out := make(map[string]any)
	raw := iv.params["variables"]
	arr, ok := raw.([]any)
	if !ok {
		return out, nil
	}
	for _, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		key := getStrParam(m, "key")
		if key == "" {
			continue
		}
		if ref := getStrParam(m, "ref"); ref != "" {
			if v, ok := iv.canvas.GetVariableValue(ref); ok {
				out[key] = v
				continue
			}
			out[key] = iv.canvas.ResolveString(ref)
			continue
		}
		if lit, ok := m["value"]; ok {
			out[key] = lit
			continue
		}
		if s := getStrParam(m, "value_str"); s != "" {
			out[key] = s
		}
	}
	return out, nil
}

func (iv *Invoke) parseHeaders() map[string]string {
	h := make(map[string]string)
	raw := iv.params["headers"]
	if raw == nil {
		return h
	}
	if m, ok := raw.(map[string]any); ok {
		for k, v := range m {
			h[k] = fmt.Sprint(v)
		}
		return h
	}
	s := strings.TrimSpace(fmt.Sprint(raw))
	if s == "" {
		return h
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return h
	}
	for k, v := range obj {
		h[k] = fmt.Sprint(v)
	}
	return h
}

func (iv *Invoke) fail(msg string) {
	iv.mu.Lock()
	iv.err = msg
	iv.mu.Unlock()
}

func (iv *Invoke) Output(key string) any {
	iv.mu.RLock()
	defer iv.mu.RUnlock()
	return iv.output[key]
}

func (iv *Invoke) OutputAll() map[string]any {
	iv.mu.RLock()
	defer iv.mu.RUnlock()
	cp := make(map[string]any)
	for k, v := range iv.output {
		cp[k] = v
	}
	return cp
}

func (iv *Invoke) SetOutput(key string, value any) {
	iv.mu.Lock()
	defer iv.mu.Unlock()
	iv.output[key] = value
}

func (iv *Invoke) Error() string {
	iv.mu.RLock()
	defer iv.mu.RUnlock()
	return iv.err
}

func (iv *Invoke) Reset() {
	iv.mu.Lock()
	defer iv.mu.Unlock()
	iv.output = make(map[string]any)
	iv.err = ""
}
