package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// 与 kg_extract_cache 一致：提示词迭代或摘要策略变更时升高版本号使旧缓存自然失效（v2：支持 map-reduce 多轮）
const kgMergeSummarizePromptVersion = "kgms-v2"

func kgMergeSummarizeCacheKey(objectKind, displayName string, segments []string) string {
	segJSON, err := json.Marshal(segments)
	if err != nil {
		segJSON = []byte("[]")
	}
	sum := sha256.Sum256([]byte(kgMergeSummarizePromptVersion + "\x00" + objectKind + "\x00" + displayName + "\x00" + string(segJSON)))
	return "lrag_kgms:" + hex.EncodeToString(sum[:])
}

func loadKgMergeSummarizeCache(key string) (string, bool) {
	v, ok := global.BlackCache.Get(key)
	if !ok || v == nil {
		return "", false
	}
	s, okStr := v.(string)
	if !okStr || s == "" {
		return "", false
	}
	return s, true
}

func saveKgMergeSummarizeCache(key, summary string, ttlSeconds int) {
	if ttlSeconds <= 0 || summary == "" {
		return
	}
	global.BlackCache.Set(key, summary, time.Duration(ttlSeconds)*time.Second)
}
