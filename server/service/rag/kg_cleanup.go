package rag

import (
	"context"
	"strconv"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"go.uber.org/zap"
)

func dedupeUints(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// CleanupKnowledgeGraphLinksForDocument 在删除或重建文档切片前移除 chunk 与图谱的关联，并清理孤立实体/关系及其向量
func CleanupKnowledgeGraphLinksForDocument(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, documentID uint) {
	db := global.LRAG_DB.WithContext(ctx)
	var chunkIDs []uint
	if err := db.Model(&rag.RagChunk{}).Where("document_id = ?", documentID).Pluck("id", &chunkIDs).Error; err != nil || len(chunkIDs) == 0 {
		return
	}
	var candidateEntIDs []uint
	if err := db.Model(&rag.RagKgEntityChunk{}).Distinct("entity_id").Where("chunk_id IN ?", chunkIDs).Pluck("entity_id", &candidateEntIDs).Error; err != nil {
		global.LRAG_LOG.Warn("failed to list kg entity candidates for document cleanup", zap.Uint("documentId", documentID), zap.Error(err))
	}
	var candidateRelIDs []uint
	if err := db.Model(&rag.RagKgRelationshipChunk{}).Distinct("relationship_id").Where("chunk_id IN ?", chunkIDs).Pluck("relationship_id", &candidateRelIDs).Error; err != nil {
		global.LRAG_LOG.Warn("failed to list kg relationship candidates for document cleanup", zap.Uint("documentId", documentID), zap.Error(err))
	}
	if err := db.Where("chunk_id IN ?", chunkIDs).Delete(&rag.RagKgEntityChunk{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg entity chunk links", zap.Uint("documentId", documentID), zap.Error(err))
	}
	if err := db.Where("chunk_id IN ?", chunkIDs).Delete(&rag.RagKgRelationshipChunk{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to cleanup kg relationship chunk links", zap.Uint("documentId", documentID), zap.Error(err))
	}
	purgeOrphanKnowledgeGraphCandidates(ctx, kb, userID, dedupeUints(candidateRelIDs), dedupeUints(candidateEntIDs))
}

// DeleteKnowledgeGraphForKnowledgeBase 删除知识库时清理图谱表及向量库中的 kg 命名空间
func DeleteKnowledgeGraphForKnowledgeBase(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) {
	withVectorStoreForDeleteOps(ctx, kb, userID, func(opCtx context.Context, store interfaces.VectorStore) {
		ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
		_ = store.DeleteByNamespace(opCtx, ns+"_kg_entity")
		_ = store.DeleteByNamespace(opCtx, ns+"_kg_rel")
	})
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

// purgeOrphanKnowledgeGraphCandidates 仅检查「曾与本次删除的 chunk 有关联」的实体/关系是否变为孤立；
// 向量删除复用同一 VectorStore，避免对每个孤立点重复初始化嵌入与向量客户端（大图谱下原先会极慢并易触发请求超时）。
func purgeOrphanKnowledgeGraphCandidates(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, candidateRelIDs, candidateEntIDs []uint) {
	if len(candidateRelIDs) == 0 && len(candidateEntIDs) == 0 {
		return
	}
	db := global.LRAG_DB.WithContext(ctx)
	kbID := kb.ID

	store, vctx, vcancel := openVectorStoreForDeleteOrNil(ctx, kb, userID)
	defer vcancel()

	for _, rid := range candidateRelIDs {
		var n int64
		if err := db.Model(&rag.RagKgRelationshipChunk{}).Where("relationship_id = ?", rid).Count(&n).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg relationship", zap.Uint("knowledgeBaseId", kbID), zap.Uint("relationshipId", rid), zap.String("phase", "count_chunks"), zap.Error(err))
			continue
		}
		if n != 0 {
			continue
		}
		var r rag.RagKgRelationship
		if err := db.Where("id = ? AND knowledge_base_id = ?", rid, kbID).First(&r).Error; err != nil {
			continue
		}
		if r.VectorStoreID != "" && store != nil {
			if err := store.DeleteByIDs(vctx, []string{r.VectorStoreID}); err != nil {
				global.LRAG_LOG.Warn("kg orphan: 删除关系向量失败（已跳过）", zap.Uint("relationshipId", rid), zap.Error(err))
			}
		}
		if err := db.Delete(&r).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg relationship", zap.Uint("knowledgeBaseId", kbID), zap.Uint("relationshipId", rid), zap.String("phase", "delete_orphan"), zap.Error(err))
		}
	}

	for _, eid := range candidateEntIDs {
		var n int64
		if err := db.Model(&rag.RagKgEntityChunk{}).Where("entity_id = ?", eid).Count(&n).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg entity", zap.Uint("knowledgeBaseId", kbID), zap.Uint("entityId", eid), zap.String("phase", "count_chunks"), zap.Error(err))
			continue
		}
		if n != 0 {
			continue
		}
		var e rag.RagKgEntity
		if err := db.Where("id = ? AND knowledge_base_id = ?", eid, kbID).First(&e).Error; err != nil {
			continue
		}
		if e.VectorStoreID != "" && store != nil {
			if err := store.DeleteByIDs(vctx, []string{e.VectorStoreID}); err != nil {
				global.LRAG_LOG.Warn("kg orphan: 删除实体向量失败（已跳过）", zap.Uint("entityId", eid), zap.Error(err))
			}
		}
		if err := db.Delete(&e).Error; err != nil {
			global.LRAG_LOG.Warn("failed to cleanup kg entity", zap.Uint("knowledgeBaseId", kbID), zap.Uint("entityId", eid), zap.String("phase", "delete_orphan"), zap.Error(err))
		}
	}
	InvalidateKgEntityPresenceCache(kbID)
}

