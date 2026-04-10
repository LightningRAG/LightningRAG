package pageindex

import "testing"

func TestBuildParentNormIDByChild(t *testing.T) {
	tree := []TreeNode{
		{NodeID: "0001", Title: "A", Nodes: []TreeNode{
			{NodeID: "0002", Title: "B"},
		}},
	}
	m := BuildParentNormIDByChild(tree)
	if m["0002"] != "0001" {
		t.Fatalf("expected parent of 0002 to be 0001, got %q", m["0002"])
	}
	if _, ok := m["0001"]; ok {
		t.Fatal("root should not be in map")
	}
}
