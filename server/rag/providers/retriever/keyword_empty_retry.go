package retriever

import (
	"context"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// KeywordEmptyRetryRetriever 全文检索零命中时，以 RelaxedKeywordSearch 再搜一轮（各 VectorStore 实现 Ragflow 式放宽）。
type KeywordEmptyRetryRetriever struct {
	inner *KeywordRetriever
}

// WrapKeywordEmptyRetry 包装 *KeywordRetriever；nil 时返回 nil。
func WrapKeywordEmptyRetry(inner *KeywordRetriever) interfaces.Retriever {
	if inner == nil {
		return nil
	}
	return &KeywordEmptyRetryRetriever{inner: inner}
}

func (w *KeywordEmptyRetryRetriever) RetrieverType() interfaces.RetrieverType {
	return w.inner.RetrieverType()
}

func (w *KeywordEmptyRetryRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	docs, err := w.inner.GetRelevantDocuments(ctx, query, numDocs)
	if err != nil || len(docs) > 0 || strings.TrimSpace(query) == "" {
		return docs, err
	}
	n := numDocs
	if n <= 0 {
		n = w.inner.numDocs
	}
	fetch := lightragconst.WideSimilarityFetchK(n, 4)
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = w.inner.namespace },
		interfaces.WithRelaxedKeywordSearch(true),
	}
	docs2, err2 := w.inner.store.KeywordSearch(ctx, strings.TrimSpace(query), fetch, opts...)
	if err2 != nil {
		return nil, err2
	}
	if len(docs2) > n {
		docs2 = docs2[:n]
	}
	return docs2, nil
}
