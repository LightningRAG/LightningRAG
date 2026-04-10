package rag

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// retrievalEnabledFilterRetriever 过滤「检索已禁用」的文档对应的切片（向量检索 + Rerank 之后统一过滤）
type retrievalEnabledFilterRetriever struct {
	inner   interfaces.Retriever
	kbID    uint
	numDocs int
}

func newRetrievalEnabledFilterRetriever(kbID uint, inner interfaces.Retriever, numDocs int) interfaces.Retriever {
	if numDocs <= 0 {
		numDocs = DefaultConversationRAGTopK
	}
	return &retrievalEnabledFilterRetriever{inner: inner, kbID: kbID, numDocs: numDocs}
}

func (r *retrievalEnabledFilterRetriever) RetrieverType() interfaces.RetrieverType {
	return r.inner.RetrieverType()
}

func (r *retrievalEnabledFilterRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = r.numDocs
	}
	var disabled []uint
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("knowledge_base_id = ? AND retrieval_enabled = ?", r.kbID, false).
		Pluck("id", &disabled).Error; err != nil {
		return nil, err
	}
	if len(disabled) == 0 {
		return r.inner.GetRelevantDocuments(ctx, query, n)
	}
	blocked := make(map[uint]struct{}, len(disabled))
	for _, id := range disabled {
		blocked[id] = struct{}{}
	}
	// 部分文档关闭检索时，过小 oversample 会在过滤后明显少于 n（Ragflow 先宽池 topk=1024 再截断）
	fetch := lightragconst.WideSimilarityFetchK(n, 8)
	if fetch < n*6 {
		fetch = n * 6
	}
	if fetch > lightragconst.MaxSimilarityFetchK {
		fetch = lightragconst.MaxSimilarityFetchK
	}
	docs, err := r.inner.GetRelevantDocuments(ctx, query, fetch)
	if err != nil {
		return nil, err
	}
	var out []schema.Document
	for _, d := range docs {
		did, ok := metadataDocumentID(d.Metadata)
		if !ok {
			out = append(out, d)
			if len(out) >= n {
				break
			}
			continue
		}
		if _, bad := blocked[did]; bad {
			continue
		}
		out = append(out, d)
		if len(out) >= n {
			break
		}
	}
	return out, nil
}

func metadataDocumentID(meta map[string]any) (uint, bool) {
	if meta == nil {
		return 0, false
	}
	raw, ok := meta["document_id"]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return uint(v), true
	case float32:
		return uint(v), true
	case int:
		return uint(v), true
	case int64:
		return uint(v), true
	case uint:
		return v, true
	case uint64:
		return uint(v), true
	case json.Number:
		u64, err := v.Int64()
		if err != nil || u64 < 0 {
			return 0, false
		}
		return uint(u64), true
	case string:
		u, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(u), true
	default:
		return 0, false
	}
}
