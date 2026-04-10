package rag

import (
	"context"
	"strconv"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
)

// CleanupKnowledgeGraphLinksForDocument 在删除或重建文档切片前移除 chunk 与图谱的关联，并清理孤立实体/关系及其向量
func CleanupKnowledgeGraphLinksForDocument(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, documentID uint) {
	db := global.LRAG_DB.WithContext(ctx)
	var chunkIDs []uint
	if err := db.Model(&rag.RagChunk{}).Where("document_id = ?", documentID).Pluck("id", &chunkIDs).Error; err != nil || len(chunkIDs) == 0 {
		return
	}
	_ = db.Where("chunk_id IN ?", chunkIDs).Delete(&rag.RagKgEntityChunk{}).Error
	_ = db.Where("chunk_id IN ?", chunkIDs).Delete(&rag.RagKgRelationshipChunk{}).Error
	purgeOrphanKnowledgeGraph(ctx, kb, userID)
}

// DeleteKnowledgeGraphForKnowledgeBase 删除知识库时清理图谱表及向量库中的 kg 命名空间
func DeleteKnowledgeGraphForKnowledgeBase(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) {
	store, err := createVectorStoreForDocDelete(ctx, kb, userID)
	if err == nil && store != nil {
		ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
		_ = store.DeleteByNamespace(ctx, ns+"_kg_entity")
		_ = store.DeleteByNamespace(ctx, ns+"_kg_rel")
	}
	db := global.LRAG_DB.WithContext(ctx)
	kbID := kb.ID
	var relIDs []uint
	_ = db.Model(&rag.RagKgRelationship{}).Where("knowledge_base_id = ?", kbID).Pluck("id", &relIDs).Error
	if len(relIDs) > 0 {
		_ = db.Where("relationship_id IN ?", relIDs).Delete(&rag.RagKgRelationshipChunk{}).Error
	}
	_ = db.Where("knowledge_base_id = ?", kbID).Delete(&rag.RagKgRelationship{}).Error
	var entIDs []uint
	_ = db.Model(&rag.RagKgEntity{}).Where("knowledge_base_id = ?", kbID).Pluck("id", &entIDs).Error
	if len(entIDs) > 0 {
		_ = db.Where("entity_id IN ?", entIDs).Delete(&rag.RagKgEntityChunk{}).Error
	}
	_ = db.Where("knowledge_base_id = ?", kbID).Delete(&rag.RagKgEntity{}).Error
	InvalidateKgEntityPresenceCache(kbID)
}

func purgeOrphanKnowledgeGraph(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) {
	db := global.LRAG_DB.WithContext(ctx)
	kbID := kb.ID

	var rels []rag.RagKgRelationship
	_ = db.Where("knowledge_base_id = ?", kbID).Find(&rels).Error
	for _, r := range rels {
		var n int64
		_ = db.Model(&rag.RagKgRelationshipChunk{}).Where("relationship_id = ?", r.ID).Count(&n).Error
		if n != 0 {
			continue
		}
		deleteRelVectorsOne(ctx, kb, userID, &r)
		_ = db.Delete(&r).Error
	}

	var ents []rag.RagKgEntity
	_ = db.Where("knowledge_base_id = ?", kbID).Find(&ents).Error
	for _, e := range ents {
		var n int64
		_ = db.Model(&rag.RagKgEntityChunk{}).Where("entity_id = ?", e.ID).Count(&n).Error
		if n != 0 {
			continue
		}
		deleteEntityVectorsOne(ctx, kb, userID, &e)
		_ = db.Delete(&e).Error
	}
	InvalidateKgEntityPresenceCache(kbID)
}

func deleteEntityVectorsOne(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, ent *rag.RagKgEntity) {
	if ent == nil || ent.VectorStoreID == "" {
		return
	}
	store, err := createVectorStoreForDocDelete(ctx, kb, userID)
	if err != nil || store == nil {
		return
	}
	_ = store.DeleteByIDs(ctx, []string{ent.VectorStoreID})
}

func deleteRelVectorsOne(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, rel *rag.RagKgRelationship) {
	if rel == nil || rel.VectorStoreID == "" {
		return
	}
	store, err := createVectorStoreForDocDelete(ctx, kb, userID)
	if err != nil || store == nil {
		return
	}
	_ = store.DeleteByIDs(ctx, []string{rel.VectorStoreID})
}
