package rag

import (
	"context"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestKgUpsertRelationshipBidirectional(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&rag.RagKgEntity{}, &rag.RagKgRelationship{}); err != nil {
		t.Fatal(err)
	}
	const kb uint = 1
	a := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Alice", NormalizedName: "alice"}
	b := rag.RagKgEntity{KnowledgeBaseID: kb, Name: "Bob", NormalizedName: "bob"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatal(err)
	}

	r1 := kgLLMRelationship{Source: "Alice", Target: "Bob", Keywords: "works_with", Description: "colleagues"}
	rel1, err := kgUpsertRelationship(context.Background(), db, kb, nil, a.ID, b.ID, r1)
	if err != nil {
		t.Fatal(err)
	}
	if rel1 == nil || rel1.ID == 0 {
		t.Fatal("expected first edge")
	}

	r2 := kgLLMRelationship{Source: "Bob", Target: "Alice", Keywords: "peer", Description: "same team"}
	rel2, err := kgUpsertRelationship(context.Background(), db, kb, nil, b.ID, a.ID, r2)
	if err != nil {
		t.Fatal(err)
	}
	if rel2 == nil || rel2.ID != rel1.ID {
		t.Fatalf("reverse should reuse same row: rel1=%d rel2=%v", rel1.ID, rel2)
	}

	var count int64
	db.Model(&rag.RagKgRelationship{}).Where("knowledge_base_id = ?", kb).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 relationship row, got %d", count)
	}

	var stored rag.RagKgRelationship
	if err := db.First(&stored, rel1.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.SourceEntityID != a.ID || stored.TargetEntityID != b.ID {
		t.Fatalf("canonical direction should stay first insert: %+v", stored)
	}
	if stored.Keywords != "works_with, peer" {
		t.Fatalf("keywords merge: %q", stored.Keywords)
	}
	if stored.Description == "" {
		t.Fatal("expected merged description")
	}
}

func TestMergeKgCommaSeparatedUnique(t *testing.T) {
	got := mergeKgCommaSeparatedUnique("A, b", "a, C")
	// "a" 与 "A" 去重后只保留先出现的写法
	if got != "A, b, C" {
		t.Fatalf("got %q", got)
	}
}

func TestMergeKgDescription(t *testing.T) {
	if mergeKgDescription("x", "x") != "x" {
		t.Fatal()
	}
	if mergeKgDescription("hello world", "world") != "hello world" {
		t.Fatal()
	}
	d := mergeKgDescription("a", "b")
	if d != "a\nb" {
		t.Fatalf("got %q", d)
	}
}
