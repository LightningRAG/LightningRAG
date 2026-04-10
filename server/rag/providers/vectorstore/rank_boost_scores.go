package vectorstore

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func clampRankBoost01(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// RankBoostFromChunkMetadata 读取 metadata.rank_boost 并钳制到 [0,1]
func RankBoostFromChunkMetadata(m map[string]any) float32 {
	if m == nil {
		return 0
	}
	v, ok := m["rank_boost"]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return clampRankBoost01(float32(t))
	case float32:
		return clampRankBoost01(t)
	case int:
		return clampRankBoost01(float32(t))
	case int64:
		return clampRankBoost01(float32(t))
	case uint:
		return clampRankBoost01(float32(t))
	case uint64:
		return clampRankBoost01(float32(t))
	default:
		return 0
	}
}

// ApplyConfiguredRankBoostToScores 检索后对 score 做乘性加成：score *= 1 + weight*rank_boost。
// 配置项 elasticsearch-score-rank-boost-weight；metadata.rank_boost 由索引入库时写入（见 default-chunk-rank-boost）。
func ApplyConfiguredRankBoostToScores(docs []schema.Document) {
	w := global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight
	if w <= 0 || len(docs) == 0 {
		return
	}
	wf := float32(w)
	if wf > 50 {
		wf = 50
	}
	for i := range docs {
		rb := RankBoostFromChunkMetadata(docs[i].Metadata)
		docs[i].Score = docs[i].Score * (1 + wf*rb)
	}
}
