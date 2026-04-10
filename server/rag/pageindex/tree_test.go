package pageindex

import (
	"testing"
)

func TestBuildTreeFromMarkdown(t *testing.T) {
	md := `# 第一章

这是第一章内容。

## 1.1 小节

小节内容。

### 1.1.1 子节

子节内容。

## 1.2 另一小节

另一小节内容。
`
	tree := BuildTreeFromMarkdown(md)
	if len(tree) == 0 {
		t.Fatal("期望至少一个根节点")
	}
	if tree[0].Title != "第一章" {
		t.Errorf("根节点标题应为 第一章，got %s", tree[0].Title)
	}
	if len(tree[0].Nodes) < 2 {
		t.Errorf("期望至少 2 个子节点，got %d", len(tree[0].Nodes))
	}
	m := CreateNodeMapping(tree)
	if len(m) == 0 {
		t.Fatal("node mapping 不应为空")
	}
}

func TestBuildTreeForPageIndex(t *testing.T) {
	md := "# A\n\nbody\n"
	tree := BuildTreeForPageIndex("md", "x.md", md)
	if len(tree) == 0 || tree[0].Title != "A" {
		t.Fatalf("unexpected tree: %+v", tree)
	}
	tree2 := BuildTreeForPageIndex("txt", "n.txt", "hello")
	if len(tree2) != 1 || tree2[0].Text != "hello" {
		t.Fatalf("txt single node: %+v", tree2)
	}
}

func TestBuildTreeFromTextChunks(t *testing.T) {
	tree := BuildTreeFromTextChunks("doc", []string{"a", "b"})
	if len(tree) != 1 || len(tree[0].Nodes) != 2 {
		t.Fatalf("expected root+2 children, got %+v", tree)
	}
	m := CreateNodeMapping(tree)
	if len(m) < 3 {
		t.Fatalf("expected 3+ nodes, got %d", len(m))
	}
}

func TestCreateNodeMapping(t *testing.T) {
	tree := []TreeNode{
		{NodeID: "0001", Title: "A", Nodes: []TreeNode{
			{NodeID: "0002", Title: "B"},
			{NodeID: "0003", Title: "C"},
		}},
	}
	m := CreateNodeMapping(tree)
	if m["0001"] == nil || m["0002"] == nil || m["0003"] == nil {
		t.Errorf("CreateNodeMapping 应包含所有节点，got %v", m)
	}
}
