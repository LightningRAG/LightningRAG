package rag

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
)

func TestEnrichMetadataRankBoostForChunk_position(t *testing.T) {
	prevP := global.LRAG_CONFIG.Rag.ChunkRankBoostByPosition
	prevF := global.LRAG_CONFIG.Rag.ChunkRankBoostPositionFloor
	prevD := global.LRAG_CONFIG.Rag.DefaultChunkRankBoost
	t.Cleanup(func() {
		global.LRAG_CONFIG.Rag.ChunkRankBoostByPosition = prevP
		global.LRAG_CONFIG.Rag.ChunkRankBoostPositionFloor = prevF
		global.LRAG_CONFIG.Rag.DefaultChunkRankBoost = prevD
	})
	global.LRAG_CONFIG.Rag.ChunkRankBoostByPosition = true
	global.LRAG_CONFIG.Rag.ChunkRankBoostPositionFloor = 0.5
	global.LRAG_CONFIG.Rag.DefaultChunkRankBoost = 0.99

	meta := map[string]any{"x": 1}
	enrichMetadataRankBoostForChunk(meta, 0, 3)
	if rb, ok := meta["rank_boost"].(float64); !ok || rb < 0.999 {
		t.Fatalf("first chunk want 1 got %v", meta["rank_boost"])
	}

	meta2 := map[string]any{}
	enrichMetadataRankBoostForChunk(meta2, 2, 3)
	if rb, ok := meta2["rank_boost"].(float64); !ok || rb < 0.49 || rb > 0.51 {
		t.Fatalf("last chunk want ~0.5 got %v", meta2["rank_boost"])
	}
}

func TestEnrichMetadataRankBoostForChunk_defaultConstant(t *testing.T) {
	prevP := global.LRAG_CONFIG.Rag.ChunkRankBoostByPosition
	prevD := global.LRAG_CONFIG.Rag.DefaultChunkRankBoost
	t.Cleanup(func() {
		global.LRAG_CONFIG.Rag.ChunkRankBoostByPosition = prevP
		global.LRAG_CONFIG.Rag.DefaultChunkRankBoost = prevD
	})
	global.LRAG_CONFIG.Rag.ChunkRankBoostByPosition = false
	global.LRAG_CONFIG.Rag.DefaultChunkRankBoost = 0.4
	meta := map[string]any{}
	enrichMetadataRankBoostForChunk(meta, 5, 100)
	if meta["rank_boost"] != 0.4 {
		t.Fatalf("%v", meta["rank_boost"])
	}
}

func TestApplyDocumentPriorityFloorToRankBoost(t *testing.T) {
	meta := map[string]any{"rank_boost": 0.2}
	applyDocumentPriorityFloorToRankBoost(meta, 0.5)
	if meta["rank_boost"] != 0.5 {
		t.Fatalf("want 0.5 got %v", meta["rank_boost"])
	}
	meta2 := map[string]any{"rank_boost": 0.8}
	applyDocumentPriorityFloorToRankBoost(meta2, 0.3)
	if meta2["rank_boost"] != 0.8 {
		t.Fatalf("want unchanged 0.8 got %v", meta2["rank_boost"])
	}
	applyDocumentPriorityFloorToRankBoost(nil, 1)
}
