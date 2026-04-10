package rag

import (
	"context"
	"testing"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/songzhibin97/gkit/cache/local_cache"
)

func TestKgEntityPresenceCachePositiveOnly(t *testing.T) {
	prev := global.BlackCache
	prevTTL := global.LRAG_CONFIG.Rag.KgEntityPresenceCacheTTLSeconds
	t.Cleanup(func() {
		global.BlackCache = prev
		global.LRAG_CONFIG.Rag.KgEntityPresenceCacheTTLSeconds = prevTTL
	})
	global.BlackCache = local_cache.NewCache(local_cache.SetDefaultExpire(time.Hour))
	global.LRAG_CONFIG.Rag.KgEntityPresenceCacheTTLSeconds = 60

	global.BlackCache.Set(kgEntityPresenceCacheKey(99), "1", time.Minute)
	if !KnowledgeGraphHasEntities(context.Background(), 99) {
		t.Fatal("expected cache hit true without DB")
	}
	InvalidateKgEntityPresenceCache(99)
	if _, ok := global.BlackCache.Get(kgEntityPresenceCacheKey(99)); ok {
		t.Fatal("invalidate should drop key")
	}
}
