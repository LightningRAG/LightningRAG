package rag

import (
	"strings"
	"testing"

	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func TestTrimDocsToRagTokenBudget(t *testing.T) {
	docs := []ragschema.Document{
		{PageContent: strings.Repeat("a", 400)}, // ~100 tok
		{PageContent: strings.Repeat("b", 400)},
	}
	out := trimDocsToRagTokenBudget(docs, 0)
	if len(out) != 2 {
		t.Fatalf("no budget: want 2 got %d", len(out))
	}
	out = trimDocsToRagTokenBudget(docs, 150)
	if len(out) != 2 {
		t.Fatalf("budget fits two: want 2 got %d", len(out))
	}
	out = trimDocsToRagTokenBudget(docs, 80)
	if len(out) != 1 {
		t.Fatalf("budget one chunk: want 1 got %d", len(out))
	}
	if !strings.HasSuffix(out[0].PageContent, "[truncated]") {
		t.Fatalf("expected truncated suffix in single partial chunk")
	}
}

func TestTrimDocsToRagTokenBudgetPrefersHigherScore(t *testing.T) {
	low := ragschema.Document{PageContent: strings.Repeat("x", 400), Score: 0.1}
	high := ragschema.Document{PageContent: strings.Repeat("y", 400), Score: 0.9}
	out := trimDocsToRagTokenBudget([]ragschema.Document{low, high}, 100)
	if len(out) != 1 {
		t.Fatalf("want 1 chunk, got %d", len(out))
	}
	if !strings.HasPrefix(out[0].PageContent, "y") {
		t.Fatalf("expected higher-score chunk (y…), got %q…", out[0].PageContent[:1])
	}
}

func TestTruncateUTF8ByApproxTokens(t *testing.T) {
	s := "你好世界" // 12 bytes utf8
	if got := truncateUTF8ByApproxTokens(s, 100); got != s {
		t.Fatalf("small string unchanged: %q", got)
	}
}
