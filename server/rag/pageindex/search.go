package pageindex

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// TreeSearch 使用 LLM 对树进行推理检索，返回相关节点 ID 列表
func TreeSearch(ctx context.Context, llm interfaces.LLM, query string, tree []TreeNode, maxNodes int) (*TreeSearchResult, error) {
	treeForSearch := RemoveFields(tree, map[string]bool{"text": true})
	treeJSON, err := json.MarshalIndent(treeForSearch, "", "  ")
	if err != nil {
		return nil, err
	}

	// 与 references/PageIndex/tutorials/tree-search 中的 LLM 树检索提示对齐
	prompt := fmt.Sprintf(`You are given a question and the tree structure of a document.
Each node has node_id, title, and summary (or prefix_summary). The full text is not shown in the tree.
Find every node that is likely to contain the answer.

Question: %s

Document tree structure:
%s

Reply with ONLY valid JSON in exactly this shape (no other text):
{
    "thinking": "<brief reasoning: which nodes are relevant?>",
    "node_list": ["node_id_1", "node_id_2", ...]
}

Return at most %d most relevant node_id values.`,
		query, string(treeJSON), maxNodes)

	msg := interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: prompt}},
	}
	resp, err := llm.GenerateContent(ctx, []interfaces.MessageContent{msg})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Content == "" {
		return &TreeSearchResult{NodeList: nil}, nil
	}

	parsed := extractJSON(resp.Choices[0].Content)
	var result TreeSearchResult
	if err := json.Unmarshal(parsed, &result); err != nil {
		return nil, fmt.Errorf("解析 LLM 树检索结果失败: %w", err)
	}
	return &result, nil
}

// extractJSON 从 LLM 输出中提取 JSON
func extractJSON(content string) []byte {
	content = strings.TrimSpace(content)
	// 尝试 ```json ... ``` 包裹
	if idx := strings.Index(content, "```json"); idx >= 0 {
		start := idx + 7
		if end := strings.Index(content[start:], "```"); end >= 0 {
			return []byte(strings.TrimSpace(content[start : start+end]))
		}
	}
	if idx := strings.Index(content, "```"); idx >= 0 {
		start := idx + 3
		if end := strings.Index(content[start:], "```"); end >= 0 {
			return []byte(strings.TrimSpace(content[start : start+end]))
		}
	}
	// 尝试找 { ... }
	start := strings.Index(content, "{")
	if start < 0 {
		return []byte(content)
	}
	depth := 0
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return []byte(content[start : i+1])
			}
		}
	}
	return []byte(content)
}

// 兼容旧版可能返回的 node_id 格式
var nodeIDClean = regexp.MustCompile(`^[0-9]+$`)

// NormalizeNodeID 确保 node_id 为 4 位格式
func NormalizeNodeID(s string) string {
	s = strings.TrimSpace(s)
	if nodeIDClean.MatchString(s) && len(s) < 4 {
		return padLeft(s, 4, '0')
	}
	return s
}
