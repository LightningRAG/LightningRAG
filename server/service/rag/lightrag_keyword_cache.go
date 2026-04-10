package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// 对齐 LightRAG llm_response_cache 思想：对「仅依赖问句+上下文摘录」的抽词结果做短期缓存，降低重复查询的模型调用。

type keywordExtractCachePayload struct {
	HL []string `json:"h"`
	LL []string `json:"l"`
}

func keywordExtractCacheKey(query, extraContext string) string {
	q := strings.TrimSpace(query)
	x := strings.TrimSpace(extraContext)
	sum := sha256.Sum256([]byte(q + "\x00" + x))
	return "lrag_kw:" + hex.EncodeToString(sum[:])
}

func loadKeywordExtractCache(key string) (hl, ll []string, ok bool) {
	raw, exists := global.BlackCache.Get(key)
	if !exists || raw == nil {
		return nil, nil, false
	}
	s, okStr := raw.(string)
	if !okStr || s == "" {
		return nil, nil, false
	}
	var p keywordExtractCachePayload
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		return nil, nil, false
	}
	return cloneKeywordSlice(p.HL), cloneKeywordSlice(p.LL), true
}

func saveKeywordExtractCache(key string, hl, ll []string, ttlSeconds int) {
	if ttlSeconds <= 0 {
		return
	}
	p := keywordExtractCachePayload{HL: hl, LL: ll}
	b, err := json.Marshal(p)
	if err != nil {
		return
	}
	global.BlackCache.Set(key, string(b), time.Duration(ttlSeconds)*time.Second)
}
