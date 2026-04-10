package retriever

import "testing"

func TestRagflowRerankCandidateFloor(t *testing.T) {
	if got := ragflowRerankCandidateFloor(1); got != 30 {
		t.Fatalf("pageSize 1: got %d", got)
	}
	if got := ragflowRerankCandidateFloor(10); got != 70 {
		t.Fatalf("pageSize 10: want 70 got %d", got)
	}
	if got := ragflowRerankCandidateFloor(32); got != 64 {
		t.Fatalf("pageSize 32: want 64 got %d", got)
	}
}
