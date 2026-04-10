package interfaces

import (
	"context"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// VectorStoreOption 向量存储选项
type VectorStoreOption func(*VectorStoreOptions)

// VectorStoreOptions 向量存储选项结构
type VectorStoreOptions struct {
	Namespace      string
	ScoreThreshold float32
	Filters        map[string]any
	// RelaxedKeywordSearch 为 true 时 KeywordSearch 使用更宽松条件（对齐 Ragflow search 第二次降低 min_match）
	RelaxedKeywordSearch bool
}

// WithRelaxedKeywordSearch 控制全文检索是否走宽松策略（由各 VectorStore 解释）
func WithRelaxedKeywordSearch(relaxed bool) VectorStoreOption {
	return func(o *VectorStoreOptions) {
		o.RelaxedKeywordSearch = relaxed
	}
}

// VectorStore 向量存储接口，参考 langchaingo VectorStore
// 支持 PostgreSQL+pgvector、Elasticsearch 等，新存储只需实现此接口
type VectorStore interface {
	// AddDocuments 添加文档切片到存储
	AddDocuments(ctx context.Context, docs []schema.Document, options ...VectorStoreOption) ([]string, error)

	// SimilaritySearch 相似度检索
	SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...VectorStoreOption) ([]schema.Document, error)

	// KeywordSearch 关键词/全文检索（与 LightningRAG 中基于词汇的检索分支对应；各存储实现 ES match、PG ts/MySQL 词项匹配等）
	KeywordSearch(ctx context.Context, query string, numDocuments int, options ...VectorStoreOption) ([]schema.Document, error)

	// DeleteByIDs 按 ID 删除
	DeleteByIDs(ctx context.Context, ids []string) error

	// DeleteByNamespace 按命名空间删除（如知识库 ID）
	DeleteByNamespace(ctx context.Context, namespace string) error

	// DeleteByMetadata 按 metadata 键值删除（如 document_id）
	DeleteByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) error

	// ListByMetadata 按 metadata 键值查询文档（如按 document_id 查询所有切片）
	// 返回匹配的文档列表和总数；用于从向量库恢复切片到关系库
	ListByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) ([]schema.Document, error)

	// ProviderName 返回存储提供商名称
	ProviderName() string
}
