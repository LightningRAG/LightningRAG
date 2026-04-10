package retriever

import (
	"context"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// VectorRetriever 基于向量存储的检索器
type VectorRetriever struct {
	store          interfaces.VectorStore
	namespace      string
	numDocs        int
	reportType     interfaces.RetrieverType
	scoreThreshold float32 // 余弦类相似度下限，0 表示不过滤（对齐 LightningRAG 类 cosine 阈值用法）
}

// WithScoreThreshold 设置向量相似度下限；t<=0 时等价于关闭过滤
func (r *VectorRetriever) WithScoreThreshold(t float32) *VectorRetriever {
	if t > 0 {
		r.scoreThreshold = t
	}
	return r
}

// NewVectorRetriever 创建向量检索器；可选 reportType 用于 global 等（默认 vector）；纯向量切片检索与 references/ragflow 向量腿一致（宽召回 topk 上限见 lightragconst.MaxSimilarityFetchK）
func NewVectorRetriever(store interfaces.VectorStore, namespace string, numDocs int, reportType ...interfaces.RetrieverType) *VectorRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	rt := interfaces.RetrieverTypeVector
	if len(reportType) > 0 && reportType[0] != "" {
		rt = reportType[0]
	}
	return &VectorRetriever{
		store:      store,
		namespace:  namespace,
		numDocs:    numDocs,
		reportType: rt,
	}
}

// GetRelevantDocuments 检索相关文档
func (r *VectorRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = r.numDocs
	}
	// 向量切片：先按宽池拉候选再截断到 n（对齐 RAGFlow Dealer.retrieval 向量 top、LightRAG chunk_top_k / rerank 前大池）
	fetchN := lightragconst.WideSimilarityFetchK(n, 4)
	if r.scoreThreshold > 0 {
		fetchN = lightragconst.WideSimilarityFetchK(n, 10)
	}
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = r.namespace },
	}
	docs, err := r.store.SimilaritySearch(ctx, query, fetchN, opts...)
	if err != nil {
		return nil, err
	}
	if r.scoreThreshold > 0 {
		filtered := make([]schema.Document, 0, len(docs))
		for _, d := range docs {
			if d.Score >= r.scoreThreshold {
				filtered = append(filtered, d)
			}
		}
		docs = filtered
	}
	if len(docs) > n {
		docs = docs[:n]
	}
	return docs, nil
}

// RetrieverType 返回检索类型
func (r *VectorRetriever) RetrieverType() interfaces.RetrieverType {
	return r.reportType
}
