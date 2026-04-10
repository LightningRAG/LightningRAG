package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// 与 lightrag_keyword_cache 一致：用 BlackCache 做短期 KV；键内嵌版本号便于提示词迭代后自然失效。
const kgExtractLLMPromptVersion = "kgex-v2"

func kgExtractBatchUserPayload(batch kgExtractBatch) string {
	var sb string
	for _, it := range batch.Items {
		sb += fmt.Sprintf("--- chunk_index=%d ---\n%s\n\n", it.ChunkIndex, it.Text)
	}
	return sb
}

func kgExtractLLMCacheKey(kbID uint, userPayload string) string {
	h := sha256.Sum256([]byte(kgExtractLLMPromptVersion + "\x00" + fmtUint(kbID) + "\x00" + userPayload))
	return "lrag_kgx:" + hex.EncodeToString(h[:])
}

func loadKgExtractLLMCache(key string) (raw string, ok bool) {
	v, exists := global.BlackCache.Get(key)
	if !exists || v == nil {
		return "", false
	}
	s, okStr := v.(string)
	if !okStr || s == "" {
		return "", false
	}
	return s, true
}

func saveKgExtractLLMCache(key, raw string, ttlSeconds int) {
	if ttlSeconds <= 0 || raw == "" {
		return
	}
	global.BlackCache.Set(key, raw, time.Duration(ttlSeconds)*time.Second)
}
