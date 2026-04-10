package rag

import (
	"encoding/json"
	"strings"
)

// FormatKnowledgeGraphPromptPrefix 将图谱实体/关系格式化为注入 RAG 上下文的文本，并按粗估 token 预算裁剪（对齐 LightRAG QueryParam.max_entity_tokens / max_relation_tokens）
// 版式对齐上游 PROMPTS["kg_query_context"]：Knowledge Graph Data (Entity)/(Relationship) + ```json``` 内每行一条 JSON 对象（JSONL）。
// maxEntityTokens、maxRelationTokens 为 0 时跳过对应区块；某区块在预算内未能写入任何一行则不输出该区块。
func FormatKnowledgeGraphPromptPrefix(entities, relationships []map[string]any, maxEntityTokens, maxRelationTokens uint) string {
	var sb strings.Builder
	if maxEntityTokens > 0 && len(entities) > 0 {
		if body := kgJSONLBlockWithinBudget(entities, int(maxEntityTokens)); body != "" {
			sb.WriteString("Knowledge Graph Data (Entity):\n\n```json\n")
			sb.WriteString(body)
			sb.WriteString("```\n")
		}
	}
	if maxRelationTokens > 0 && len(relationships) > 0 {
		if body := kgJSONLBlockWithinBudget(relationships, int(maxRelationTokens)); body != "" {
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString("Knowledge Graph Data (Relationship):\n\n```json\n")
			sb.WriteString(body)
			sb.WriteString("```\n")
		}
	}
	return strings.TrimSpace(sb.String())
}

// kgJSONLBlockWithinBudget 将若干 map 序列化为 JSONL（每行一个对象），在 token 预算内尽可能多写行。
func kgJSONLBlockWithinBudget(rows []map[string]any, budget int) string {
	if budget <= 0 || len(rows) == 0 {
		return ""
	}
	var sub strings.Builder
	for _, row := range rows {
		raw, err := json.Marshal(row)
		if err != nil {
			continue
		}
		line := string(raw) + "\n"
		if !appendLineWithinTokenBudget(&sub, &budget, line) {
			break
		}
	}
	return sub.String()
}

func appendLineWithinTokenBudget(sb *strings.Builder, budget *int, line string) bool {
	if *budget <= 0 {
		return false
	}
	t := estimateTokens(line)
	if t > *budget {
		const minTail = 24
		if *budget < minTail {
			return false
		}
		tokLimit := *budget
		line = truncateUTF8ByApproxTokens(strings.TrimSuffix(line, "\n"), tokLimit) + "\n"
		t = estimateTokens(line)
		// 粗算 token 与按字节截断可能差 1～2，用 tokLimit 递减重截断，勿动 *budget（仍留给后续 JSONL 行）
		for n := 0; n < 64 && t > *budget && tokLimit > minTail; n++ {
			tokLimit--
			line = truncateUTF8ByApproxTokens(strings.TrimSuffix(line, "\n"), tokLimit) + "\n"
			t = estimateTokens(line)
		}
	}
	if t > *budget {
		return false
	}
	sb.WriteString(line)
	*budget -= t
	return true
}
