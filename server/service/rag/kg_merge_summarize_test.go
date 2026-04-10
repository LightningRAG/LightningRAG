package rag

import "testing"

func TestKgDescriptionSegmentsForSummarize(t *testing.T) {
	if s := kgDescriptionSegmentsForSummarize(""); len(s) != 0 {
		t.Fatalf("empty: got %v", s)
	}
	if s := kgDescriptionSegmentsForSummarize("  a  \n\n b \n"); len(s) != 2 || s[0] != "a" || s[1] != "b" {
		t.Fatalf("got %v", s)
	}
}
