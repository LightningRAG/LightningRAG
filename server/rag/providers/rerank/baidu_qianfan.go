package rerank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// BaiduQianfan 百度千帆 v2 Rerank（https://qianfan.baidubce.com/v2/rerank）
// API Key：可直接使用控制台「API Key」作为 Bearer；或使用 JSON {"yiyan_ak":"","yiyan_sk":""} 走 OAuth 换 token（与上游参考 BaiduYiyanRerank 字段兼容）
type BaiduQianfan struct {
	baseURL  string
	model    string
	client   *http.Client
	mu       sync.Mutex
	bearer   string
	oauthAK  string
	oauthSK  string
	tokenExp time.Time
	useOAuth bool
}

// NewBaiduQianfan baseURL 默认 https://qianfan.baidubce.com/v2/rerank；model 默认 bce-reranker-base
func NewBaiduQianfan(apiKey, baseURL, model string) (*BaiduQianfan, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("baiduyiyan rerank: 需要 API Key 或 AK/SK JSON")
	}
	if baseURL == "" {
		baseURL = "https://qianfan.baidubce.com/v2/rerank"
	}
	if model == "" {
		model = "bce-reranker-base"
	}
	b := &BaiduQianfan{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
	if err := b.initCredentials(apiKey); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *BaiduQianfan) initCredentials(apiKey string) error {
	if apiKey[0] == '{' || strings.Contains(apiKey, "yiyan_ak") {
		var m map[string]any
		if err := json.Unmarshal([]byte(apiKey), &m); err != nil {
			return fmt.Errorf("baiduyiyan rerank: 解析密钥 JSON: %w", err)
		}
		if v, ok := m["api_key"].(string); ok && strings.TrimSpace(v) != "" {
			b.bearer = strings.TrimPrefix(strings.TrimSpace(v), "Bearer ")
			return nil
		}
		if v, ok := m["qianfan_api_key"].(string); ok && strings.TrimSpace(v) != "" {
			b.bearer = strings.TrimPrefix(strings.TrimSpace(v), "Bearer ")
			return nil
		}
		ak, _ := m["yiyan_ak"].(string)
		sk, _ := m["yiyan_sk"].(string)
		if ak != "" && sk != "" {
			b.oauthAK, b.oauthSK = ak, sk
			b.useOAuth = true
			return nil
		}
		return fmt.Errorf("baiduyiyan rerank: JSON 中需 api_key / qianfan_api_key 或 yiyan_ak+yiyan_sk")
	}
	b.bearer = strings.TrimPrefix(apiKey, "Bearer ")
	return nil
}

func (b *BaiduQianfan) effectiveBearer(ctx context.Context) (string, error) {
	if !b.useOAuth {
		return b.bearer, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if time.Now().Before(b.tokenExp.Add(-2*time.Minute)) && b.bearer != "" {
		return b.bearer, nil
	}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", b.oauthAK)
	form.Set("client_secret", b.oauthSK)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://aip.baidubce.com/oauth/2.0/token?"+form.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baidu oauth: HTTP %d %s", resp.StatusCode, string(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("baidu oauth: %s %s", tok.Error, tok.Description)
	}
	b.bearer = tok.AccessToken
	if tok.ExpiresIn > 0 {
		b.tokenExp = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	} else {
		b.tokenExp = time.Now().Add(25 * 24 * time.Hour)
	}
	return b.bearer, nil
}

type baiduRerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n,omitempty"`
}

type baiduRerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
	Score          float64 `json:"score"`
	Document       string  `json:"document"`
}

type baiduRerankResponse struct {
	Results []baiduRerankResult `json:"results"`
	Message string              `json:"message"`
}

func (b *BaiduQianfan) Rerank(ctx context.Context, query string, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	truncated := make([]string, len(texts))
	for i, t := range texts {
		truncated[i] = truncateCompat(t, maxDocLenCompat)
	}
	token, err := b.effectiveBearer(ctx)
	if err != nil {
		return nil, err
	}
	reqBody := baiduRerankRequest{
		Model:     b.model,
		Query:     query,
		Documents: truncated,
		TopN:      len(texts),
	}
	payload, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("baiduyiyan rerank: HTTP %d %s", resp.StatusCode, string(body))
	}
	var out baiduRerankResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("baiduyiyan rerank: 解析响应: %w", err)
	}
	rows := make([]scoreRow, len(out.Results))
	for i, r := range out.Results {
		score := r.RelevanceScore
		if score == 0 && r.Score != 0 {
			score = r.Score
		}
		rows[i] = scoreRow{Index: r.Index, Score: score}
	}
	return fillScoresForDocuments(len(texts), rows), nil
}

func (b *BaiduQianfan) ProviderName() string { return "baiduyiyan" }
func (b *BaiduQianfan) ModelName() string    { return b.model }
