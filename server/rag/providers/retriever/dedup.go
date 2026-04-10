package retriever

import (
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// DocumentDedupKey 多路检索合并时的去重键：优先 rag_chunk_id；否则 document_id+chunk_index；最后回退正文前缀。
// 用于 Fusion / MultiRetriever 轮询合并时占槽（与 references/LightRAG 多源合并思路一致）。
func DocumentDedupKey(d schema.Document) string {
	if d.Metadata != nil {
		if rc, ok := d.Metadata["rag_chunk_id"]; ok && rc != nil && fmt.Sprint(rc) != "" {
			return fmt.Sprintf("rag_chunk:%v", rc)
		}
		did, ok1 := d.Metadata["document_id"]
		ci, ok2 := d.Metadata["chunk_index"]
		if ok1 && ok2 {
			return fmt.Sprintf("doc:%v|chunk:%v", did, ci)
		}
	}
	key := d.PageContent
	if len(key) > 200 {
		key = key[:200]
	}
	return key
}

// RetrievedChunkIdentityKey 对齐 references/ragflow 按 chunk_id 合并：仅当有稳定切片标识时才返回可去重键。
// 无 rag_chunk_id 且无 document_id+chunk_index 时不与其它结果合并（避免误把不同切片当重复删掉）。
func RetrievedChunkIdentityKey(d schema.Document) (key string, ok bool) {
	if d.Metadata == nil {
		return "", false
	}
	if rc, has := d.Metadata["rag_chunk_id"]; has && rc != nil && fmt.Sprint(rc) != "" {
		return fmt.Sprintf("rag_chunk:%v", rc), true
	}
	did, ok1 := d.Metadata["document_id"]
	ci, ok2 := d.Metadata["chunk_index"]
	if ok1 && ok2 {
		return fmt.Sprintf("doc:%v|chunk:%v", did, ci), true
	}
	return "", false
}

// DeduplicateRetrievedDocuments 去掉「同一逻辑切片」因向量表重复行、多路重复返回产生的多余条（保留分数更高的一条）。
// 仅按 RetrievedChunkIdentityKey 合并，不按正文全文合并（避免 references/ragflow 所避免的、不同 chunk_id 同正文被错误折叠导致条数骤降）。
// 无稳定标识的条目逐条保留，顺序为首次出现顺序。
func DeduplicateRetrievedDocuments(docs []schema.Document) []schema.Document {
	if len(docs) <= 1 {
		return docs
	}
	keys := make([]string, len(docs))
	anon := 0
	for i, d := range docs {
		if k, ok := RetrievedChunkIdentityKey(d); ok {
			keys[i] = k
		} else {
			keys[i] = fmt.Sprintf("__unkeyed__:%d", anon)
			anon++
		}
	}
	best := make(map[string]schema.Document, len(docs))
	for i, d := range docs {
		k := keys[i]
		if old, has := best[k]; !has || d.Score > old.Score {
			best[k] = d
		}
	}
	seen := make(map[string]bool, len(best))
	out := make([]schema.Document, 0, len(best))
	for i := range docs {
		k := keys[i]
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, best[k])
	}
	return out
}
