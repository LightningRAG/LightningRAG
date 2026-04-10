package rag

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
)

func TestApplyDefaultConversationRetrievePoolIfNeeded(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.DefaultConversationRetrievePoolTopK
	oldCand := global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK
	t.Cleanup(func() {
		global.LRAG_CONFIG.Rag.DefaultConversationRetrievePoolTopK = old
		global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK = oldCand
	})

	s := RetrieverSessionFromLightningRAGParams("vector", nil, nil, nil, nil)
	ApplyDefaultConversationRetrievePoolIfNeeded(&s, 6)
	if s.RetrievePoolTopK == nil || *s.RetrievePoolTopK != lightragconst.MaxSimilarityFetchK {
		t.Fatalf("config 0 -> Ragflow wide pool %d, got %#v", lightragconst.MaxSimilarityFetchK, s.RetrievePoolTopK)
	}
	global.LRAG_CONFIG.Rag.DefaultConversationRetrievePoolTopK = 20
	s2 := RetrieverSessionFromLightningRAGParams("vector", nil, nil, nil, nil)
	ApplyDefaultConversationRetrievePoolIfNeeded(&s2, 6)
	if s2.RetrievePoolTopK == nil || *s2.RetrievePoolTopK != 20 {
		t.Fatalf("expected pool 20, got %#v", s2.RetrievePoolTopK)
	}
	s3 := RetrieverSessionFromLightningRAGParams("vector", intPtr(30), nil, nil, nil)
	ApplyDefaultConversationRetrievePoolIfNeeded(&s3, 6)
	if *s3.RetrievePoolTopK != 30 {
		t.Fatal("explicit topK should not be overwritten")
	}
}

func intPtr(n int) *int { return &n }

func TestEffectiveDefaultKnowledgeBaseRetrievePoolTopK(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK
	oldMax := global.LRAG_CONFIG.Rag.MaxRetrieveTopN
	oldCand := global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK
	t.Cleanup(func() {
		global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK = old
		global.LRAG_CONFIG.Rag.MaxRetrieveTopN = oldMax
		global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK = oldCand
	})
	global.LRAG_CONFIG.Rag.MaxRetrieveTopN = 50
	global.LRAG_CONFIG.Rag.MaxRetrieveCandidateTopK = 0
	global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK = 0
	if got := EffectiveDefaultKnowledgeBaseRetrievePoolTopK(); got != lightragconst.MaxSimilarityFetchK {
		t.Fatalf("0 -> wide pool: want %d got %d", lightragconst.MaxSimilarityFetchK, got)
	}
	global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK = 25
	if EffectiveDefaultKnowledgeBaseRetrievePoolTopK() != 25 {
		t.Fatalf("got %d", EffectiveDefaultKnowledgeBaseRetrievePoolTopK())
	}
	global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK = 100
	if EffectiveDefaultKnowledgeBaseRetrievePoolTopK() != 100 {
		t.Fatalf("expected 100 (no longer capped by max-retrieve-top-n), got %d", EffectiveDefaultKnowledgeBaseRetrievePoolTopK())
	}
	global.LRAG_CONFIG.Rag.DefaultKnowledgeBaseRetrievePoolTopK = 99999
	if got := EffectiveDefaultKnowledgeBaseRetrievePoolTopK(); got != lightragconst.MaxSimilarityFetchK {
		t.Fatalf("expected cap at MaxSimilarityFetchK, got %d", got)
	}
}

func TestRetrieverSessionFromLightningRAGParamsDefaultCosine(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.DefaultCosineThreshold
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.DefaultCosineThreshold = old })

	global.LRAG_CONFIG.Rag.DefaultCosineThreshold = 0.2
	s := RetrieverSessionFromLightningRAGParams("vector", nil, nil, nil, nil)
	if s.CosineThreshold == nil || *s.CosineThreshold < 0.19 || *s.CosineThreshold > 0.21 {
		t.Fatalf("expected default ~0.2, got %#v", s.CosineThreshold)
	}
	ct := float32(0.55)
	s2 := RetrieverSessionFromLightningRAGParams("vector", nil, nil, &ct, nil)
	if s2.CosineThreshold == nil || *s2.CosineThreshold != 0.55 {
		t.Fatalf("request cosine should win, got %#v", s2.CosineThreshold)
	}
}
