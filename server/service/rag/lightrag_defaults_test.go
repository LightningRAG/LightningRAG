package rag

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
)

func TestClampRetrieveTopN(t *testing.T) {
	if got := ClampRetrieveTopN(0, 8); got != 8 {
		t.Fatalf("zero -> default: got %d", got)
	}
	if got := ClampRetrieveTopN(-1, 8); got != 8 {
		t.Fatalf("negative -> default: got %d", got)
	}
	if got := ClampRetrieveTopN(100, 8); got != MaxRetrieveTopN {
		t.Fatalf("cap max: got %d", got)
	}
	if got := ClampRetrieveTopN(3, 8); got != 3 {
		t.Fatalf("in range: got %d", got)
	}
}

func TestResolveRetrieveCandidateCount(t *testing.T) {
	if got := resolveRetrieveCandidateCount(6, nil); got != lightragconst.DefaultTopK {
		t.Fatalf("no pool: want %d got %d", lightragconst.DefaultTopK, got)
	}
	small := 4
	if got := resolveRetrieveCandidateCount(6, &small); got != 6 {
		t.Fatalf("pool smaller than final: got %d", got)
	}
	large := 30
	if got := resolveRetrieveCandidateCount(6, &large); got != 30 {
		t.Fatalf("pool larger: want 30 got %d", got)
	}
}

func TestEffectiveHybridFusionWeightsDefault(t *testing.T) {
	tw, vw := EffectiveHybridFusionWeights()
	if tw < 0.049 || tw > 0.051 || vw < 0.949 || vw > 0.951 {
		t.Fatalf("tw=%v vw=%v", tw, vw)
	}
}

func TestEffectiveElasticsearchScoreRankBoostWeight(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight = prev })
	global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight = 0
	if EffectiveElasticsearchScoreRankBoostWeight() != 0 {
		t.Fatal()
	}
	global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight = 2.5
	if got := EffectiveElasticsearchScoreRankBoostWeight(); got < 2.49 || got > 2.51 {
		t.Fatalf("%v", got)
	}
	global.LRAG_CONFIG.Rag.ElasticsearchScoreRankBoostWeight = 100
	if got := EffectiveElasticsearchScoreRankBoostWeight(); got != 50 {
		t.Fatalf("cap: %v", got)
	}
}

func TestEffectiveKeywordEmptyRetry(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.KeywordSkipEmptyRetry
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.KeywordSkipEmptyRetry = prev })
	if !EffectiveKeywordEmptyRetry() {
		t.Fatal("default should retry")
	}
	global.LRAG_CONFIG.Rag.KeywordSkipEmptyRetry = true
	if EffectiveKeywordEmptyRetry() {
		t.Fatal("skip should disable")
	}
}

func TestEffectiveVectorEmptyRetry(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry = prev })
	if !EffectiveVectorEmptyRetry() {
		t.Fatal("default should retry")
	}
	global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry = true
	if EffectiveVectorEmptyRetry() {
		t.Fatal("skip should disable")
	}
}

func TestEffectiveHybridFusionMinScore(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.HybridFusionMinScore
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.HybridFusionMinScore = prev })
	global.LRAG_CONFIG.Rag.HybridFusionMinScore = 0.2
	if got := EffectiveHybridFusionMinScore(); got < 0.199 || got > 0.201 {
		t.Fatalf("got %v", got)
	}
	global.LRAG_CONFIG.Rag.HybridFusionMinScore = 0
	if EffectiveHybridFusionMinScore() != 0 {
		t.Fatal()
	}
}

func TestResolveRetrieveCandidateCountCapsAtMaxCandidate(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK = old })
	global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK = 80
	huge := 5000
	if got := resolveRetrieveCandidateCount(6, &huge); got != 80 {
		t.Fatalf("pool should cap at max candidate: got %d", got)
	}
}

func TestEffectiveMaxEntityContextTokensExplicit(t *testing.T) {
	u := uint(42)
	if got := EffectiveMaxEntityContextTokens(&u); got != 42 {
		t.Fatalf("explicit: got %d", got)
	}
	if got := EffectiveMaxEntityContextTokens(nil); got != 0 {
		t.Fatalf("nil with zero config: got %d", got)
	}
}

func TestEffectiveConversationHistoryLimit(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.ConversationHistoryMaxMessages
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.ConversationHistoryMaxMessages = prev })
	if EffectiveConversationHistoryLimit() != DefaultConversationHistoryMaxMessages {
		t.Fatalf("default: %d", EffectiveConversationHistoryLimit())
	}
	global.LRAG_CONFIG.Rag.ConversationHistoryMaxMessages = 50
	if EffectiveConversationHistoryLimit() != 50 {
		t.Fatal()
	}
	global.LRAG_CONFIG.Rag.ConversationHistoryMaxMessages = 99999
	if EffectiveConversationHistoryLimit() != constMaxConversationHistoryMessages {
		t.Fatalf("cap: %d", EffectiveConversationHistoryLimit())
	}
}

func TestEffectiveMaxRagContextTokens(t *testing.T) {
	prev := global.LRAG_CONFIG.Rag.DefaultMaxRagContextTokens
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.DefaultMaxRagContextTokens = prev })
	ex := uint(3000)
	if got := EffectiveMaxRagContextTokens(&ex); got != 3000 {
		t.Fatalf("explicit: got %d", got)
	}
	global.LRAG_CONFIG.Rag.DefaultMaxRagContextTokens = 8000
	if got := EffectiveMaxRagContextTokens(nil); got != 8000 {
		t.Fatalf("config default: got %d", got)
	}
	z := uint(0)
	if got := EffectiveMaxRagContextTokens(&z); got != 8000 {
		t.Fatalf("explicit zero falls back to config: got %d", got)
	}
}
