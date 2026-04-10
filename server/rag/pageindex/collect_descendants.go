package pageindex

// CollectDescendantNodeIDs 返回以 rootNorm 为根的子树中所有节点的规范化 node_id（前序遍历，含根）。
// 对齐 references/ragflow 中 TOC 条目的 ids：一条目录对应其下直至下一同级标题前的多个 chunk。
// 树形 PageIndex 下等价于「该标题节点及其全部后代」对应的切片集合。
// 若树中找不到 rootNorm，返回 nil。
func CollectDescendantNodeIDs(tree []TreeNode, rootNorm string) []string {
	want := NormalizeNodeID(rootNorm)
	var out []string
	var walk func(nodes []TreeNode) bool
	walk = func(nodes []TreeNode) bool {
		for i := range nodes {
			nid := NormalizeNodeID(nodes[i].NodeID)
			if nid == want {
				collectPreorderNode(&nodes[i], &out)
				return true
			}
			if walk(nodes[i].Nodes) {
				return true
			}
		}
		return false
	}
	if walk(tree) {
		return out
	}
	return nil
}

func collectPreorderNode(n *TreeNode, out *[]string) {
	if n == nil {
		return
	}
	*out = append(*out, NormalizeNodeID(n.NodeID))
	for i := range n.Nodes {
		collectPreorderNode(&n.Nodes[i], out)
	}
}
