// Package lightragconst 与 references/LightRAG/lightrag/constants.py 中查询/检索默认值对齐，
// 供 service/rag 与各 retriever 共用，避免魔法数分叉。
package lightragconst

const (
	// DefaultTopK 对应 DEFAULT_TOP_K：实体/关系向量检索等「宽池」条数。
	DefaultTopK = 40
	// DefaultChunkTopK 对应 DEFAULT_CHUNK_TOP_K：注入上下文的文本切片条数。
	DefaultChunkTopK = 20
	// MaxSimilarityFetchK 单次向量/BM25 查询上限；与 references/ragflow/rag/nlp/search.py 中 Dealer.retrieval(top=1024)、Dealer.search(topk=1024) 同量级，避免宽召回被过早截断。
	MaxSimilarityFetchK = 1024
)

// WideSimilarityFetchK 在最终截断到 finalN 条之前，向向量库请求的候选条数（与 LightRAG operate 中 mix/hybrid 先宽召回再融合、再截断一致）。
func WideSimilarityFetchK(finalN int, multiplier int) int {
	if finalN <= 0 {
		finalN = DefaultChunkTopK
	}
	if multiplier < 1 {
		multiplier = 1
	}
	k := finalN * multiplier
	if k < DefaultTopK {
		k = DefaultTopK
	}
	if k > MaxSimilarityFetchK {
		k = MaxSimilarityFetchK
	}
	return k
}
