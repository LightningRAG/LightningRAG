package rag

import (
	"context"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestKnowledgeGraphMapsIncludesRelationshipEndpoints(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&rag.RagKgEntity{}, &rag.RagKgRelationship{}, &rag.RagKgEntityChunk{}, &rag.RagKgRelationshipChunk{}); err != nil {
		t.Fatal(err)
	}
	prevDB := global.LRAG_DB
	prevLim := global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit
	t.Cleanup(func() {
		global.LRAG_DB = prevDB
		global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit = prevLim
	})
	global.LRAG_DB = db
	global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit = 0

	const kb uint = 1
	a := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Alpha", NormalizedName: "alpha", EntityType: "T", Description: "da"}
	b := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Beta", NormalizedName: "beta", EntityType: "T", Description: "db"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	rel := rag.RagKgRelationship{
		KnowledgeBaseID: kb, SourceEntityID: a.ID, TargetEntityID: b.ID,
		Keywords: "k", Description: "d",
	}
	if err := db.Create(&rel).Error; err != nil {
		t.Fatal(err)
	}
	const chunkID uint = 42
	if err := db.Create(&rag.RagKgRelationshipChunk{RelationshipID: rel.ID, ChunkID: chunkID}).Error; err != nil {
		t.Fatal(err)
	}

	ents, rels := KnowledgeGraphMapsForChunkIDs(context.Background(), []uint{kb}, []uint{chunkID})
	if len(ents) != 2 {
		t.Fatalf("entities: want 2 (endpoints), got %d", len(ents))
	}
	if len(rels) != 1 {
		t.Fatalf("relationships: want 1, got %d", len(rels))
	}
}

func TestKnowledgeGraphMapsNeighborRelLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&rag.RagKgEntity{}, &rag.RagKgRelationship{}, &rag.RagKgEntityChunk{}, &rag.RagKgRelationshipChunk{}); err != nil {
		t.Fatal(err)
	}
	prevDB := global.LRAG_DB
	prevLim := global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit
	t.Cleanup(func() {
		global.LRAG_DB = prevDB
		global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit = prevLim
	})
	global.LRAG_DB = db
	global.LRAG_CONFIG.Rag.KgPromptNeighborRelLimit = 10

	const kb uint = 1
	alpha := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Alpha", NormalizedName: "alpha", EntityType: "T", Description: ""}
	beta := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Beta", NormalizedName: "beta", EntityType: "T", Description: ""}
	gamma := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Gamma", NormalizedName: "gamma", EntityType: "T", Description: ""}
	for _, e := range []*rag.RagKgEntity{&alpha, &beta, &gamma} {
		if err := db.Create(e).Error; err != nil {
			t.Fatal(err)
		}
	}
	rAB := rag.RagKgRelationship{KnowledgeBaseID: kb, SourceEntityID: alpha.ID, TargetEntityID: beta.ID, Keywords: "ab", Description: ""}
	rBC := rag.RagKgRelationship{KnowledgeBaseID: kb, SourceEntityID: beta.ID, TargetEntityID: gamma.ID, Keywords: "bc", Description: ""}
	if err := db.Create(&rAB).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&rBC).Error; err != nil {
		t.Fatal(err)
	}
	const chunkID uint = 7
	if err := db.Create(&rag.RagKgRelationshipChunk{RelationshipID: rAB.ID, ChunkID: chunkID}).Error; err != nil {
		t.Fatal(err)
	}

	ents, rels := KnowledgeGraphMapsForChunkIDs(context.Background(), []uint{kb}, []uint{chunkID})
	if len(ents) != 3 {
		t.Fatalf("want 3 entities (A,B,C), got %d", len(ents))
	}
	if len(rels) != 2 {
		t.Fatalf("want 2 relationships (AB + neighbor BC), got %d", len(rels))
	}
}
