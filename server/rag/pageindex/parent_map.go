package pageindex

// BuildParentNormIDByChild 为树上每个非根节点记录其**父节点**的规范化 node_id（根节点不出现在 map 中）。
// 用于对齐 references/ragflow 中 chunk 的 mom_id 语义（子块指向父块）。
func BuildParentNormIDByChild(tree []TreeNode) map[string]string {
	out := make(map[string]string)
	var walk func(nodes []TreeNode, parentNorm string)
	walk = func(nodes []TreeNode, parentNorm string) {
		for i := range nodes {
			nid := NormalizeNodeID(nodes[i].NodeID)
			if parentNorm != "" {
				out[nid] = parentNorm
			}
			if len(nodes[i].Nodes) > 0 {
				walk(nodes[i].Nodes, nid)
			}
		}
	}
	walk(tree, "")
	return out
}
