package rag

import (
	"strings"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
)

func TestKgGleaningFitsInputBudget(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.KgExtractGleaningMaxInputTokens
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.KgExtractGleaningMaxInputTokens = old })

	global.LRAG_CONFIG.Rag.KgExtractGleaningMaxInputTokens = 100000
	if !kgGleaningFitsInputBudget("hello", "world") {
		t.Fatal("small payload should fit")
	}
	global.LRAG_CONFIG.Rag.KgExtractGleaningMaxInputTokens = 800
	huge := strings.Repeat("x", 20000)
	if kgGleaningFitsInputBudget(huge, huge) {
		t.Fatal("huge payload should exceed low token limit")
	}
}

func TestParseKgExtractJSON(t *testing.T) {
	raw := "```json\n{\"per_chunk\":[{\"chunk_index\":0,\"entities\":[{\"name\":\"Acme\",\"type\":\"Organization\",\"description\":\"A company\"}],\"relationships\":[]}]}\n```"
	res, err := parseKgExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.PerChunk) != 1 || len(res.PerChunk[0].Entities) != 1 {
		t.Fatalf("unexpected: %+v", res)
	}
	if res.PerChunk[0].Entities[0].Name != "Acme" {
		t.Fatalf("name: %q", res.PerChunk[0].Entities[0].Name)
	}
}
