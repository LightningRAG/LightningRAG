package rag

import (
	"context"
	"fmt"
	"math"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
)

const maxBatchDocuments = 200

func normalizeBatchDocumentIDs(ids []uint) []uint {
	seen := make(map[uint]struct{})
	var out []uint
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
	if len(out) > maxBatchDocuments {
		out = out[:maxBatchDocuments]
	}
	return out
}

func (s *KnowledgeBaseService) assertDocumentsInKnowledgeBase(ctx context.Context, kbID uint, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("请选择文档")
	}
	var n int64
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("knowledge_base_id = ? AND id IN ?", kbID, ids).
		Count(&n).Error; err != nil {
		return err
	}
	if int(n) != len(ids) {
		return fmt.Errorf("部分文档不存在或不属于该知识库")
	}
	return nil
}

func assertKBOwnedByUser(ctx context.Context, kbID, uid uint) error {
	var kb rag.RagKnowledgeBase
	return global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", kbID, uid).First(&kb).Error
}

// BatchDeleteDocuments 批量删除文档
func (s *KnowledgeBaseService) BatchDeleteDocuments(ctx context.Context, uid uint, req request.DocumentBatchByIDs) error {
	ids := normalizeBatchDocumentIDs(req.DocumentIDs)
	if err := assertKBOwnedByUser(ctx, req.KnowledgeBaseID, uid); err != nil {
		return err
	}
	if err := s.assertDocumentsInKnowledgeBase(ctx, req.KnowledgeBaseID, ids); err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.DeleteDocument(ctx, uid, id); err != nil {
			return err
		}
	}
	return nil
}

// BatchReindexDocuments 批量重新切片（失败/已完成/已取消且已落盘）
func (s *KnowledgeBaseService) BatchReindexDocuments(ctx context.Context, uid uint, req request.DocumentBatchByIDs) error {
	ids := normalizeBatchDocumentIDs(req.DocumentIDs)
	if err := assertKBOwnedByUser(ctx, req.KnowledgeBaseID, uid); err != nil {
		return err
	}
	if err := s.assertDocumentsInKnowledgeBase(ctx, req.KnowledgeBaseID, ids); err != nil {
		return err
	}
	for _, id := range ids {
		var doc rag.RagDocument
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND knowledge_base_id = ?", id, req.KnowledgeBaseID).First(&doc).Error; err != nil {
			continue
		}
		if doc.Status == "processing" || doc.StoragePath == "" {
			continue
		}
		if doc.Status != "failed" && doc.Status != "completed" && doc.Status != "cancelled" {
			continue
		}
		global.LRAG_DB.WithContext(ctx).Model(&doc).Updates(map[string]any{"status": "processing", "error_msg": ""})
		EnqueueDocumentIndexing(id, uid)
	}
	return nil
}

// BatchCancelDocumentIndexing 批量取消进行中的切片任务（尽力而为：已进入向量化阶段可能仍会跑完）
func (s *KnowledgeBaseService) BatchCancelDocumentIndexing(ctx context.Context, uid uint, req request.DocumentBatchByIDs) error {
	ids := normalizeBatchDocumentIDs(req.DocumentIDs)
	if err := assertKBOwnedByUser(ctx, req.KnowledgeBaseID, uid); err != nil {
		return err
	}
	if err := s.assertDocumentsInKnowledgeBase(ctx, req.KnowledgeBaseID, ids); err != nil {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("knowledge_base_id = ? AND id IN ? AND status = ?", req.KnowledgeBaseID, ids, "processing").
		Updates(map[string]any{
			"status":    "cancelled",
			"error_msg": "用户取消",
		}).Error
}

// BatchSetDocumentRetrieval 批量启用/禁用文档参与 RAG 检索
func (s *KnowledgeBaseService) BatchSetDocumentRetrieval(ctx context.Context, uid uint, req request.DocumentBatchRetrieval) error {
	ids := normalizeBatchDocumentIDs(req.DocumentIDs)
	if err := assertKBOwnedByUser(ctx, req.KnowledgeBaseID, uid); err != nil {
		return err
	}
	if err := s.assertDocumentsInKnowledgeBase(ctx, req.KnowledgeBaseID, ids); err != nil {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("knowledge_base_id = ? AND id IN ?", req.KnowledgeBaseID, ids).
		Update("retrieval_enabled", req.Enabled).Error
}

// BatchSetDocumentPriority 批量设置文档 priority；对已入库向量且状态为 completed 的文档触发异步重索引以刷新 rank_boost
func (s *KnowledgeBaseService) BatchSetDocumentPriority(ctx context.Context, uid uint, req request.DocumentBatchPriority) error {
	ids := normalizeBatchDocumentIDs(req.DocumentIDs)
	if err := assertKBOwnedByUser(ctx, req.KnowledgeBaseID, uid); err != nil {
		return err
	}
	if err := s.assertDocumentsInKnowledgeBase(ctx, req.KnowledgeBaseID, ids); err != nil {
		return err
	}
	p := req.Priority
	if math.IsNaN(p) || p < 0 || p > 1 {
		return fmt.Errorf("priority 须在 0~1 之间")
	}
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("knowledge_base_id = ? AND id IN ?", req.KnowledgeBaseID, ids).
		Update("priority", p).Error; err != nil {
		return err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).First(&kb, req.KnowledgeBaseID).Error; err != nil {
		return err
	}
	for _, id := range ids {
		var doc rag.RagDocument
		if err := global.LRAG_DB.WithContext(ctx).Where("knowledge_base_id = ? AND id = ?", req.KnowledgeBaseID, id).First(&doc).Error; err != nil {
			continue
		}
		if doc.Status != "completed" {
			continue
		}
		var n int64
		if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChunk{}).
			Where("document_id = ? AND vector_store_id != ?", id, "").
			Count(&n).Error; err != nil || n == 0 {
			continue
		}
		go s.reindexDocumentChunks(doc, kb, uid)
	}
	return nil
}
