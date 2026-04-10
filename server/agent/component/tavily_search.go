package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("TavilySearch", NewTavilySearch)
}

// TavilySearch 调用 Tavily Search API（参考 references 目录内 agent/tools/tavily.py）
type TavilySearch struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewTavilySearch 创建 TavilySearch 组件
func NewTavilySearch(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &TavilySearch{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (t *TavilySearch) ComponentName() string { return "TavilySearch" }

func (t *TavilySearch) Invoke(map[string]any) error {
	t.mu.Lock()
	t.err = ""
	t.mu.Unlock()

	q := getStrParam(t.params, "query")
	if q == "" {
		q = "sys.query"
	}
	query := resolveWorkflowQuery(t.canvas, q)
	if query == "" {
		t.mu.Lock()
		t.output["formalized_content"] = ""
		t.output["json"] = []any{}
		t.mu.Unlock()
		return nil
	}

	apiKey := getStrParam(t.params, "api_key")
	if apiKey == "" {
		apiKey = strings.TrimSpace(getStrParam(t.canvas.GetGlobals(), "env.tavily_api_key"))
	}
	if apiKey == "" {
		err := fmt.Errorf("Tavily 需要 api_key 参数，或在 globals 中设置 env.tavily_api_key")
		t.fail(err.Error())
		return err
	}

	maxResults := getIntParam(t.params, "max_results", 6)
	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 20 {
		maxResults = 20
	}

	topic := strings.ToLower(getStrParam(t.params, "topic"))
	if topic != "news" {
		topic = "general"
	}
	depth := strings.ToLower(getStrParam(t.params, "search_depth"))
	if depth != "advanced" {
		depth = "basic"
	}

	payload := map[string]any{
		"api_key":        apiKey,
		"query":          query,
		"topic":          topic,
		"search_depth":   depth,
		"max_results":    maxResults,
		"include_answer": getBoolParam(t.params, "include_answer"),
	}
	if inc := getStrSliceParam(t.params, "include_domains"); len(inc) > 0 {
		payload["include_domains"] = inc
	} else if csv := getStrParam(t.params, "include_domains_csv"); csv != "" {
		if parts := splitCommaFields(csv); len(parts) > 0 {
			payload["include_domains"] = parts
		}
	}
	if exc := getStrSliceParam(t.params, "exclude_domains"); len(exc) > 0 {
		payload["exclude_domains"] = exc
	} else if csv := getStrParam(t.params, "exclude_domains_csv"); csv != "" {
		if parts := splitCommaFields(csv); len(parts) > 0 {
			payload["exclude_domains"] = parts
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.fail(err.Error())
		return err
	}

	timeoutSec := getIntParam(t.params, "timeout", 30)
	if timeoutSec < 5 {
		timeoutSec = 5
	}
	if timeoutSec > 120 {
		timeoutSec = 120
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(body))
	if err != nil {
		t.fail(err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "LightningRAG-Agent/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.fail(err.Error())
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		t.fail(err.Error())
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("Tavily HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		t.fail(err.Error())
		return err
	}

	var root struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Raw     string  `json:"raw_content"`
			Score   float64 `json:"score"`
		} `json:"results"`
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal(respBody, &root); err != nil {
		t.fail(err.Error())
		return err
	}

	var snips []webSnippet
	var jsonRows []any
	if strings.TrimSpace(root.Answer) != "" {
		snips = append(snips, webSnippet{Title: "Answer", URL: "", Body: root.Answer})
		jsonRows = append(jsonRows, map[string]any{"title": "Answer", "url": "", "body": root.Answer})
	}
	for _, r := range root.Results {
		bodyText := strings.TrimSpace(r.Raw)
		if bodyText == "" {
			bodyText = strings.TrimSpace(r.Content)
		}
		snips = append(snips, webSnippet{Title: r.Title, URL: r.URL, Body: bodyText})
		jsonRows = append(jsonRows, map[string]any{
			"title": r.Title, "url": r.URL, "body": bodyText, "score": r.Score,
		})
	}

	formalized := formatWebSnippets(snips)
	t.mu.Lock()
	t.output["formalized_content"] = formalized
	t.output["json"] = jsonRows
	if root.Answer != "" {
		t.output["answer"] = root.Answer
	}
	t.mu.Unlock()
	return nil
}

func (t *TavilySearch) fail(msg string) {
	t.mu.Lock()
	t.err = msg
	t.mu.Unlock()
}

func (t *TavilySearch) Output(key string) any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.output[key]
}

func (t *TavilySearch) OutputAll() map[string]any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range t.output {
		out[k] = v
	}
	return out
}

func (t *TavilySearch) SetOutput(key string, value any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output[key] = value
}

func (t *TavilySearch) Error() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.err
}

func (t *TavilySearch) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output = make(map[string]any)
	t.err = ""
}
