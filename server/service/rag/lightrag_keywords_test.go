package rag

import "testing"

func TestAugmentQueryWithLightningRAGKeywords(t *testing.T) {
	got := AugmentQueryWithLightningRAGKeywords("hello", []string{"A"}, []string{"b", "c"})
	if want := "hello\nb c A"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if AugmentQueryWithLightningRAGKeywords("  x  ", nil, nil) != "x" {
		t.Fatal("trim")
	}
}
