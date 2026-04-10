package interfaces

import (
	"context"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// RetrieverType 检索类型
type RetrieverType string

const (
	RetrieverTypeVector    RetrieverType = "vector"    // 向量相似度检索
	RetrieverTypePageIndex RetrieverType = "pageindex" // PageIndex 树状推理检索（如支持）
	RetrieverTypeKeyword   RetrieverType = "keyword"   // 关键词/全文检索
	// 以下与 references/LightRAG QueryParam.mode 对齐；知识库启用图谱且已有实体数据时 local/global/hybrid/mix 走图谱检索，否则为向量/关键词近似策略
	RetrieverTypeNaive  RetrieverType = "naive"  // 已并入 vector，仅保留常量供历史数据/序列化兼容
	RetrieverTypeLocal  RetrieverType = "local"  // 近似：偏词汇/全文（无图谱时用 KeywordSearch）
	RetrieverTypeGlobal RetrieverType = "global" // 近似：偏语义向量（等同 vector）
	RetrieverTypeHybrid RetrieverType = "hybrid" // 向量 + 关键词融合（近似 LightningRAG local+global）
	RetrieverTypeMix    RetrieverType = "mix"    // 融合且更侧重向量（近似 LightningRAG 图谱+向量）
	RetrieverTypeBypass RetrieverType = "bypass" // 不检索，空上下文（直连 LLM）
)

// Retriever 检索器接口
// 支持多种检索方式：向量检索、PageIndex 树检索、全文检索
type Retriever interface {
	// GetRelevantDocuments 根据查询获取相关文档
	GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error)

	// RetrieverType 返回检索类型
	RetrieverType() RetrieverType
}
