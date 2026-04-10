package rag

import (
	"context"
	"strings"
	"testing"
)

func TestPrepareLightningRAGSearchQueriesManualKeywords(t *testing.T) {
	combined, kgEnt, kgRel, hl, ll := PrepareLightningRAGSearchQueries(context.Background(), 0, nil, "What did Alice do?",
		[]string{"team dynamics"}, []string{"Alice", "Bob"}, "")
	if combined == "" {
		t.Fatal("empty combined")
	}
	if kgEnt == "" || kgRel == "" {
		t.Fatalf("kgEnt=%q kgRel=%q", kgEnt, kgRel)
	}
	if len(hl) != 1 || hl[0] != "team dynamics" {
		t.Fatalf("resolved hl: %#v", hl)
	}
	if len(ll) != 2 || ll[0] != "Alice" || ll[1] != "Bob" {
		t.Fatalf("resolved ll: %#v", ll)
	}
	if !strings.Contains(kgEnt, "Alice") {
		t.Fatalf("entity query should mention low-level term: %q", kgEnt)
	}
	if !strings.Contains(kgRel, "team") {
		t.Fatalf("rel query should mention high-level term: %q", kgRel)
	}
}
