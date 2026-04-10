package rerank

import (
	"testing"
)

func TestFillScoresForDocuments_uniqueIndices(t *testing.T) {
	rows := []scoreRow{
		{Index: 1, Score: 0.9},
		{Index: 0, Score: 0.1},
	}
	s := fillScoresForDocuments(2, rows)
	if s[0] != 0.1 || s[1] != 0.9 {
		t.Fatalf("got %v", s)
	}
}

func TestFillScoresForDocuments_allIndexZeroSequentialFallback(t *testing.T) {
	rows := []scoreRow{
		{Index: 0, Score: 0.2},
		{Index: 0, Score: 0.8},
	}
	s := fillScoresForDocuments(2, rows)
	if s[0] != 0.2 || s[1] != 0.8 {
		t.Fatalf("got %v", s)
	}
}

func TestFillScoresForDocuments_singleUniqueNonZeroNoFallback(t *testing.T) {
	rows := []scoreRow{
		{Index: 1, Score: 0.5},
		{Index: 1, Score: 0.9},
	}
	s := fillScoresForDocuments(2, rows)
	if s[0] != 0 || s[1] != 0.9 {
		t.Fatalf("got %v", s)
	}
}
