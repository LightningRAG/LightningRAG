package interfaces

import (
	"context"
)

// Embedder 向量嵌入接口，参考 langchaingo Embedder 和 LightningRAG embedding_model
// 新提供商只需实现此接口即可接入
type Embedder interface {
	// EmbedDocuments 为多个文本生成向量
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)

	// EmbedQuery 为单个查询文本生成向量
	EmbedQuery(ctx context.Context, text string) ([]float32, error)

	// ProviderName 返回提供商名称
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string

	// Dimensions 返回向量维度，0 表示不固定
	Dimensions() int
}
