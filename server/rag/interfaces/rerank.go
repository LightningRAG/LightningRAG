// Package interfaces 定义 Rerank 重排序接口，参考 references 目录内 rerank_model
// 用于对检索结果按相关性重新排序
package interfaces

import "context"

// Reranker 重排序接口，参考 references 目录内 rerank_model.Base
// 新提供商只需实现此接口即可接入
type Reranker interface {
	// Rerank 对 query 与 texts 计算相关性分数，返回与 texts 同序的分数数组
	// 分数越高表示越相关
	Rerank(ctx context.Context, query string, texts []string) ([]float32, error)

	// ProviderName 返回提供商名称，如 "jina", "cohere"
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string
}
