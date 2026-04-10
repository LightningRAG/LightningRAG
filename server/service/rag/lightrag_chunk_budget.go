package rag

import (
	"slices"
	"sort"
	"strings"

	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// trimDocsToRagTokenBudget 按粗估 token 数限制「切片正文」总量（对齐 LightningRAG 对 chunk 上下文预算的思想；不含系统提示与 citation 模板开销）
// 会先按 Score 降序稳定排序再裁剪，使预算不足时优先保留高相关切片（与上游按相关性组织 chunks 的思路一致）。
// maxChunkTokens==0 时不裁剪正文，仍返回按分数排序后的副本（不修改入参切片元素顺序）。
func trimDocsToRagTokenBudget(docs []ragschema.Document, maxChunkTokens uint) []ragschema.Document {
	if len(docs) == 0 {
		return docs
	}
	ordered := slices.Clone(docs)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Score > ordered[j].Score
	})
	if maxChunkTokens == 0 {
		return ordered
	}
	budget := int(maxChunkTokens)
	out := make([]ragschema.Document, 0, len(ordered))
	for _, d := range ordered {
		n := estimateTokens(d.PageContent)
		if n <= budget {
			out = append(out, d)
			budget -= n
			continue
		}
		const minTailTok = 48
		if budget < minTailTok {
			break
		}
		trunc := truncateUTF8ByApproxTokens(d.PageContent, budget)
		if strings.TrimSpace(trunc) == "" {
			break
		}
		dc := d
		dc.PageContent = trunc
		out = append(out, dc)
		break
	}
	return out
}

func truncateUTF8ByApproxTokens(s string, maxTok int) string {
	if maxTok <= 0 {
		return ""
	}
	maxBytes := maxTok * 4
	if len(s) <= maxBytes {
		return s
	}
	b := []byte(s)
	for maxBytes > 0 && maxBytes < len(b) && b[maxBytes]&0xc0 == 0x80 {
		maxBytes--
	}
	if maxBytes <= 0 {
		return "…[truncated]"
	}
	return string(b[:maxBytes]) + "\n…[truncated]"
}
