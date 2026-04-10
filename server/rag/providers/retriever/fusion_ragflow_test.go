package retriever

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func TestMergeRagflowWeightedFusion_sameChunkBothLegs(t *testing.T) {
	meta := map[string]any{"rag_chunk_id": uint64(1)}
	vec := []schema.Document{{PageContent: "hello", Score: 0.9, Metadata: meta}}
	kw := []schema.Document{{PageContent: "hello", Score: 0.1, Metadata: meta}}
	out := mergeRagflowWeightedFusion(vec, kw, 5, 0.05, 0.95, 0)
	if len(out) != 1 {
		t.Fatalf("len %d", len(out))
	}
	// 两路各自 min-max 后为 1，融合 0.95+0.05=1
	if out[0].Score < 0.999 || out[0].Score > 1.001 {
		t.Fatalf("score %v", out[0].Score)
	}
}

func TestMergeRagflowWeightedFusion_ordersByFused(t *testing.T) {
	vec := []schema.Document{
		{PageContent: "a", Score: 0.5, Metadata: map[string]any{"rag_chunk_id": uint64(1)}},
		{PageContent: "b", Score: 1.0, Metadata: map[string]any{"rag_chunk_id": uint64(2)}},
	}
	kw := []schema.Document{
		{PageContent: "a", Score: 1.0, Metadata: map[string]any{"rag_chunk_id": uint64(1)}},
		{PageContent: "b", Score: 0.0, Metadata: map[string]any{"rag_chunk_id": uint64(2)}},
	}
	out := mergeRagflowWeightedFusion(vec, kw, 2, 0.05, 0.95, 0)
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
	// chunk1: vecN=0,kwN=1 -> 0.05; chunk2: vecN=1,kwN=0 -> 0.95
	if out[0].PageContent != "b" || out[1].PageContent != "a" {
		t.Fatalf("order: %#v", out)
	}
}

func TestMergeRagflowWeightedFusion_minFused(t *testing.T) {
	meta := map[string]any{"rag_chunk_id": uint64(1)}
	vec := []schema.Document{{PageContent: "x", Score: 0.1, Metadata: meta}}
	kw := []schema.Document{{PageContent: "x", Score: 0.1, Metadata: meta}}
	// 单条一路时 min-max 全为 1，融合分为 1.0；阈值高于 1 则全部过滤
	out := mergeRagflowWeightedFusion(vec, kw, 5, 0.05, 0.95, 1.01)
	if len(out) != 0 {
		t.Fatalf("expected filter out, got %d", len(out))
	}
}

func TestMinMaxNormalizeScores(t *testing.T) {
	docs := []schema.Document{{Score: 10}, {Score: 20}, {Score: 15}}
	n := minMaxNormalizeScores(docs)
	if len(n) != 3 || n[0] != 0 || n[1] != 1 || n[2] != 0.5 {
		t.Fatalf("%v", n)
	}
}
