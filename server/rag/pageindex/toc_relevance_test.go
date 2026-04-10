package pageindex

import (
	"encoding/json"
	"testing"
)

func TestFlattenTreeToTOC(t *testing.T) {
	tree := []TreeNode{
		{NodeID: "0001", Title: "Part A", Nodes: []TreeNode{
			{NodeID: "0002", Title: "Sub 1"},
			{NodeID: "0003", Title: "Sub 2"},
		}},
	}
	flat := FlattenTreeToTOC(tree)
	if len(flat) != 3 {
		t.Fatalf("len=%d want 3", len(flat))
	}
	if flat[0].Level != 1 || flat[0].Title != "Part A" || flat[0].NodeID != "0001" {
		t.Errorf("root: %+v", flat[0])
	}
	if flat[1].Level != 2 || flat[1].Title != "Sub 1" {
		t.Errorf("child1: %+v", flat[1])
	}
}

func TestCapTOCEntries(t *testing.T) {
	var entries []TOCFlattenEntry
	for i := 0; i < MaxTocEntriesForLLM+10; i++ {
		entries = append(entries, TOCFlattenEntry{Level: 1, Title: "x", NodeID: "0001"})
	}
	c := CapTOCEntries(entries)
	if len(c) != MaxTocEntriesForLLM {
		t.Fatalf("cap len=%d", len(c))
	}
}

func TestParseTOCScoreInt(t *testing.T) {
	v, ok := parseTOCScoreInt(json.RawMessage(`5`))
	if !ok || v != 5 {
		t.Fatalf("got %d ok=%v", v, ok)
	}
}
