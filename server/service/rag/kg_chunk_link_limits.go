package rag

import (
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const kgMaxChunkLinksPerSideCap = 10000

func effectiveKgMaxChunksPerEntity() int {
	c := global.LRAG_CONFIG.Rag.KgMaxChunksPerEntity
	if c <= 0 {
		return 0
	}
	if c > kgMaxChunkLinksPerSideCap {
		return kgMaxChunkLinksPerSideCap
	}
	return c
}

func effectiveKgMaxChunksPerRelationship() int {
	c := global.LRAG_CONFIG.Rag.KgMaxChunksPerRelationship
	if c <= 0 {
		return 0
	}
	if c > kgMaxChunkLinksPerSideCap {
		return kgMaxChunkLinksPerSideCap
	}
	return c
}

// kgChunkLinkUseFIFO true 时与 LightRAG SOURCE_IDS_LIMIT_METHOD_FIFO 一致：保留按 rag_chunks.id 较新的关联
func kgChunkLinkUseFIFO() bool {
	m := strings.ToLower(strings.TrimSpace(global.LRAG_CONFIG.Rag.KgChunkLinkLimitMethod))
	if m == "keep" {
		return false
	}
	return true
}

func kgOrderedEntityChunkIDs(db *gorm.DB, entityID uint) ([]uint, error) {
	var rows []struct {
		ChunkID uint `gorm:"column:chunk_id"`
	}
	err := db.Table("rag_kg_entity_chunks AS ec").
		Select("ec.chunk_id").
		Joins("INNER JOIN rag_chunks c ON c.id = ec.chunk_id").
		Where("ec.entity_id = ?", entityID).
		Order("c.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ChunkID)
	}
	return out, nil
}

func kgOrderedRelationshipChunkIDs(db *gorm.DB, relID uint) ([]uint, error) {
	var rows []struct {
		ChunkID uint `gorm:"column:chunk_id"`
	}
	err := db.Table("rag_kg_relationship_chunks AS rc").
		Select("rc.chunk_id").
		Joins("INNER JOIN rag_chunks c ON c.id = rc.chunk_id").
		Where("rc.relationship_id = ?", relID).
		Order("c.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ChunkID)
	}
	return out, nil
}

// kgChunkIDsToDropForLimit 按 LightRAG apply_source_ids_limit：ids 为 rag_chunks.id 升序；fifo 删最旧 id，keep 删最新 id
func kgChunkIDsToDropForLimit(ids []uint, max int, fifo bool) []uint {
	if max <= 0 || len(ids) <= max {
		return nil
	}
	excess := len(ids) - max
	if fifo {
		return append([]uint(nil), ids[:excess]...)
	}
	return append([]uint(nil), ids[max:]...)
}

// kgTrimEntityChunkLinksIfNeeded 在新增关联后调用；按 LightRAG apply_source_ids_limit 语义裁剪
func kgTrimEntityChunkLinksIfNeeded(db *gorm.DB, entityID uint) {
	max := effectiveKgMaxChunksPerEntity()
	if max <= 0 || entityID == 0 {
		return
	}
	ids, err := kgOrderedEntityChunkIDs(db, entityID)
	if err != nil {
		return
	}
	del := kgChunkIDsToDropForLimit(ids, max, kgChunkLinkUseFIFO())
	if len(del) == 0 {
		return
	}
	if err := db.Where("entity_id = ? AND chunk_id IN ?", entityID, del).Delete(&rag.RagKgEntityChunk{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to trim entity chunk links", zap.Uint("entityId", entityID), zap.Error(err))
	}
}

func kgTrimRelationshipChunkLinksIfNeeded(db *gorm.DB, relID uint) {
	max := effectiveKgMaxChunksPerRelationship()
	if max <= 0 || relID == 0 {
		return
	}
	ids, err := kgOrderedRelationshipChunkIDs(db, relID)
	if err != nil {
		return
	}
	del := kgChunkIDsToDropForLimit(ids, max, kgChunkLinkUseFIFO())
	if len(del) == 0 {
		return
	}
	if err := db.Where("relationship_id = ? AND chunk_id IN ?", relID, del).Delete(&rag.RagKgRelationshipChunk{}).Error; err != nil {
		global.LRAG_LOG.Warn("failed to trim relationship chunk links", zap.Uint("relationshipId", relID), zap.Error(err))
	}
}
