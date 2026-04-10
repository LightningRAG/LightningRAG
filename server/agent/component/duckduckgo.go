package component

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("DuckDuckGo", NewDuckDuckGo)
}

// DuckDuckGo 使用 DuckDuckGo Instant Answer JSON API 做网页检索（参考 references 目录内 agent/tools/duckduckgo.py）
type DuckDuckGo struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewDuckDuckGo 创建 DuckDuckGo 组件
func NewDuckDuckGo(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &DuckDuckGo{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (d *DuckDuckGo) ComponentName() string { return "DuckDuckGo" }

func (d *DuckDuckGo) Invoke(map[string]any) error {
	d.mu.Lock()
	d.err = ""
	d.mu.Unlock()

	q := getStrParam(d.params, "query")
	if q == "" {
		q = "sys.query"
	}
	query := resolveWorkflowQuery(d.canvas, q)
	if query == "" {
		d.mu.Lock()
		d.output["formalized_content"] = ""
		d.output["json"] = []any{}
		d.mu.Unlock()
		return nil
	}

	channel := strings.ToLower(getStrParam(d.params, "channel"))
	if channel == "news" {
		query = query + " news"
	}

	topN := getIntParam(d.params, "top_n", 10)
	if topN < 1 {
		topN = 1
	}
	if topN > 25 {
		topN = 25
	}

	timeoutSec := getIntParam(d.params, "timeout", 15)
	if timeoutSec < 5 {
		timeoutSec = 5
	}
	if timeoutSec > 60 {
		timeoutSec = 60
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	u := "https://api.duckduckgo.com/?format=json&no_html=1&skip_disambig=1&q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		d.fail(err.Error())
		return err
	}
	req.Header.Set("User-Agent", "LightningRAG-Agent/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		d.fail(err.Error())
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		d.fail(err.Error())
		return err
	}

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		d.fail(err.Error())
		return err
	}

	snips := ddgCollectSnippets(root, topN)
	formalized := formatWebSnippets(snips)
	jsonRows := make([]any, len(snips))
	for i, s := range snips {
		jsonRows[i] = map[string]any{"title": s.Title, "url": s.URL, "body": s.Body}
	}

	d.mu.Lock()
	d.output["formalized_content"] = formalized
	d.output["json"] = jsonRows
	d.mu.Unlock()
	return nil
}

func ddgCollectSnippets(root map[string]any, topN int) []webSnippet {
	var out []webSnippet
	if at, ok := root["AbstractText"].(string); ok && strings.TrimSpace(at) != "" {
		u, _ := root["AbstractURL"].(string)
		title := "Summary"
		if h, ok := root["Heading"].(string); ok && strings.TrimSpace(h) != "" {
			title = h
		}
		out = append(out, webSnippet{Title: title, URL: u, Body: at})
	}
	if len(out) >= topN {
		return out[:topN]
	}
	ddgWalkRelated(root["RelatedTopics"], &out, topN)
	if len(out) > topN {
		return out[:topN]
	}
	return out
}

func ddgWalkRelated(v any, out *[]webSnippet, limit int) {
	if len(*out) >= limit {
		return
	}
	switch t := v.(type) {
	case []any:
		for _, item := range t {
			ddgWalkRelated(item, out, limit)
			if len(*out) >= limit {
				return
			}
		}
	case map[string]any:
		if text, ok := t["Text"].(string); ok && strings.TrimSpace(text) != "" {
			u, _ := t["FirstURL"].(string)
			*out = append(*out, webSnippet{Title: "", URL: u, Body: text})
		}
		if topics, ok := t["Topics"]; ok {
			ddgWalkRelated(topics, out, limit)
		}
	}
}

func (d *DuckDuckGo) fail(msg string) {
	d.mu.Lock()
	d.err = msg
	d.mu.Unlock()
}

func (d *DuckDuckGo) Output(key string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.output[key]
}

func (d *DuckDuckGo) OutputAll() map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range d.output {
		out[k] = v
	}
	return out
}

func (d *DuckDuckGo) SetOutput(key string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.output[key] = value
}

func (d *DuckDuckGo) Error() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.err
}

func (d *DuckDuckGo) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.output = make(map[string]any)
	d.err = ""
}
