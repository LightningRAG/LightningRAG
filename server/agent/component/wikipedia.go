package component

import (
	"context"
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

var wikiLangRe = regexp.MustCompile(`^[a-z][a-z0-9-]{1,14}$`)

func sanitizeWikiLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if wikiLangRe.MatchString(lang) {
		return lang
	}
	return "zh"
}

func init() {
	Register("Wikipedia", NewWikipedia)
}

// Wikipedia 使用 MediaWiki API 检索维基百科摘要（参考 references 目录内 agent/tools/wikipedia.py）
type Wikipedia struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewWikipedia 创建 Wikipedia 组件
func NewWikipedia(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Wikipedia{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (w *Wikipedia) ComponentName() string { return "Wikipedia" }

func (w *Wikipedia) Invoke(map[string]any) error {
	w.mu.Lock()
	w.err = ""
	w.mu.Unlock()

	q := getStrParam(w.params, "query")
	if q == "" {
		q = "sys.query"
	}
	query := resolveWorkflowQuery(w.canvas, q)
	if query == "" {
		w.mu.Lock()
		w.output["formalized_content"] = ""
		w.output["json"] = []any{}
		w.mu.Unlock()
		return nil
	}

	lang := getStrParam(w.params, "language")
	if lang == "" {
		lang = "zh"
	}
	topN := getIntParam(w.params, "top_n", 5)
	if topN < 1 {
		topN = 1
	}
	if topN > 15 {
		topN = 15
	}

	timeoutSec := getIntParam(w.params, "timeout", 30)
	if timeoutSec < 5 {
		timeoutSec = 5
	}
	if timeoutSec > 90 {
		timeoutSec = 90
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	lang = sanitizeWikiLang(lang)
	base := fmt.Sprintf("https://%s.wikipedia.org/w/api.php", lang)
	searchURL := base + "?action=query&list=search&format=json&utf8=1&srlimit=" +
		fmt.Sprintf("%d", topN) + "&srsearch=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		w.fail(err.Error())
		return err
	}
	req.Header.Set("User-Agent", "LightningRAG-Agent/1.0 (https://github.com/LightningRAG/LightningRAG)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.fail(err.Error())
		return err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if err != nil {
		w.fail(err.Error())
		return err
	}

	var searchRoot struct {
		Query *struct {
			Search []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}
	if err := json.Unmarshal(body, &searchRoot); err != nil {
		w.fail(err.Error())
		return err
	}
	if searchRoot.Query == nil || len(searchRoot.Query.Search) == 0 {
		w.mu.Lock()
		w.output["formalized_content"] = ""
		w.output["json"] = []any{}
		w.mu.Unlock()
		return nil
	}

	titles := make([]string, 0, len(searchRoot.Query.Search))
	for _, s := range searchRoot.Query.Search {
		if strings.TrimSpace(s.Title) != "" {
			titles = append(titles, s.Title)
		}
	}
	if len(titles) == 0 {
		w.mu.Lock()
		w.output["formalized_content"] = ""
		w.output["json"] = []any{}
		w.mu.Unlock()
		return nil
	}
	if len(titles) > 8 {
		titles = titles[:8]
	}

	extractURL := base + "?action=query&prop=extracts&exintro&explaintext&format=json&utf8=1&titles=" +
		url.QueryEscape(strings.Join(titles, "|"))

	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, extractURL, nil)
	if err != nil {
		w.fail(err.Error())
		return err
	}
	req2.Header.Set("User-Agent", req.Header.Get("User-Agent"))

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		w.fail(err.Error())
		return err
	}
	body2, err := io.ReadAll(io.LimitReader(resp2.Body, 2<<20))
	resp2.Body.Close()
	if err != nil {
		w.fail(err.Error())
		return err
	}

	var extractRoot struct {
		Query *struct {
			Pages map[string]struct {
				Title   string `json:"title"`
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(body2, &extractRoot); err != nil {
		w.fail(err.Error())
		return err
	}

	var snips []webSnippet
	var jsonRows []any
	if extractRoot.Query != nil {
		for _, p := range extractRoot.Query.Pages {
			title := strings.TrimSpace(p.Title)
			ex := strings.TrimSpace(p.Extract)
			if title == "" && ex == "" {
				continue
			}
			pageURL := fmt.Sprintf("https://%s.wikipedia.org/wiki/%s", lang, url.PathEscape(strings.ReplaceAll(title, " ", "_")))
			snips = append(snips, webSnippet{Title: title, URL: pageURL, Body: ex})
			jsonRows = append(jsonRows, map[string]any{"title": title, "url": pageURL, "body": ex})
		}
	}

	formalized := formatWebSnippets(snips)
	w.mu.Lock()
	w.output["formalized_content"] = formalized
	w.output["json"] = jsonRows
	w.mu.Unlock()
	return nil
}

func (w *Wikipedia) fail(msg string) {
	w.mu.Lock()
	w.err = msg
	w.mu.Unlock()
}

func (w *Wikipedia) Output(key string) any {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.output[key]
}

func (w *Wikipedia) OutputAll() map[string]any {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range w.output {
		out[k] = v
	}
	return out
}

func (w *Wikipedia) SetOutput(key string, value any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.output[key] = value
}

func (w *Wikipedia) Error() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.err
}

func (w *Wikipedia) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.output = make(map[string]any)
	w.err = ""
}
