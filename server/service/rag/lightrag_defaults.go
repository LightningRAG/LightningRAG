package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
)

// 本文件集中 RAG 检索与对话的默认参数，与 references/LightRAG/lightrag/constants.py 中
// DEFAULT_TOP_K / DEFAULT_CHUNK_TOP_K 及 lightrag/lightrag.py QueryParam 对齐。
// max_entity_tokens / max_relation_tokens：请求体优先，亦可通过 config.rag.default-max-entity-context-tokens 等设服务端默认（见 EffectiveMaxEntityContextTokens）
// 切片正文预算：max_rag_context_tokens 类字段见 EffectiveMaxRagContextTokens 与 default-max-rag-context-tokens

const (
	// DefaultConversationRAGTopK 单轮对话最终保留的切片条数（对齐 LightRAG chunk_top_k，默认 20）
	// 变更时请同步前端 web/src/view/rag/conversation/conversation.vue 中 RAG_CODE_DEFAULT_CHUNK_TOP_K
	DefaultConversationRAGTopK = lightragconst.DefaultChunkTopK
	// DefaultKnowledgeBaseRetrieveTopN 「文档检索」页默认返回条数（同 chunk_top_k 默认 20）
	DefaultKnowledgeBaseRetrieveTopN = lightragconst.DefaultChunkTopK
	// MaxRetrieveTopN 单次检索最终返回/注入条数上限（Ragflow 的 dialog page_size 量级）
	MaxRetrieveTopN = 50
	// MaxRetrieveCandidateTopK 向量宽召回候选上限（Ragflow Dealer.retrieval 默认 top=1024）
	MaxRetrieveCandidateTopK = lightragconst.MaxSimilarityFetchK
	// MinRetrieveTopN 单次检索条数下限
	MinRetrieveTopN = 1
	// DefaultConversationHistoryMaxMessages 对话从 DB 拉取历史条数默认上限（单条消息计 1）
	DefaultConversationHistoryMaxMessages = 20
	constMaxConversationHistoryMessages   = 200
)

// EffectiveMaxRetrieveTopN 单次检索条数上限（config.rag.max-retrieve-top-n 可覆盖默认 50）
func EffectiveMaxRetrieveTopN() int {
	m := global.LRAG_CONFIG.Rag.MaxRetrieveTopN
	if m <= 0 {
		return MaxRetrieveTopN
	}
	if m < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	return m
}

// EffectiveMaxRetrieveCandidateTopK 向量/融合检索候选池上限（config.rag.max-retrieve-candidate-top-k；0 表示 1024，与 Ragflow topk 默认一致）
// EffectiveHybridFusionWeights 返回 hybrid/mix 全文与向量融合权重（和为 1）；均未配置时用 Ragflow 0.05/0.95
func EffectiveHybridFusionWeights() (termW, vecW float32) {
	t := global.LRAG_CONFIG.Rag.HybridFusionTermWeight
	v := global.LRAG_CONFIG.Rag.HybridFusionVectorWeight
	if t <= 0 && v <= 0 {
		return 0.05, 0.95
	}
	if t < 0 {
		t = 0
	}
	if v < 0 {
		v = 0
	}
	ft, fv := float32(t), float32(v)
	sum := ft + fv
	if sum <= 0 {
		return 0.05, 0.95
	}
	return ft / sum, fv / sum
}

// EffectiveHybridFusionMinScore 融合分过滤下限；<=0 关闭
func EffectiveHybridFusionMinScore() float32 {
	c := global.LRAG_CONFIG.Rag.HybridFusionMinScore
	if c <= 0 {
		return 0
	}
	if c > 1 {
		return 1
	}
	return float32(c)
}

// EffectiveHybridFusionEmptyRetry hybrid/mix 是否在零命中后做纯向量宽召回重试
func EffectiveHybridFusionEmptyRetry() bool {
	return !global.LRAG_CONFIG.Rag.HybridFusionSkipEmptyRetry
}

// EffectiveVectorEmptyRetry 纯向量 / PageIndex 向量腿是否在零命中后做无阈值宽池重试（Ragflow search 第二次放宽）
func EffectiveVectorEmptyRetry() bool {
	return !global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry
}

// EffectiveKeywordEmptyRetry keyword/local（无图谱）是否在零命中后做宽松全文重试
func EffectiveKeywordEmptyRetry() bool {
	return !global.LRAG_CONFIG.Rag.KeywordSkipEmptyRetry
}

// EffectiveElasticsearchScoreRankBoostWeight 向量存储检索对 metadata.rank_boost 的乘性加成系数（配置名含 elasticsearch，PG/MySQL 同样生效）；<=0 关闭
func EffectiveElasticsearchScoreRankBoostWeight() float32 {
	w := global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight
	if w <= 0 {
		return 0
	}
	if w > 50 {
		return 50
	}
	return float32(w)
}

func EffectiveMaxRetrieveCandidateTopK() int {
	m := global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK
	if m <= 0 {
		m = MaxRetrieveCandidateTopK
	}
	if m < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	maxCap := lightragconst.MaxSimilarityFetchK
	if m > maxCap {
		return maxCap
	}
	return m
}

// DefaultConversationChunkTopKFromConfig 对话默认切片条数（config.rag.default-conversation-chunk-top-k）
func DefaultConversationChunkTopKFromConfig() int {
	c := global.LRAG_CONFIG.Rag.DefaultConversationChunkTopK
	if c <= 0 {
		return DefaultConversationRAGTopK
	}
	maxN := EffectiveMaxRetrieveTopN()
	if c > maxN {
		return maxN
	}
	if c < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	return c
}

// DefaultKnowledgeBaseRetrieveTopNFromConfig 文档检索页默认条数（config.rag.default-knowledge-base-retrieve-top-n）
func DefaultKnowledgeBaseRetrieveTopNFromConfig() int {
	c := global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrieveTopN
	if c <= 0 {
		return DefaultKnowledgeBaseRetrieveTopN
	}
	maxN := EffectiveMaxRetrieveTopN()
	if c > maxN {
		return maxN
	}
	if c < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	return c
}

// EffectiveDefaultKnowledgeBaseRetrievePoolTopK 文档检索未传 topK 时的默认候选池上限（配置 0 时使用 Ragflow 式宽召回，与 max-retrieve-candidate-top-k 对齐）
func EffectiveDefaultKnowledgeBaseRetrievePoolTopK() int {
	c := global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK
	maxCand := EffectiveMaxRetrieveCandidateTopK()
	if c <= 0 {
		c = maxCand
	} else if c > maxCand {
		c = maxCand
	}
	if c < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	return c
}

// EffectiveDefaultConversationRetrievePoolTopK 对话 / queryData 未传 topK 时的默认候选池上限（配置 0 时使用 Ragflow 式宽召回）
func EffectiveDefaultConversationRetrievePoolTopK() int {
	c := global.LRAG_CONFIG.Rag.DefaultConversationRetrievePoolTopK
	maxCand := EffectiveMaxRetrieveCandidateTopK()
	if c <= 0 {
		c = maxCand
	} else if c > maxCand {
		c = maxCand
	}
	if c < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	return c
}

// resolveRetrieveCandidateCount 最终保留 finalN 条；若 poolTopK（LightningRAG top_k）更大则先按更大候选数检索再截断到 finalN。候选池上限与「最终返回上限」解耦，对齐 Ragflow topk=1024 vs page_size。
func resolveRetrieveCandidateCount(finalN int, poolTopK *int) int {
	maxReturn := EffectiveMaxRetrieveTopN()
	maxCand := EffectiveMaxRetrieveCandidateTopK()
	if finalN > maxReturn {
		finalN = maxReturn
	}
	candidate := finalN
	if poolTopK != nil && *poolTopK > 0 {
		p := *poolTopK
		if p > maxCand {
			p = maxCand
		}
		if p < MinRetrieveTopN {
			p = MinRetrieveTopN
		}
		if p > candidate {
			candidate = p
		}
	} else {
		// 未传 poolTopK：检索阶段至少拉 LightRAG top_k（40）条候选，再在 fetch 末尾截断到 finalN（chunk_top_k）
		if candidate < lightragconst.DefaultTopK {
			candidate = lightragconst.DefaultTopK
		}
		if candidate > maxCand {
			candidate = maxCand
		}
	}
	return candidate
}

// ClampRetrieveTopN 将 topN 限制在 [MinRetrieveTopN, EffectiveMaxRetrieveTopN]；n<=0 时用 defaultN
func ClampRetrieveTopN(n, defaultN int) int {
	maxN := EffectiveMaxRetrieveTopN()
	if defaultN <= 0 {
		defaultN = DefaultConversationRAGTopK
	}
	if n <= 0 {
		n = defaultN
	}
	if n > maxN {
		return maxN
	}
	if n < MinRetrieveTopN {
		return MinRetrieveTopN
	}
	return n
}

// EffectiveMaxEntityContextTokens 请求 maxEntityTokens 优先；未指定或非正时用 config.rag.default-max-entity-context-tokens（对齐 LightRAG max_entity_tokens 可配置默认）
func EffectiveMaxEntityContextTokens(p *uint) uint {
	if p != nil && *p > 0 {
		return *p
	}
	c := global.LRAG_CONFIG.Rag.DefaultMaxEntityContextTokens
	if c > 0 {
		return uint(c)
	}
	return 0
}

// EffectiveMaxRelationContextTokens 请求 maxRelationTokens 优先；未指定或非正时用 config.rag.default-max-relation-context-tokens
func EffectiveMaxRelationContextTokens(p *uint) uint {
	if p != nil && *p > 0 {
		return *p
	}
	c := global.LRAG_CONFIG.Rag.DefaultMaxRelationContextTokens
	if c > 0 {
		return uint(c)
	}
	return 0
}

// EffectiveMaxRagContextTokens 请求 maxRagContextTokens 优先；未指定或非正时用 config.rag.default-max-rag-context-tokens（对齐 LightRAG 对 chunk 上下文可设默认上限）
func EffectiveMaxRagContextTokens(p *uint) uint {
	if p != nil && *p > 0 {
		return *p
	}
	c := global.LRAG_CONFIG.Rag.DefaultMaxRagContextTokens
	if c > 0 {
		return uint(c)
	}
	return 0
}

const maxKgPromptNeighborRelLimitCap = 500

// EffectiveKgPromptNeighborRelLimit 图谱 prompt / QueryData 中除切片直连关系外，最多追加多少条邻接关系（一跳）；0 关闭
func EffectiveKgPromptNeighborRelLimit() int {
	c := global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit
	if c <= 0 {
		return 0
	}
	if c > maxKgPromptNeighborRelLimitCap {
		return maxKgPromptNeighborRelLimitCap
	}
	return c
}

// EffectiveDefaultCosineThreshold 请求未传 cosineThreshold 时使用的相似度下限（与 LightRAG DEFAULT_COSINE_THRESHOLD 同源）；<=0 或 >1 表示关闭服务端默认
func EffectiveDefaultCosineThreshold() float32 {
	v := global.LRAG_CONFIG.Rag.DefaultCosineThreshold
	if v <= 0 || v > 1.0 {
		return 0
	}
	return float32(v)
}

// EffectiveConversationHistoryLimit 对话/Agent 加载服务端历史消息的条数上限
func EffectiveConversationHistoryLimit() int {
	c := global.LRAG_CONFIG.Rag.ConversationHistoryMaxMessages
	if c <= 0 {
		return DefaultConversationHistoryMaxMessages
	}
	if c > constMaxConversationHistoryMessages {
		return constMaxConversationHistoryMessages
	}
	if c < 2 {
		return 2
	}
	return c
}
