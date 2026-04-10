package retriever

import (
	"context"
	"sort"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// ragflowRerankCandidateFloor 对齐 references/ragflow/rag/nlp/search.py：RERANK_LIMIT = max(30, ceil(64/page_size)*page_size)
func ragflowRerankCandidateFloor(pageSize int) int {
	if pageSize <= 1 {
		return 30
	}
	r := ((64 + pageSize - 1) / pageSize) * pageSize
	if r < 30 {
		return 30
	}
	return r
}

// RerankRetriever 带 Rerank 重排序的检索器包装器
// 先通过 inner 检索器获取候选文档，再用 Reranker 重新计算相关性分数并排序
type RerankRetriever struct {
	inner          interfaces.Retriever
	reranker       interfaces.Reranker
	topK           int     // 从 inner 检索的候选数量（0=自动，取 numDocs 的 3 倍且至少 30）
	minRerankScore float32 // Rerank 分数下限，0 表示不过滤（各提供商分数尺度可能不同）
}

// NewRerankRetriever 创建 Rerank 检索器；minRerankScore<=0 时不按分数丢弃结果
func NewRerankRetriever(inner interfaces.Retriever, reranker interfaces.Reranker, topK int, minRerankScore float32) *RerankRetriever {
	return &RerankRetriever{
		inner:          inner,
		reranker:       reranker,
		topK:           topK,
		minRerankScore: minRerankScore,
	}
}

func (r *RerankRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	candidateK := r.topK
	if candidateK <= 0 {
		candidateK = numDocs * 3
		if candidateK < 30 {
			candidateK = 30
		}
	}
	// 显式 RerankTopK 过小时，重排前候选不足会导致截断后条数远少于 numDocs（对齐 LightRAG 先宽池再 chunk_top_k 截断）
	if candidateK < numDocs {
		candidateK = numDocs
	}
	if candidateK < lightragconst.DefaultTopK {
		candidateK = lightragconst.DefaultTopK
	}
	if candidateK > lightragconst.MaxSimilarityFetchK {
		candidateK = lightragconst.MaxSimilarityFetchK
	}
	if floor := ragflowRerankCandidateFloor(numDocs); candidateK < floor {
		candidateK = floor
		if candidateK > lightragconst.MaxSimilarityFetchK {
			candidateK = lightragconst.MaxSimilarityFetchK
		}
	}

	candidates, err := r.inner.GetRelevantDocuments(ctx, query, candidateK)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return candidates, nil
	}

	texts := make([]string, len(candidates))
	for i, doc := range candidates {
		texts[i] = doc.PageContent
	}

	scores, err := r.reranker.Rerank(ctx, query, texts)
	if err != nil {
		// Rerank 失败时降级为原始排序结果
		if numDocs > 0 && len(candidates) > numDocs {
			candidates = candidates[:numDocs]
		}
		return candidates, nil
	}

	for i := range candidates {
		if i < len(scores) {
			candidates[i].Score = scores[i]
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if r.minRerankScore > 0 {
		kept := make([]schema.Document, 0, len(candidates))
		for _, d := range candidates {
			if d.Score >= r.minRerankScore {
				kept = append(kept, d)
			}
		}
		candidates = kept
	}

	if numDocs > 0 && len(candidates) > numDocs {
		candidates = candidates[:numDocs]
	}

	return candidates, nil
}

func (r *RerankRetriever) RetrieverType() interfaces.RetrieverType {
	return r.inner.RetrieverType()
}
