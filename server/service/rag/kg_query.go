package rag

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// KnowledgeGraphHasEntities 知识库是否已有图谱实体（用于选择图谱检索或回退近似实现）
func KnowledgeGraphHasEntities(ctx context.Context, kbID uint) bool {
	ttlSec := global.LRAG_CONFIG.Rag.KgEntityPresenceCacheTTLSeconds
	if ttlSec > 0 {
		if v, ok := global.BlackCache.Get(kgEntityPresenceCacheKey(kbID)); ok {
			if s, _ := v.(string); s == "1" {
				return true
			}
		}
	}
	var n int64
	_ = global.LRAG_DB.WithContext(ctx).Model(&rag.RagKgEntity{}).
		Where("knowledge_base_id = ?", kbID).Count(&n).Error
	if n > 0 {
		if ttlSec > 0 {
			global.BlackCache.Set(kgEntityPresenceCacheKey(kbID), "1", time.Duration(ttlSec)*time.Second)
		}
		return true
	}
	return false
}

// KnowledgeGraphMapsForChunkIDs 根据切片 ID 列表返回 query/data 用的 entities / relationships 结构。
// 实体集 = 切片直连实体 ∪ 所有相关关系的端点；可选按 config.rag.kg-prompt-neighbor-rel-limit 再并入一跳邻接关系（与 LightRAG 式图上下文扩展方向一致，仍由 max_entity/max_relation token 在 FormatKnowledgeGraphPromptPrefix 侧裁剪）。
func KnowledgeGraphMapsForChunkIDs(ctx context.Context, kbIDs []uint, chunkIDs []uint) (entities []map[string]any, relationships []map[string]any) {
	if len(chunkIDs) == 0 || len(kbIDs) == 0 {
		return nil, nil
	}
	db := global.LRAG_DB.WithContext(ctx)

	var entLinkIDs []uint
	_ = db.Model(&rag.RagKgEntityChunk{}).Where("chunk_id IN ?", chunkIDs).Distinct().Pluck("entity_id", &entLinkIDs).Error

	var relLinkIDs []uint
	_ = db.Model(&rag.RagKgRelationshipChunk{}).Where("chunk_id IN ?", chunkIDs).Distinct().Pluck("relationship_id", &relLinkIDs).Error

	var directRels []rag.RagKgRelationship
	if len(relLinkIDs) > 0 {
		_ = db.Where("id IN ? AND knowledge_base_id IN ?", relLinkIDs, kbIDs).Find(&directRels).Error
	}

	seenRel := make(map[uint]struct{}, len(directRels)+16)
	for _, r := range directRels {
		seenRel[r.ID] = struct{}{}
	}

	seedEntities := make(map[uint]struct{}, len(entLinkIDs)+8)
	for _, id := range entLinkIDs {
		if id > 0 {
			seedEntities[id] = struct{}{}
		}
	}
	for _, r := range directRels {
		seedEntities[r.SourceEntityID] = struct{}{}
		seedEntities[r.TargetEntityID] = struct{}{}
	}

	allRels := append([]rag.RagKgRelationship(nil), directRels...)
	lim := EffectiveKgPromptNeighborRelLimit()
	if lim > 0 && len(seedEntities) > 0 {
		seedSlice := uintSetToSortedSlice(seedEntities)
		q := db.Where("knowledge_base_id IN ?", kbIDs).
			Where("(source_entity_id IN ? OR target_entity_id IN ?)", seedSlice, seedSlice)
		if len(seenRel) > 0 {
			excl := uintSetToSortedSlice(seenRel)
			q = q.Where("id NOT IN ?", excl)
		}
		var extra []rag.RagKgRelationship
		_ = q.Order("id ASC").Limit(lim).Find(&extra).Error
		for _, r := range extra {
			if _, ok := seenRel[r.ID]; ok {
				continue
			}
			seenRel[r.ID] = struct{}{}
			allRels = append(allRels, r)
		}
	}

	needEnt := make(map[uint]struct{}, len(entLinkIDs)+len(allRels)*2)
	for _, id := range entLinkIDs {
		if id > 0 {
			needEnt[id] = struct{}{}
		}
	}
	for _, r := range allRels {
		needEnt[r.SourceEntityID] = struct{}{}
		needEnt[r.TargetEntityID] = struct{}{}
	}
	entIDList := uintSetToSortedSlice(needEnt)
	var entRows []rag.RagKgEntity
	if len(entIDList) > 0 {
		_ = db.Where("id IN ? AND knowledge_base_id IN ?", entIDList, kbIDs).Find(&entRows).Error
	}
	nameByID := make(map[uint]string, len(entRows))
	seenEntOut := make(map[uint]struct{}, len(entRows))
	for _, e := range entRows {
		nameByID[e.ID] = e.Name
		if _, ok := seenEntOut[e.ID]; ok {
			continue
		}
		seenEntOut[e.ID] = struct{}{}
		entities = append(entities, map[string]any{
			"entity_name":        e.Name,
			"entity_type":        e.EntityType,
			"entity_description": e.Description,
		})
	}

	seenRelOut := make(map[uint]struct{}, len(allRels))
	for _, r := range allRels {
		if _, ok := seenRelOut[r.ID]; ok {
			continue
		}
		seenRelOut[r.ID] = struct{}{}
		relationships = append(relationships, map[string]any{
			"src_id":                   r.SourceEntityID,
			"tgt_id":                   r.TargetEntityID,
			"src_entity":               nameByID[r.SourceEntityID],
			"tgt_entity":               nameByID[r.TargetEntityID],
			"relationship_keywords":    r.Keywords,
			"relationship_description": r.Description,
		})
	}

	sort.Slice(entities, func(i, j int) bool {
		return strings.Compare(fmt.Sprint(entities[i]["entity_name"]), fmt.Sprint(entities[j]["entity_name"])) < 0
	})
	sort.Slice(relationships, func(i, j int) bool {
		a, b := relationships[i], relationships[j]
		sa, sb := fmt.Sprint(a["src_entity"]), fmt.Sprint(b["src_entity"])
		if sa != sb {
			return sa < sb
		}
		ta, tb := fmt.Sprint(a["tgt_entity"]), fmt.Sprint(b["tgt_entity"])
		if ta != tb {
			return ta < tb
		}
		return fmt.Sprint(a["relationship_keywords"]) < fmt.Sprint(b["relationship_keywords"])
	})
	return entities, relationships
}

func uintSetToSortedSlice(m map[uint]struct{}) []uint {
	out := make([]uint, 0, len(m))
	for id := range m {
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}

// ChunkIDsFromRAGDocs 从检索结果元数据中解析 rag_chunk_id；缺失时按 document_id 批量反查 chunk_index，减少 N+1 查询
func ChunkIDsFromRAGDocs(ctx context.Context, kbIDs []uint, docs []ragschema.Document) []uint {
	seen := make(map[uint]bool)
	var out []uint
	db := global.LRAG_DB.WithContext(ctx)

	pendingByDoc := make(map[uint]map[int]struct{})
	for _, d := range docs {
		if d.Metadata == nil {
			continue
		}
		if v, ok := d.Metadata["rag_chunk_id"]; ok {
			if id := anyToUint(v); id > 0 && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
			continue
		}
		docID := anyToUint(d.Metadata["document_id"])
		cidx := anyToInt(d.Metadata["chunk_index"])
		if docID == 0 {
			continue
		}
		if pendingByDoc[docID] == nil {
			pendingByDoc[docID] = make(map[int]struct{})
		}
		pendingByDoc[docID][cidx] = struct{}{}
	}

	var docKB map[uint]uint
	if len(kbIDs) > 0 && len(pendingByDoc) > 0 {
		docIDs := make([]uint, 0, len(pendingByDoc))
		for id := range pendingByDoc {
			docIDs = append(docIDs, id)
		}
		slices.Sort(docIDs)
		var drows []rag.RagDocument
		_ = db.Where("id IN ?", docIDs).Find(&drows).Error
		docKB = make(map[uint]uint, len(drows))
		for _, row := range drows {
			docKB[row.ID] = row.KnowledgeBaseID
		}
	}

	for docID, idxSet := range pendingByDoc {
		if len(idxSet) == 0 {
			continue
		}
		if len(kbIDs) > 0 {
			kb, ok := docKB[docID]
			if !ok || !containsUint(kbIDs, kb) {
				continue
			}
		}
		indices := make([]int, 0, len(idxSet))
		for ix := range idxSet {
			indices = append(indices, ix)
		}
		slices.Sort(indices)
		var chunks []rag.RagChunk
		if err := db.Where("document_id = ? AND chunk_index IN ?", docID, indices).Find(&chunks).Error; err != nil {
			continue
		}
		for _, ch := range chunks {
			if !seen[ch.ID] {
				seen[ch.ID] = true
				out = append(out, ch.ID)
			}
		}
	}
	return out
}

func anyToUint(v any) uint {
	switch x := v.(type) {
	case float64:
		return uint(x)
	case float32:
		return uint(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint(x)
	case int64:
		if x < 0 {
			return 0
		}
		return uint(x)
	case uint:
		return x
	case uint64:
		return uint(x)
	default:
		return 0
	}
}

func anyToInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}

func containsUint(slice []uint, x uint) bool {
	for _, s := range slice {
		if s == x {
			return true
		}
	}
	return false
}
