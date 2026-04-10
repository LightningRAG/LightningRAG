package pageindex

import (
	"reflect"
	"testing"
)

func TestCollectDescendantNodeIDs(t *testing.T) {
	tree := []TreeNode{
		{NodeID: "0001", Title: "A", Nodes: []TreeNode{
			{NodeID: "0002", Title: "B", Nodes: []TreeNode{
				{NodeID: "0003", Title: "C"},
			}},
		}},
	}
	got := CollectDescendantNodeIDs(tree, "0002")
	want := []string{"0002", "0003"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if CollectDescendantNodeIDs(tree, "9999") != nil {
		t.Fatal("expected nil for missing node")
	}
}
