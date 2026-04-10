package rerank

import "github.com/LightningRAG/LightningRAG/server/rag/interfaces"

// ProviderNameAdapter 包装 Reranker，使同一实现可注册为多个提供商名（如 dashscope / tongyi）
type ProviderNameAdapter struct {
	interfaces.Reranker
	DisplayName string
}

func (p *ProviderNameAdapter) ProviderName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Reranker.ProviderName()
}
