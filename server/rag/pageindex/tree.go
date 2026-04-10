package pageindex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CreateNodeMapping 递归创建 node_id -> 节点 映射
func CreateNodeMapping(tree []TreeNode) map[string]*TreeNode {
	m := make(map[string]*TreeNode)
	for i := range tree {
		collectNodes(&tree[i], m)
	}
	return m
}

func collectNodes(n *TreeNode, m map[string]*TreeNode) {
	if n.NodeID != "" {
		m[n.NodeID] = n
	}
	for i := range n.Nodes {
		collectNodes(&n.Nodes[i], m)
	}
}

// GetNodeText 获取节点文本，优先用 Text，否则用 Summary
func GetNodeText(n *TreeNode) string {
	if n.Text != "" {
		return n.Text
	}
	if n.Summary != "" {
		return n.Summary
	}
	if n.PrefixSummary != "" {
		return n.PrefixSummary
	}
	return n.Title
}

// RemoveFields 移除指定字段，用于树检索时减小 payload
func RemoveFields(tree []TreeNode, fields map[string]bool) []TreeNode {
	out := make([]TreeNode, len(tree))
	for i := range tree {
		out[i] = removeFieldsFromNode(tree[i], fields)
	}
	return out
}

func removeFieldsFromNode(n TreeNode, fields map[string]bool) TreeNode {
	out := n
	if fields["text"] && out.Text != "" {
		out.Text = ""
	}
	if len(out.Nodes) > 0 {
		out.Nodes = make([]TreeNode, len(n.Nodes))
		for i := range n.Nodes {
			out.Nodes[i] = removeFieldsFromNode(n.Nodes[i], fields)
		}
	}
	return out
}

// StructureToList 将树展平为节点列表
func StructureToList(tree []TreeNode) []*TreeNode {
	var list []*TreeNode
	for i := range tree {
		flatten(&tree[i], &list)
	}
	return list
}

func flatten(n *TreeNode, list *[]*TreeNode) {
	*list = append(*list, n)
	for i := range n.Nodes {
		flatten(&n.Nodes[i], list)
	}
}

// WriteNodeID 递归为节点分配 node_id
func WriteNodeID(tree []TreeNode, startID int) int {
	for i := range tree {
		tree[i].NodeID = strconv.FormatInt(int64(startID), 10)
		if len(tree[i].NodeID) < 4 {
			tree[i].NodeID = padLeft(tree[i].NodeID, 4, '0')
		}
		startID++
		if len(tree[i].Nodes) > 0 {
			startID = WriteNodeID(tree[i].Nodes, startID)
		}
	}
	return startID
}

func padLeft(s string, n int, pad rune) string {
	for len(s) < n {
		s = string(pad) + s
	}
	return s
}

// headerPattern Markdown 标题正则
var headerPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// ExtractNodesFromMarkdown 从 Markdown 提取标题节点
func ExtractNodesFromMarkdown(content string) ([]mdNode, []string) {
	lines := splitLines(content)
	var nodes []mdNode
	inCodeBlock := false
	codeBlockPattern := regexp.MustCompile(`^` + "`" + "`" + "`")

	for i, line := range lines {
		stripped := trimSpace(line)
		if codeBlockPattern.MatchString(stripped) {
			inCodeBlock = !inCodeBlock
			continue
		}
		if stripped == "" {
			continue
		}
		if !inCodeBlock {
			if m := headerPattern.FindStringSubmatch(stripped); len(m) >= 3 {
				nodes = append(nodes, mdNode{
					Title:   m[2],
					LineNum: i + 1,
					Level:   len(m[1]),
				})
			}
		}
	}
	return nodes, lines
}

type mdNode struct {
	Title   string
	LineNum int
	Level   int
}

// ExtractNodeTextContent 为每个节点提取对应文本内容
func ExtractNodeTextContent(nodes []mdNode, lines []string) []mdNodeWithText {
	result := make([]mdNodeWithText, len(nodes))
	for i, n := range nodes {
		start := n.LineNum - 1
		end := len(lines)
		if i+1 < len(nodes) {
			end = nodes[i+1].LineNum - 1
		}
		text := ""
		for j := start; j < end && j < len(lines); j++ {
			if j > start {
				text += "\n"
			}
			text += lines[j]
		}
		result[i] = mdNodeWithText{
			mdNode: n,
			Text:   trimSpace(text),
		}
	}
	return result
}

type mdNodeWithText struct {
	mdNode
	Text string
}

// BuildTreeFromMarkdown 从 Markdown 构建树结构
func BuildTreeFromMarkdown(content string) []TreeNode {
	nodes, lines := ExtractNodesFromMarkdown(content)
	if len(nodes) == 0 {
		return nil
	}
	withText := ExtractNodeTextContent(nodes, lines)
	return buildTreeFromNodes(withText, 1)
}

func buildTreeFromNodes(nodes []mdNodeWithText, idStart int) []TreeNode {
	if len(nodes) == 0 {
		return nil
	}
	tree := buildTreeRecursive(nodes, 0, len(nodes), 0)
	WriteNodeID(tree, idStart)
	return tree
}

// buildTreeRecursive 递归构建树，[start,end) 为当前层级范围
func buildTreeRecursive(nodes []mdNodeWithText, start, end, level int) []TreeNode {
	var result []TreeNode
	i := start
	for i < end {
		n := nodes[i]
		if i > start && n.Level <= level {
			break
		}
		childStart := i + 1
		childEnd := childStart
		for childEnd < end && nodes[childEnd].Level > n.Level {
			childEnd++
		}
		tn := TreeNode{
			Title:   n.Title,
			Text:    n.Text,
			LineNum: n.LineNum,
		}
		if childEnd > childStart {
			tn.Nodes = buildTreeRecursive(nodes, childStart, childEnd, n.Level)
		}
		result = append(result, tn)
		i = childEnd
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	return lines
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// BuildTreeForPageIndex 与 references/PageIndex 的 md_to_tree 主流程一致：Markdown 按标题建树，否则整篇作为单节点（txt/无标题 md/pdf 文本等）
func BuildTreeForPageIndex(fileType, docName, content string) []TreeNode {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	ft := strings.ToLower(strings.TrimSpace(fileType))
	var tree []TreeNode
	if ft == "md" || ft == "mdx" {
		tree = BuildTreeFromMarkdown(content)
	}
	if len(tree) == 0 {
		tree = []TreeNode{{
			Title: docName,
			Text:  content,
			Nodes: nil,
		}}
		WriteNodeID(tree, 1)
	}
	return tree
}

// BuildTreeFromTextChunks 将已有切片（如向量阶段产生的 chunk）合成浅层树，供 PageIndex 推理检索；多片段时根节点为文档名。
func BuildTreeFromTextChunks(docName string, chunks []string) []TreeNode {
	var parts []string
	for _, c := range chunks {
		if t := strings.TrimSpace(c); t != "" {
			parts = append(parts, t)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		t := []TreeNode{{Title: docName, Text: parts[0]}}
		WriteNodeID(t, 1)
		return t
	}
	children := make([]TreeNode, 0, len(parts))
	for i, c := range parts {
		children = append(children, TreeNode{
			Title: fmt.Sprintf("Part %d", i+1),
			Text:  c,
		})
	}
	root := []TreeNode{{Title: docName, Nodes: children}}
	WriteNodeID(root, 1)
	return root
}
