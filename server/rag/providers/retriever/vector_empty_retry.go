package retriever

import (
	"context"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// VectorEmptyRetryRetriever 纯向量检索零命中时，用无相似度阈值、宽池再搜一轮（对齐 references/ragflow search 第二次放宽 similarity / 仅向量腿）。
type VectorEmptyRetryRetriever struct {
	inner *VectorRetriever
}

// WrapVectorEmptyRetry 包装 *VectorRetriever；nil 时返回 nil。
func WrapVectorEmptyRetry(inner *VectorRetriever) interfaces.Retriever {
	if inner == nil {
		return nil
	}
	return &VectorEmptyRetryRetriever{inner: inner}
}

func (w *VectorEmptyRetryRetriever) RetrieverType() interfaces.RetrieverType {
	return w.inner.RetrieverType()
}

func (w *VectorEmptyRetryRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	docs, err := w.inner.GetRelevantDocuments(ctx, query, numDocs)
	if err != nil || len(docs) > 0 || strings.TrimSpace(query) == "" {
		return docs, err
	}
	relax := NewVectorRetriever(w.inner.store, w.inner.namespace, lightragconst.MaxSimilarityFetchK)
	relax.reportType = w.inner.reportType
	relax.scoreThreshold = 0
	return relax.GetRelevantDocuments(ctx, query, numDocs)
}
