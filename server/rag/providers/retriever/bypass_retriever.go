package retriever

import (
	"context"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// BypassRetriever 不返回任何切片，对应 LightningRAG 的 bypass 模式（直连 LLM、无知识库上下文）
type BypassRetriever struct{}

// NewBypassRetriever 创建 bypass 检索器
func NewBypassRetriever() *BypassRetriever {
	return &BypassRetriever{}
}

// GetRelevantDocuments 始终返回空列表
func (r *BypassRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	return nil, nil
}

// RetrieverType 返回 bypass
func (r *BypassRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypeBypass
}
