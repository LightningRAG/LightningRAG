package rag

import (
	"testing"
)

func TestExposeReferencesForAPI(t *testing.T) {
	refs := []map[string]any{
		{"index": 0, "content": "hello", "score": float32(0.9)},
	}
	f := false
	tr := true

	if got := ExposeReferencesForAPI(refs, &f, nil); got != nil {
		t.Fatalf("includeReferences false: want nil, got %#v", got)
	}
	if got := ExposeReferencesForAPI(refs, nil, nil); len(got) != 1 || got[0]["content"] != "hello" {
		t.Fatalf("defaults: want full refs, got %#v", got)
	}
	if got := ExposeReferencesForAPI(refs, nil, &tr); len(got) != 1 || got[0]["content"] != "hello" {
		t.Fatalf("includeChunkContent true: want content, got %#v", got)
	}
	if got := ExposeReferencesForAPI(refs, nil, &f); len(got) != 1 {
		t.Fatalf("strip content: want len 1, got %#v", got)
	} else if _, ok := got[0]["content"]; ok {
		t.Fatalf("strip content: content key should be absent, got %#v", got[0])
	}
}
