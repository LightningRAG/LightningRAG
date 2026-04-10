package pageindex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// TOCFlattenEntry 将文档树压平为目录项，与 references/ragflow 中 TOC 的 level/title 一致，并保留 node_id 用于回查正文
type TOCFlattenEntry struct {
	Level  int
	Title  string
	NodeID string
}

// TOCScoredNode TOC 相关性打分后的节点（Score 为 0~1，与 ragflow relevant_chunks_with_toc 中 score/5 一致）
type TOCScoredNode struct {
	NodeID string
	Score  float64
}

// MaxTocEntriesForLLM 与 references/ragflow 中 run_toc_from_text 对超长目录的裁剪量级一致（合并后仍以数百级为上限，避免 LLM 上下文溢出）
const MaxTocEntriesForLLM = 512

// CapTOCEntries 截断送入 TOC 相关性模型的目录条目数
func CapTOCEntries(entries []TOCFlattenEntry) []TOCFlattenEntry {
	if len(entries) <= MaxTocEntriesForLLM {
		return entries
	}
	return entries[:MaxTocEntriesForLLM]
}

// ragflowTocRelevanceSystem 摘自 references/ragflow/rag/prompts/toc_relevance_system.md（Ragflow PageIndex 检索用）
const ragflowTocRelevanceSystem = `# System Prompt: TOC Relevance Evaluation

You are an expert logical reasoning assistant specializing in hierarchical Table of Contents (TOC) relevance evaluation.

## GOAL
You will receive:
1. A JSON list of TOC items, each with fields:
   - "level": integer (e.g., 1, 2, 3)
   - "title": string (section title)
2. A user query (natural language question).

You must assign a **relevance score** (integer) to every TOC entry, based on how related its title is to the query.

## RULES

### Scoring System
- 5 → highly relevant (directly answers or matches the query intent)
- 3 → somewhat related (same topic or partially overlaps)
- 1 → weakly related (vague or tangential)
- 0 → no clear relation
- -1 → explicitly irrelevant or contradictory

### Hierarchy Traversal
- The TOC is hierarchical: smaller level = higher layer (e.g., level 1 is top-level, level 2 is a subsection).
- Traverse in hierarchical order — interpret the structure based on levels (1 > 2 > 3).
- If a high-level item (level 1) is strongly related (score 5), its child items (level 2, 3) are likely relevant too.
- If a high-level item is unrelated (-1 or 0), its deeper children are usually less relevant unless the titles clearly match the query.
- Lower (deeper) levels provide more specific content; prefer assigning higher scores if they directly match the query.

### Output Format
Return a **JSON array**, preserving the input order but adding a new key "score":

[
  {"level": 1, "title": "Introduction", "score": 0},
  {"level": 2, "title": "Definition of Sustainability", "score": 5}
]

### Constraints
- Output **only the JSON array** — no explanations or reasoning text.
`

const ragflowTocRelevanceUserTemplate = `You will now receive:
1. A JSON list of TOC items (each with level and title)
2. A user query string.

Traverse the TOC hierarchically based on level numbers and assign scores (5,3,1,0,-1) according to the rules in the system prompt.
Output **only** the JSON array with the added "score" field.

---

**Input TOC:**
%s

**Query:**
%s
`

// FlattenTreeToTOC 前序遍历树，level 从 1 起，与 Ragflow TOC 层级语义一致
func FlattenTreeToTOC(tree []TreeNode) []TOCFlattenEntry {
	var out []TOCFlattenEntry
	var walk func(nodes []TreeNode, depth int)
	walk = func(nodes []TreeNode, depth int) {
		for i := range nodes {
			n := &nodes[i]
			title := strings.TrimSpace(n.Title)
			if title == "" {
				title = strings.TrimSpace(GetNodeText(n))
			}
			if title == "" {
				title = "(untitled)"
			}
			id := NormalizeNodeID(n.NodeID)
			out = append(out, TOCFlattenEntry{Level: depth, Title: title, NodeID: id})
			if len(n.Nodes) > 0 {
				walk(n.Nodes, depth+1)
			}
		}
	}
	walk(tree, 1)
	return out
}

// TocRelevanceSearch 对齐 references/ragflow/rag/prompts/generator.py 中 relevant_chunks_with_toc：
// 仅向 LLM 暴露 level/title，按返回的 score 筛选（score>=1 且 score/5>=0.3），再映射回 node_id。
// Ragflow 中每条 TOC 还可对应多个 chunk id（ids）；混合检索路径见 PageIndexRagflowRetriever 对子树的 CollectDescendantNodeIDs 扩展。
func TocRelevanceSearch(ctx context.Context, llm interfaces.LLM, query string, entries []TOCFlattenEntry) ([]TOCScoredNode, error) {
	if llm == nil || len(entries) == 0 {
		return nil, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	lines := make([]string, len(entries))
	for i, e := range entries {
		b, err := json.Marshal(map[string]any{"level": e.Level, "title": e.Title})
		if err != nil {
			return nil, err
		}
		lines[i] = string(b)
	}
	tocJSON := "[\n" + strings.Join(lines, ",\n") + "\n]"
	user := fmt.Sprintf(ragflowTocRelevanceUserTemplate, tocJSON, query)

	msgs := []interfaces.MessageContent{
		{
			Role:  interfaces.MessageRoleSystem,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: ragflowTocRelevanceSystem}},
		},
		{
			Role:  interfaces.MessageRoleHuman,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: user}},
		},
	}
	resp, err := llm.GenerateContent(ctx, msgs,
		interfaces.WithTemperature(0),
		interfaces.WithTopP(0.9),
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Content) == "" {
		return nil, nil
	}
	raw := extractJSON(resp.Choices[0].Content)
	var scoredItems []struct {
		Level int             `json:"level"`
		Title string          `json:"title"`
		Score json.RawMessage `json:"score"`
	}
	if err := json.Unmarshal(raw, &scoredItems); err != nil {
		return nil, fmt.Errorf("解析 TOC 相关性 JSON 失败: %w", err)
	}
	var out []TOCScoredNode
	for i, item := range scoredItems {
		if i >= len(entries) {
			break
		}
		sv, ok := parseTOCScoreInt(item.Score)
		if !ok || sv < 1 {
			continue
		}
		norm := float64(sv) / 5.0
		if norm < 0.3 {
			continue
		}
		nid := NormalizeNodeID(entries[i].NodeID)
		out = append(out, TOCScoredNode{NodeID: nid, Score: norm})
	}
	return out, nil
}

func parseTOCScoreInt(raw json.RawMessage) (int, bool) {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return 0, false
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0, false
	}
	return int(f), true
}
