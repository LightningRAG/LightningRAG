package rag

import "strings"

// AugmentQueryWithLightningRAGKeywords 将 LightningRAG 风格的低层/高层关键词拼入检索用查询（用于向量与关键词分支的 query 文本）
// 展示给用户的正文仍应使用原始问题；仅检索调用使用本函数结果
func AugmentQueryWithLightningRAGKeywords(query string, hlKeywords, llKeywords []string) string {
	q := strings.TrimSpace(query)
	var parts []string
	for _, k := range llKeywords {
		if s := strings.TrimSpace(k); s != "" {
			parts = append(parts, s)
		}
	}
	for _, k := range hlKeywords {
		if s := strings.TrimSpace(k); s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return q
	}
	if q == "" {
		return strings.Join(parts, " ")
	}
	return q + "\n" + strings.Join(parts, " ")
}
