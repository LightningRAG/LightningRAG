// Package pageindex 实现 PageIndex 检索：目录树 + LLM。
// 索引与树结构参考 references/PageIndex；检索阶段主路径对齐 references/ragflow 的 TOC 相关性打分（toc_relevance），
// 失败时回退为 PageIndex 式树检索（node_list）。
package pageindex

// TreeNode PageIndex 树节点，对应 Python 版结构
type TreeNode struct {
	Title         string     `json:"title"`
	NodeID        string     `json:"node_id,omitempty"`
	Summary       string     `json:"summary,omitempty"`
	PrefixSummary string     `json:"prefix_summary,omitempty"`
	Text          string     `json:"text,omitempty"`
	StartIndex    int        `json:"start_index,omitempty"` // 起始页码
	EndIndex      int        `json:"end_index,omitempty"`   // 结束页码
	LineNum       int        `json:"line_num,omitempty"`    // Markdown 行号
	Nodes         []TreeNode `json:"nodes,omitempty"`
}

// TreeSearchResult 树检索结果，LLM 返回的 JSON 格式
type TreeSearchResult struct {
	Thinking string   `json:"thinking"`
	NodeList []string `json:"node_list"`
}
