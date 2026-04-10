package rag

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func TestNormalizeRetrieverBucketScoresForMerge(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		normalizeRetrieverBucketScoresForMerge(nil)
	})
	t.Run("single", func(t *testing.T) {
		d := []schema.Document{{Score: 0.3, PageContent: "a"}}
		normalizeRetrieverBucketScoresForMerge(d)
		if d[0].Score != 1 {
			t.Fatalf("got %v", d[0].Score)
		}
	})
	t.Run("min_max", func(t *testing.T) {
		d := []schema.Document{
			{Score: 10, PageContent: "a"},
			{Score: 20, PageContent: "b"},
			{Score: 15, PageContent: "c"},
		}
		normalizeRetrieverBucketScoresForMerge(d)
		if d[0].Score != 0 || d[1].Score != 1 || d[2].Score != 0.5 {
			t.Fatalf("got %#v %#v %#v", d[0].Score, d[1].Score, d[2].Score)
		}
	})
	t.Run("flat", func(t *testing.T) {
		d := []schema.Document{{Score: 0.5}, {Score: 0.5}}
		normalizeRetrieverBucketScoresForMerge(d)
		if d[0].Score != 1 || d[1].Score != 1 {
			t.Fatalf("got %#v %#v", d[0].Score, d[1].Score)
		}
	})
}
