package rag

import (
	"testing"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/songzhibin97/gkit/cache/local_cache"
)

func TestRetrieveCacheRoundTrip(t *testing.T) {
	prev := global.BlackCache
	t.Cleanup(func() { global.BlackCache = prev })
	global.BlackCache = local_cache.NewCache(local_cache.SetDefaultExpire(time.Hour))

	k := retrieveCacheKey(7, []uint{2, 1}, "hello", 5, 10, RetrieverSessionOptions{ModeOverride: "hybrid"})
	docs := []ragschema.Document{{PageContent: "x", Score: 0.9, Metadata: map[string]any{"a": float64(1)}}}
	saveRetrieveCache(k, docs, 120)
	got, ok := loadRetrieveCache(k)
	if !ok || len(got) != 1 || got[0].PageContent != "x" || got[0].Score != 0.9 {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
}
