package rag

import (
	"context"
	"errors"
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"gorm.io/gorm"
)

const (
	kgVizDefaultMaxEntities      = 400
	kgVizDefaultMaxRelationships = 800
	kgVizMinLimit                = 50
	kgVizMaxEntitiesCap          = 800
	kgVizMaxRelationshipsCap     = 2000
)

func clampKgVizLimit(n, def, min, max int) int {
	if n <= 0 {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// GetKnowledgeGraphViz 返回知识库图谱子集供前端可视化（需为知识库所有者）
func (s *KnowledgeBaseService) GetKnowledgeGraphViz(ctx context.Context, uid uint, req request.KnowledgeBaseKnowledgeGraph) (*rag.KnowledgeGraphViz, error) {
	kb, err := s.Get(ctx, uid, req.ID)
	if err != nil {
		return nil, err
	}
	db := global.LRAG_DB.WithContext(ctx)
	kbID := kb.ID

	var entityCount, relCount int64
	_ = db.Model(&rag.RagKgEntity{}).Where("knowledge_base_id = ?", kbID).Count(&entityCount).Error
	_ = db.Model(&rag.RagKgRelationship{}).Where("knowledge_base_id = ?", kbID).Count(&relCount).Error

	maxEnt := clampKgVizLimit(req.MaxEntities, kgVizDefaultMaxEntities, kgVizMinLimit, kgVizMaxEntitiesCap)
	maxRel := clampKgVizLimit(req.MaxRelationships, kgVizDefaultMaxRelationships, kgVizMinLimit, kgVizMaxRelationshipsCap)

	if req.DocumentID > 0 {
		return knowledgeGraphVizForDocument(ctx, db, kb, kbID, entityCount, relCount, maxEnt, maxRel, req.DocumentID)
	}

	var entRows []rag.RagKgEntity
	if err := db.Where("knowledge_base_id = ?", kbID).Order("id ASC").Limit(maxEnt).Find(&entRows).Error; err != nil {
		return nil, err
	}
	idSet := make(map[uint]struct{}, len(entRows))
	for _, e := range entRows {
		idSet[e.ID] = struct{}{}
	}
	entIDList := make([]uint, 0, len(idSet))
	for id := range idSet {
		entIDList = append(entIDList, id)
	}

	var relRows []rag.RagKgRelationship
	if len(entIDList) > 0 {
		q := db.Where("knowledge_base_id = ?", kbID).
			Where("source_entity_id IN ? AND target_entity_id IN ?", entIDList, entIDList).
			Order("id ASC").
			Limit(maxRel)
		if err := q.Find(&relRows).Error; err != nil {
			return nil, err
		}
	}

	outEnt := kgVizEntityRowsToDTO(entRows)
	outRel := kgVizRelRowsToDTO(relRows)
	truncated := entityCount > int64(len(entRows)) || relCount > int64(len(relRows))

	return &rag.KnowledgeGraphViz{
		KnowledgeBaseID:           kbID,
		EnableKnowledgeGraph:      kb.EnableKnowledgeGraph,
		EntityCount:               entityCount,
		RelationshipCount:         relCount,
		Truncated:                 truncated,
		Entities:                  outEnt,
		Relationships:             outRel,
		MaxEntitiesRequested:      maxEnt,
		MaxRelationshipsRequested: maxRel,
	}, nil
}

func knowledgeGraphVizForDocument(ctx context.Context, db *gorm.DB, kb *rag.RagKnowledgeBase, kbID uint, entityCount, relCount int64, maxEnt, maxRel int, documentID uint) (*rag.KnowledgeGraphViz, error) {
	_ = ctx
	var doc rag.RagDocument
	if err := db.Where("id = ? AND knowledge_base_id = ?", documentID, kbID).First(&doc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("文档不存在或不属于该知识库")
		}
		return nil, err
	}

	var chunkIDs []uint
	if err := db.Model(&rag.RagChunk{}).Where("document_id = ?", documentID).Pluck("id", &chunkIDs).Error; err != nil {
		return nil, err
	}
	if len(chunkIDs) == 0 {
		return &rag.KnowledgeGraphViz{
			KnowledgeBaseID:           kbID,
			EnableKnowledgeGraph:      kb.EnableKnowledgeGraph,
			EntityCount:               entityCount,
			RelationshipCount:         relCount,
			Truncated:                 false,
			Entities:                  nil,
			Relationships:             nil,
			MaxEntitiesRequested:      maxEnt,
			MaxRelationshipsRequested: maxRel,
			ScopeDocumentID:           documentID,
		}, nil
	}

	var entCandidateIDs []uint
	if err := db.Model(&rag.RagKgEntityChunk{}).Where("chunk_id IN ?", chunkIDs).Distinct().Pluck("entity_id", &entCandidateIDs).Error; err != nil {
		return nil, err
	}
	var relCandidateIDs []uint
	if err := db.Model(&rag.RagKgRelationshipChunk{}).Where("chunk_id IN ?", chunkIDs).Distinct().Pluck("relationship_id", &relCandidateIDs).Error; err != nil {
		return nil, err
	}

	scopedEntTotal := int64(len(entCandidateIDs))
	scopedRelTotal := int64(len(relCandidateIDs))

	var entRows []rag.RagKgEntity
	if len(entCandidateIDs) > 0 {
		if err := db.Where("knowledge_base_id = ? AND id IN ?", kbID, entCandidateIDs).Order("id ASC").Limit(maxEnt).Find(&entRows).Error; err != nil {
			return nil, err
		}
	}

	idSet := make(map[uint]struct{}, len(entRows))
	for _, e := range entRows {
		idSet[e.ID] = struct{}{}
	}
	entIDList := make([]uint, 0, len(idSet))
	for id := range idSet {
		entIDList = append(entIDList, id)
	}

	var relRows []rag.RagKgRelationship
	if len(entIDList) > 0 && len(relCandidateIDs) > 0 {
		q := db.Where("knowledge_base_id = ? AND id IN ?", kbID, relCandidateIDs).
			Where("source_entity_id IN ? AND target_entity_id IN ?", entIDList, entIDList).
			Order("id ASC").
			Limit(maxRel)
		if err := q.Find(&relRows).Error; err != nil {
			return nil, err
		}
	}

	outEnt := kgVizEntityRowsToDTO(entRows)
	outRel := kgVizRelRowsToDTO(relRows)
	truncated := scopedEntTotal > int64(len(entRows)) || scopedRelTotal > int64(len(relRows))

	return &rag.KnowledgeGraphViz{
		KnowledgeBaseID:           kbID,
		EnableKnowledgeGraph:      kb.EnableKnowledgeGraph,
		EntityCount:               entityCount,
		RelationshipCount:         relCount,
		Truncated:                 truncated,
		Entities:                  outEnt,
		Relationships:             outRel,
		MaxEntitiesRequested:      maxEnt,
		MaxRelationshipsRequested: maxRel,
		ScopeDocumentID:           documentID,
	}, nil
}

func kgVizEntityRowsToDTO(entRows []rag.RagKgEntity) []rag.KnowledgeGraphVizEntity {
	out := make([]rag.KnowledgeGraphVizEntity, 0, len(entRows))
	for _, e := range entRows {
		out = append(out, rag.KnowledgeGraphVizEntity{
			ID:          e.ID,
			Name:        e.Name,
			EntityType:  e.EntityType,
			Description: e.Description,
		})
	}
	return out
}

func kgVizRelRowsToDTO(relRows []rag.RagKgRelationship) []rag.KnowledgeGraphVizRelationship {
	out := make([]rag.KnowledgeGraphVizRelationship, 0, len(relRows))
	for _, r := range relRows {
		out = append(out, rag.KnowledgeGraphVizRelationship{
			ID:             r.ID,
			SourceEntityID: r.SourceEntityID,
			TargetEntityID: r.TargetEntityID,
			Keywords:       r.Keywords,
			Description:    r.Description,
		})
	}
	return out
}
