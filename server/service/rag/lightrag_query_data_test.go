package rag

import "testing"

func TestRetrieverSessionFromLightningRAGParams(t *testing.T) {
	tk := 20
	s := RetrieverSessionFromLightningRAGParams("mix", &tk, nil, nil, nil)
	if s.ModeOverride != "mix" || s.RetrievePoolTopK == nil || *s.RetrievePoolTopK != 20 {
		t.Fatalf("session: %#v", s)
	}
	s2 := RetrieverSessionFromLightningRAGParams("  INVALID  ", nil, nil, nil, nil)
	if s2.ModeOverride != "" {
		t.Fatalf("invalid mode: %#v", s2)
	}
	s3 := RetrieverSessionFromLightningRAGParams("naive", nil, nil, nil, nil)
	if s3.ModeOverride != "vector" {
		t.Fatalf("naive should normalize to vector, got %q", s3.ModeOverride)
	}
}

func TestChunksMapsFromDocs(t *testing.T) {
	// smoke: empty
	if n := len(chunksMapsFromDocs(nil, true)); n != 0 {
		t.Fatalf("empty: %d", n)
	}
}
