package rag

// ExposeReferencesForAPI 按 references/LightRAG QueryRequest.include_references、include_chunk_content 裁剪返回给客户端的引用。
// 库内 prompt 仍使用完整切片；仅影响 HTTP 响应与流式 references 事件。
// includeReferences 显式 false 时不返回引用列表；includeChunkContent 显式 false 时从每条引用中去掉 content 字段。
// includeChunkContent 为 nil 时保留 content，以兼容现有前端。
func ExposeReferencesForAPI(refs []map[string]any, includeReferences *bool, includeChunkContent *bool) []map[string]any {
	if includeReferences != nil && !*includeReferences {
		return nil
	}
	if len(refs) == 0 {
		return refs
	}
	if includeChunkContent == nil || *includeChunkContent {
		return refs
	}
	out := make([]map[string]any, len(refs))
	for i, ref := range refs {
		cp := make(map[string]any, len(ref))
		for k, v := range ref {
			if k == "content" {
				continue
			}
			cp[k] = v
		}
		out[i] = cp
	}
	return out
}
