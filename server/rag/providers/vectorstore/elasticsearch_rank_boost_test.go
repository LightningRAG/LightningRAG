package vectorstore

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func TestApplyElasticsearchRankBoost(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight = prev })
	global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight = 1
	docs := []schema.Document{{Score: 10, Metadata: map[string]any{"rank_boost": 0.5}}}
	ApplyConfiguredRankBoostToScores(docs)
	// 10 * (1 + 1*0.5) = 15
	if docs[0].Score < 14.99 || docs[0].Score > 15.01 {
		t.Fatalf("got %v", docs[0].Score)
	}
}
