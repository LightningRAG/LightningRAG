package retriever

import (
	"context"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// KeywordRetriever 基于 VectorStore.KeywordSearch 的全文/关键词检索
type KeywordRetriever struct {
	store      interfaces.VectorStore
	namespace  string
	numDocs    int
	reportType interfaces.RetrieverType
}

// NewKeywordRetriever 创建关键词检索器；可选 reportType 用于 local 等模式（默认 keyword）
func NewKeywordRetriever(store interfaces.VectorStore, namespace string, numDocs int, reportType ...interfaces.RetrieverType) *KeywordRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	rt := interfaces.RetrieverTypeKeyword
	if len(reportType) > 0 && reportType[0] != "" {
		rt = reportType[0]
	}
	return &KeywordRetriever{store: store, namespace: namespace, numDocs: numDocs, reportType: rt}
}

// GetRelevantDocuments 检索相关文档
func (r *KeywordRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = r.numDocs
	}
	fetch := lightragconst.WideSimilarityFetchK(n, 4)
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = r.namespace },
	}
	docs, err := r.store.KeywordSearch(ctx, query, fetch, opts...)
	if err != nil {
		return nil, err
	}
	if len(docs) > n {
		docs = docs[:n]
	}
	return docs, nil
}

// RetrieverType 返回检索类型
func (r *KeywordRetriever) RetrieverType() interfaces.RetrieverType {
	return r.reportType
}
