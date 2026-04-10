package rag

import (
	"testing"
)

func TestKgExtractLLMCacheKeyStable(t *testing.T) {
	b1 := kgExtractBatch{Items: []kgExtractBatchItem{{ChunkIndex: 0, Text: "a"}}}
	b2 := kgExtractBatch{Items: []kgExtractBatchItem{{ChunkIndex: 0, Text: "a"}}}
	k1 := kgExtractLLMCacheKey(1, kgExtractBatchUserPayload(b1))
	k2 := kgExtractLLMCacheKey(1, kgExtractBatchUserPayload(b2))
	if k1 != k2 {
		t.Fatalf("same payload should match: %q %q", k1, k2)
	}
	if k1 == kgExtractLLMCacheKey(2, kgExtractBatchUserPayload(b1)) {
		t.Fatal("kb id should affect key")
	}
}
