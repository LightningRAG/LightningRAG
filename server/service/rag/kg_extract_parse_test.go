package rag

import (
	"testing"
)

func TestParseKgExtractJSONTupleRelationships(t *testing.T) {
	raw := `{"per_chunk":[{"chunk_index":1,"entities":[["Acme","Organization","company"]],"relationships":[["Alice","Acme","works_for","employee"]]}]}`
	res, err := parseKgExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.PerChunk) != 1 {
		t.Fatalf("chunks %d", len(res.PerChunk))
	}
	pc := res.PerChunk[0]
	if pc.ChunkIndex != 1 || len(pc.Entities) != 1 || pc.Entities[0].Name != "Acme" {
		t.Fatalf("entity: %+v", pc.Entities)
	}
	if len(pc.Relationships) != 1 || pc.Relationships[0].Source != "Alice" || pc.Relationships[0].Target != "Acme" {
		t.Fatalf("rel: %+v", pc.Relationships)
	}
	if pc.Relationships[0].Keywords != "works_for" || pc.Relationships[0].Description != "employee" {
		t.Fatalf("rel fields: %+v", pc.Relationships[0])
	}
}

func TestParseKgExtractJSONLegacyObjectForm(t *testing.T) {
	raw := `{"per_chunk":[{"chunk_index":0,"entities":[{"name":"X","type":"P","description":"d"}],"relationships":[{"source":"A","target":"B","keywords":"k","description":"r"}]}]}`
	res, err := parseKgExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.PerChunk) != 1 || len(res.PerChunk[0].Relationships) != 1 {
		t.Fatalf("%+v", res.PerChunk)
	}
}
