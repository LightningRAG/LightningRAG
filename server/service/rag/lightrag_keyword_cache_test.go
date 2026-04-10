package rag

import (
	"testing"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/songzhibin97/gkit/cache/local_cache"
)

func TestKeywordExtractCacheKeyStable(t *testing.T) {
	k1 := keywordExtractCacheKey("  hello  ", " ctx ")
	k2 := keywordExtractCacheKey("hello", "ctx")
	if k1 != k2 {
		t.Fatalf("expected trim-stable key: %q vs %q", k1, k2)
	}
	if k1 == keywordExtractCacheKey("hello", "other") {
		t.Fatal("context should affect key")
	}
}

func TestKeywordExtractCacheRoundTrip(t *testing.T) {
	prev := global.BlackCache
	t.Cleanup(func() { global.BlackCache = prev })
	global.BlackCache = local_cache.NewCache(local_cache.SetDefaultExpire(time.Hour))

	key := keywordExtractCacheKey("q", "x")
	saveKeywordExtractCache(key, []string{"a"}, []string{"b", "c"}, 60)
	h, l, ok := loadKeywordExtractCache(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(h) != 1 || h[0] != "a" || len(l) != 2 {
		t.Fatalf("hl=%v ll=%v", h, l)
	}
}
