package rag

import (
	"strings"
	"testing"
)

func TestFormatKnowledgeGraphPromptPrefixBudget(t *testing.T) {
	ents := []map[string]any{
		{"entity_name": "A", "entity_type": "T", "entity_description": strings.Repeat("word ", 2000)},
	}
	s := FormatKnowledgeGraphPromptPrefix(ents, nil, 80, 0)
	if s == "" {
		t.Fatal("expected some output")
	}
	if !strings.Contains(s, "Knowledge Graph Data (Entity):") || !strings.Contains(s, "```json") {
		t.Fatalf("expected LightRAG-style KG headers + json fence, got: %q", s)
	}
	if estimateTokens(s) > 240 {
		t.Fatalf("expected bounded output, got ~%d tokens", estimateTokens(s))
	}
}

func TestFormatKnowledgeGraphPromptPrefixRelationshipJSONL(t *testing.T) {
	rels := []map[string]any{
		{"src_entity": "A", "tgt_entity": "B", "relationship_keywords": "k", "relationship_description": "d"},
	}
	s := FormatKnowledgeGraphPromptPrefix(nil, rels, 0, 120)
	if !strings.Contains(s, "Knowledge Graph Data (Relationship):") {
		t.Fatal(s)
	}
	if !strings.Contains(s, `"src_entity":"A"`) {
		t.Fatal(s)
	}
}
