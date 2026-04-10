package retriever

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func TestDeduplicateRetrievedDocuments_sameChunkIDKeepsHigherScore(t *testing.T) {
	docs := []schema.Document{
		{PageContent: "hello", Score: 0.5, Metadata: map[string]any{"rag_chunk_id": float64(1)}},
		{PageContent: "hello", Score: 0.9, Metadata: map[string]any{"rag_chunk_id": float64(1)}},
	}
	out := DeduplicateRetrievedDocuments(docs)
	if len(out) != 1 || out[0].Score != 0.9 {
		t.Fatalf("got %#v", out)
	}
}

func TestDeduplicateRetrievedDocuments_sameContentDifferentChunkIDKeepsBoth(t *testing.T) {
	docs := []schema.Document{
		{PageContent: "same body", Score: 0.4, Metadata: map[string]any{"rag_chunk_id": float64(1)}},
		{PageContent: "same body", Score: 0.8, Metadata: map[string]any{"rag_chunk_id": float64(2)}},
	}
	out := DeduplicateRetrievedDocuments(docs)
	if len(out) != 2 {
		t.Fatalf("expected 2 distinct chunks, got %d", len(out))
	}
}

func TestDeduplicateRetrievedDocuments_unkeyedNeverMergedWithEachOther(t *testing.T) {
	docs := []schema.Document{
		{PageContent: "a", Score: 0.5, Metadata: nil},
		{PageContent: "a", Score: 0.9, Metadata: nil},
	}
	out := DeduplicateRetrievedDocuments(docs)
	if len(out) != 2 {
		t.Fatalf("unkeyed docs must not merge by content, got len=%d", len(out))
	}
}

func TestMergeFusionDocuments_upgradesScoreOnDuplicateKey(t *testing.T) {
	vec := []schema.Document{
		{PageContent: "a", Score: 0.5, Metadata: map[string]any{"rag_chunk_id": float64(1)}},
	}
	kw := []schema.Document{
		{PageContent: "a", Score: 0.95, Metadata: map[string]any{"rag_chunk_id": float64(1)}},
	}
	out := MergeFusionDocuments(vec, kw, FusionHybrid, 4)
	if len(out) != 1 || out[0].Score != 0.95 {
		t.Fatalf("got %#v", out)
	}
}
