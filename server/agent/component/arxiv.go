package component

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("ArXiv", NewArXiv)
}

// ArXiv 使用 arXiv Atom API 检索论文摘要（参考 references 目录内 agent/tools/arxiv.py，component_name 与上游一致）
type ArXiv struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewArXiv 创建 ArXiv 组件
func NewArXiv(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &ArXiv{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (a *ArXiv) ComponentName() string { return "ArXiv" }

func (a *ArXiv) Invoke(map[string]any) error {
	a.mu.Lock()
	a.err = ""
	a.mu.Unlock()

	q := getStrParam(a.params, "query")
	if q == "" {
		q = "sys.query"
	}
	query := resolveWorkflowQuery(a.canvas, q)
	if query == "" {
		a.mu.Lock()
		a.output["formalized_content"] = ""
		a.output["json"] = []any{}
		a.mu.Unlock()
		return nil
	}

	topN := getIntParam(a.params, "top_n", 10)
	if topN < 1 {
		topN = 1
	}
	if topN > 50 {
		topN = 50
	}

	sortBy := strings.ToLower(getStrParam(a.params, "sort_by"))
	if sortBy == "" {
		sortBy = "submitteddate"
	}
	sortParam := "submittedDate"
	switch sortBy {
	case "lastupdateddate", "last_updated_date":
		sortParam = "lastUpdatedDate"
	case "relevance":
		sortParam = "relevance"
	case "submitteddate", "submitted_date":
		sortParam = "submittedDate"
	}

	timeoutSec := getIntParam(a.params, "timeout", 20)
	if timeoutSec < 5 {
		timeoutSec = 5
	}
	if timeoutSec > 90 {
		timeoutSec = 90
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	apiURL := "http://export.arxiv.org/api/query?search_query=all:" + url.QueryEscape(query) +
		"&start=0&max_results=" + fmt.Sprintf("%d", topN) + "&sortBy=" + url.QueryEscape(sortParam)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		a.fail(err.Error())
		return err
	}
	req.Header.Set("User-Agent", "LightningRAG-Agent/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.fail(err.Error())
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		a.fail(err.Error())
		return err
	}

	var feed struct {
		Entry []struct {
			Title   string `xml:"http://www.w3.org/2005/Atom title"`
			ID      string `xml:"http://www.w3.org/2005/Atom id"`
			Summary string `xml:"http://www.w3.org/2005/Atom summary"`
		} `xml:"http://www.w3.org/2005/Atom entry"`
	}
	if err := xml.Unmarshal(body, &feed); err != nil {
		a.fail(err.Error())
		return err
	}

	var snips []webSnippet
	var jsonRows []any
	for _, e := range feed.Entry {
		title := strings.TrimSpace(e.Title)
		title = strings.Join(strings.Fields(title), " ")
		sum := strings.TrimSpace(e.Summary)
		sum = strings.Join(strings.Fields(sum), " ")
		id := strings.TrimSpace(e.ID)
		if title == "" && sum == "" {
			continue
		}
		snips = append(snips, webSnippet{Title: title, URL: id, Body: sum})
		jsonRows = append(jsonRows, map[string]any{"title": title, "url": id, "body": sum})
	}

	formalized := formatWebSnippets(snips)
	a.mu.Lock()
	a.output["formalized_content"] = formalized
	a.output["json"] = jsonRows
	a.mu.Unlock()
	return nil
}

func (a *ArXiv) fail(msg string) {
	a.mu.Lock()
	a.err = msg
	a.mu.Unlock()
}

func (a *ArXiv) Output(key string) any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.output[key]
}

func (a *ArXiv) OutputAll() map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range a.output {
		out[k] = v
	}
	return out
}

func (a *ArXiv) SetOutput(key string, value any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output[key] = value
}

func (a *ArXiv) Error() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.err
}

func (a *ArXiv) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output = make(map[string]any)
	a.err = ""
}
