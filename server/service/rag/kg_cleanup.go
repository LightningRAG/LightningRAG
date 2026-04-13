package rag

import (
	"context"
	"strconv"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"go.uber.org/zap"
)

// CleanupKnowledgeGraphLinksForDocument 在删除或重建文档切片前移除 chunk 与图谱的关联，并清理孤立实体/关系及其向量
func CleanupKnowledgeGraphLinksForDocument(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, documentID uint) {
	db := global.LRAG_DB.WithContext(ctx)
	var chunkIDs []uint
	if err := db.Model(&rag.RagChunk{}).Where("document_id = ?", documentID).Pluck("id", &chunkIDs).Error; err != nil || len(chunkIDs) == 0 {
		return
	}
	if err := db.Where("chunk_id IN ?", chunkIDs).Delete(&rag.RagKgEntityChunk{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg entity chunk links", zap.Uint("documentId", documentID), zap.Error(err))
	}
	if err := db.Where("chunk_id IN ?", chunkIDs).Delete(&rag.RagKgRelationshipChunk{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg relationship chunk links", zap.Uint("documentId", documentID), zap.Error(err))
	}
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
	if err := db.Model(&rag.RagKgRelationship{}).Where("knowledge_base_id = ?", kbID).Pluck("id", &relIDs).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg relationships", zap.Uint("knowledgeBaseId", kbID), zap.Error(err))
	}
	if len(relIDs) > 0 {
		if err := db.Where("relationship_id IN ?", relIDs).Delete(&rag.RagKgRelationshipChunk{}).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg relationship chunk links", zap.Uint("knowledgeBaseId", kbID), zap.Error(err))
		}
	}
	if err := db.Where("knowledge_base_id = ?", kbID).Delete(&rag.RagKgRelationship{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg relationships", zap.Uint("knowledgeBaseId", kbID), zap.Error(err))
	}
	var entIDs []uint
	if err := db.Model(&rag.RagKgEntity{}).Where("knowledge_base_id = ?", kbID).Pluck("id", &entIDs).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg entities", zap.Uint("knowledgeBaseId", kbID), zap.Error(err))
	}
	if len(entIDs) > 0 {
		if err := db.Where("entity_id IN ?", entIDs).Delete(&rag.RagKgEntityChunk{}).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg entity chunk links", zap.Uint("knowledgeBaseId", kbID), zap.Error(err))
		}
	}
	if err := db.Where("knowledge_base_id = ?", kbID).Delete(&rag.RagKgEntity{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg entities", zap.Uint("knowledgeBaseId", kbID), zap.Error(err))
	}
	InvalidateKgEntityPresenceCache(kbID)
}

func purgeOrphanKnowledgeGraph(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) {
	db := global.LRAG_DB.WithContext(ctx)
	kbID := kb.ID

	var rels []rag.RagKgRelationship
	if err := db.Where("knowledge_base_id = ?", kbID).Find(&rels).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg relationships", zap.Uint("knowledgeBaseId", kbID), zap.String("phase", "list_for_orphan_purge"), zap.Error(err))
	}
	for _, r := range rels {
		var n int64
		if err := db.Model(&rag.RagKgRelationshipChunk{}).Where("relationship_id = ?", r.ID).Count(&n).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg relationship", zap.Uint("knowledgeBaseId", kbID), zap.Uint("relationshipId", r.ID), zap.String("phase", "count_chunks"), zap.Error(err))
			continue
		}
		if n != 0 {
			continue
		}
		deleteRelVectorsOne(ctx, kb, userID, &r)
		if err := db.Delete(&r).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg relationship", zap.Uint("knowledgeBaseId", kbID), zap.Uint("relationshipId", r.ID), zap.String("phase", "delete_orphan"), zap.Error(err))
		}
	}

	var ents []rag.RagKgEntity
	if err := db.Where("knowledge_base_id = ?", kbID).Find(&ents).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg entities", zap.Uint("knowledgeBaseId", kbID), zap.String("phase", "list_for_orphan_purge"), zap.Error(err))
	}
	for _, e := range ents {
		var n int64
		if err := db.Model(&rag.RagKgEntityChunk{}).Where("entity_id = ?", e.ID).Count(&n).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg entity", zap.Uint("knowledgeBaseId", kbID), zap.Uint("entityId", e.ID), zap.String("phase", "count_chunks"), zap.Error(err))
			continue
		}
		if n != 0 {
			continue
		}
		deleteEntityVectorsOne(ctx, kb, userID, &e)
		if err := db.Delete(&e).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg entity", zap.Uint("knowledgeBaseId", kbID), zap.Uint("entityId", e.ID), zap.String("phase", "delete_orphan"), zap.Error(err))
		}
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
