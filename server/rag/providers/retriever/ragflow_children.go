package retriever

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// batchRagChunksByIDs 批量加载切片，用于给向量命中补全 DB 中的 parent_rag_chunk_id（旧索引未写入向量元数据时仍可对齐 Ragflow retrieval_by_children）。
func batchRagChunksByIDs(ctx context.Context, docs []schema.Document) map[uint]*rag.RagChunk {
	seen := make(map[uint]struct{})
	var ids []uint
	for _, d := range docs {
		rid, ok := metadataUintFromMeta(d.Metadata, "rag_chunk_id")
		if !ok || rid == 0 {
			continue
		}
		if _, dup := seen[rid]; dup {
			continue
		}
		seen[rid] = struct{}{}
		ids = append(ids, rid)
	}
	if len(ids) == 0 {
		return nil
	}
	var rows []rag.RagChunk
	if err := global.LRAG_DB.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil
	}
	m := make(map[uint]*rag.RagChunk, len(rows))
	for i := range rows {
		m[rows[i].ID] = &rows[i]
	}
	return m
}

// mergeChunkMetadataFromDB 将 rag_chunks.metadata（JSON）中的父子线索并入检索 metadata，供 retrieval_by_children 使用。
func mergeChunkMetadataFromDB(ch *rag.RagChunk, meta map[string]any) {
	if ch == nil || meta == nil || strings.TrimSpace(ch.Metadata) == "" {
		return
	}
	var extra map[string]any
	if err := json.Unmarshal([]byte(ch.Metadata), &extra); err != nil {
		return
	}
	for _, k := range []string{"parent_rag_chunk_id", "node_id"} {
		if v, ok := extra[k]; ok {
			if _, exists := meta[k]; !exists {
				meta[k] = v
			}
		}
	}
}

// MergeRetrievalByChildren 对齐 references/ragflow/rag/nlp/search.py retrieval_by_children：
// 对 metadata 中含 parent_rag_chunk_id 的命中（子切片）按父 rag_chunk_id 分组合并：分数取子块均值，正文用父块在 DB 中的内容；并从结果中移除已合并的子块。
// 无父子元数据时原样返回。
func MergeRetrievalByChildren(ctx context.Context, docs []schema.Document) []schema.Document {
	if len(docs) == 0 {
		return docs
	}
	groups := make(map[uint][]schema.Document)
	var roots []schema.Document
	for _, d := range docs {
		pid, ok := metadataUintFromMeta(d.Metadata, "parent_rag_chunk_id")
		if !ok || pid == 0 {
			roots = append(roots, d)
			continue
		}
		groups[pid] = append(groups[pid], d)
	}
	if len(groups) == 0 {
		return docs
	}

	parentHit := make(map[uint]struct{})
	for pid := range groups {
		parentHit[pid] = struct{}{}
	}

	filtered := roots[:0]
	for _, d := range roots {
		rid, ok := metadataUintFromMeta(d.Metadata, "rag_chunk_id")
		if ok && rid != 0 {
			if _, dup := parentHit[rid]; dup {
				continue
			}
		}
		filtered = append(filtered, d)
	}

	out := make([]schema.Document, 0, len(filtered)+len(groups))
	out = append(out, filtered...)

	for parentID, children := range groups {
		if len(children) == 0 {
			continue
		}
		var parentCh rag.RagChunk
		if err := global.LRAG_DB.WithContext(ctx).First(&parentCh, parentID).Error; err != nil {
			out = append(out, children...)
			continue
		}
		var sum float64
		for _, c := range children {
			sum += float64(c.Score)
		}
		mean := float32(sum / float64(len(children)))
		meta := map[string]any{
			"document_id":        parentCh.DocumentID,
			"rag_chunk_id":       parentCh.ID,
			"chunk_index":        parentCh.ChunkIndex,
			"pageindex_mode":     "ragflow_parent_merge",
			"merged_child_count": len(children),
		}
		if children[0].Metadata != nil {
			if v, ok := children[0].Metadata["doc_name"]; ok {
				meta["doc_name"] = v
			}
		}
		out = append(out, schema.Document{
			PageContent: parentCh.Content,
			Score:       mean,
			Metadata:    meta,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		si, sj := out[i].Score, out[j].Score
		if si != sj {
			return si > sj
		}
		li, lj := len(out[i].PageContent), len(out[j].PageContent)
		if li != lj {
			return li > lj
		}
		return out[i].PageContent < out[j].PageContent
	})
	return out
}
