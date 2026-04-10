package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
)

const retrieveCacheVersion = "ret-v2"

var (
	retrieveKBEpochMu sync.RWMutex
	retrieveKBEpoch   = map[uint]uint64{} // 每库数据变更代数，写入 cache key，避免索引/删文后仍命中旧检索结果
)

// BumpRetrieveCacheEpochForKnowledgeBase 文档索引完成或删除等变更向量/切片后调用，使该库相关检索缓存失效
func BumpRetrieveCacheEpochForKnowledgeBase(kbID uint) {
	if kbID == 0 {
		return
	}
	retrieveKBEpochMu.Lock()
	retrieveKBEpoch[kbID]++
	retrieveKBEpochMu.Unlock()
}

func retrieveKBEpochFingerprint(kbIDs []uint) string {
	ids := append([]uint(nil), kbIDs...)
	slices.Sort(ids)
	var b strings.Builder
	retrieveKBEpochMu.RLock()
	for _, id := range ids {
		b.WriteString(strconv.FormatUint(uint64(id), 10))
		b.WriteByte(':')
		b.WriteString(strconv.FormatUint(retrieveKBEpoch[id], 10))
		b.WriteByte('|')
	}
	retrieveKBEpochMu.RUnlock()
	return b.String()
}

type retrieveCachePayload struct {
	Docs []retrieveCacheDoc `json:"docs"`
}

type retrieveCacheDoc struct {
	PageContent string         `json:"c"`
	Score       float32        `json:"s"`
	Metadata    map[string]any `json:"m,omitempty"`
}

func retrieveCacheKey(uid uint, kbIDs []uint, question string, topN, candidateN int, session RetrieverSessionOptions) string {
	ids := append([]uint(nil), kbIDs...)
	slices.Sort(ids)
	rerank := "*"
	if session.RerankOverride != nil {
		if *session.RerankOverride {
			rerank = "t"
		} else {
			rerank = "f"
		}
	}
	pool := "*"
	if session.RetrievePoolTopK != nil {
		pool = strconv.Itoa(*session.RetrievePoolTopK)
	}
	ct := "*"
	if session.CosineThreshold != nil {
		ct = fmt.Sprintf("%g", *session.CosineThreshold)
	}
	mr := "*"
	if session.MinRerankScore != nil {
		mr = fmt.Sprintf("%g", *session.MinRerankScore)
	}
	tocEnh := "*"
	if session.PageIndexTocEnhance != nil {
		if *session.PageIndexTocEnhance {
			tocEnh = "t"
		} else {
			tocEnh = "f"
		}
	}
	payload := fmt.Sprintf("%s\x00%s\x00%d\x00%v\x00%s\x00%d\x00%d\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s",
		retrieveCacheVersion, retrieveKBEpochFingerprint(ids), uid, ids, question, topN, candidateN,
		session.ModeOverride, pool, rerank, ct, mr,
		session.KgEntitySearchQuery, session.KgRelSearchQuery, tocEnh)
	sum := sha256.Sum256([]byte(payload))
	return "lrag_ret:" + hex.EncodeToString(sum[:])
}

func loadRetrieveCache(key string) ([]ragschema.Document, bool) {
	v, ok := global.BlackCache.Get(key)
	if !ok || v == nil {
		return nil, false
	}
	s, okStr := v.(string)
	if !okStr || s == "" {
		return nil, false
	}
	var p retrieveCachePayload
	if err := json.Unmarshal([]byte(s), &p); err != nil || len(p.Docs) == 0 {
		return nil, false
	}
	out := make([]ragschema.Document, len(p.Docs))
	for i, d := range p.Docs {
		out[i] = ragschema.Document{
			PageContent: d.PageContent,
			Score:       d.Score,
			Metadata:    cloneMetadataMap(d.Metadata),
		}
	}
	return out, true
}

func saveRetrieveCache(key string, docs []ragschema.Document, ttlSeconds int) {
	if ttlSeconds <= 0 || len(docs) == 0 {
		return
	}
	p := retrieveCachePayload{Docs: make([]retrieveCacheDoc, len(docs))}
	for i, d := range docs {
		p.Docs[i] = retrieveCacheDoc{
			PageContent: d.PageContent,
			Score:       d.Score,
			Metadata:    cloneMetadataMap(d.Metadata),
		}
	}
	b, err := json.Marshal(p)
	if err != nil {
		return
	}
	const maxRetrieveCacheBytes = 2 << 20 // 2MiB，避免占满本地缓存
	if len(b) > maxRetrieveCacheBytes {
		return
	}
	global.BlackCache.Set(key, string(b), time.Duration(ttlSeconds)*time.Second)
}

func cloneMetadataMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneDocSlice(docs []ragschema.Document) []ragschema.Document {
	if len(docs) == 0 {
		return nil
	}
	out := make([]ragschema.Document, len(docs))
	for i := range docs {
		out[i] = ragschema.Document{
			PageContent: docs[i].PageContent,
			Score:       docs[i].Score,
			Metadata:    cloneMetadataMap(docs[i].Metadata),
		}
	}
	return out
}
