package rag

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/pageindex"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TextSplitter 段落和句子感知的文本分块器
// 优先按段落分割，段落过大时按句子分割，句子仍过大时按字符切分，支持重叠
func TextSplitter(text string, chunkSize, chunkOverlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 500
	}
	if chunkOverlap < 0 || chunkOverlap >= chunkSize {
		chunkOverlap = 50
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	runes := []rune(text)
	if len(runes) <= chunkSize {
		return []string{text}
	}

	paragraphs := splitParagraphs(text)
	var chunks []string
	var current strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		paraRunes := []rune(para)

		if current.Len() == 0 && len(paraRunes) <= chunkSize {
			current.WriteString(para)
			continue
		}

		if current.Len() > 0 && len([]rune(current.String()))+1+len(paraRunes) <= chunkSize {
			current.WriteString("\n")
			current.WriteString(para)
			continue
		}

		if current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			overlap := extractOverlap(current.String(), chunkOverlap)
			current.Reset()
			if overlap != "" {
				current.WriteString(overlap)
			}
		}

		if len(paraRunes) <= chunkSize {
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			current.WriteString(para)
		} else {
			sentences := splitSentences(para)
			for _, sent := range sentences {
				sent = strings.TrimSpace(sent)
				if sent == "" {
					continue
				}
				sentRunes := []rune(sent)

				if current.Len() > 0 && len([]rune(current.String()))+1+len(sentRunes) <= chunkSize {
					current.WriteString(" ")
					current.WriteString(sent)
				} else {
					if current.Len() > 0 {
						chunks = append(chunks, strings.TrimSpace(current.String()))
						overlap := extractOverlap(current.String(), chunkOverlap)
						current.Reset()
						if overlap != "" {
							current.WriteString(overlap)
						}
					}
					if len(sentRunes) <= chunkSize {
						if current.Len() > 0 {
							current.WriteString(" ")
						}
						current.WriteString(sent)
					} else {
						chunks = append(chunks, splitByChars(sent, chunkSize, chunkOverlap)...)
					}
				}
			}
		}
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	var result []string
	for _, c := range chunks {
		c = strings.TrimSpace(c)
		if c != "" {
			result = append(result, c)
		}
	}
	return result
}

// splitParagraphs 按空行分段
func splitParagraphs(text string) []string {
	parts := strings.Split(text, "\n\n")
	if len(parts) <= 1 {
		return strings.Split(text, "\n")
	}
	return parts
}

// splitSentences 按句子边界分割（中英文标点）
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])
		if isSentenceEnd(runes[i]) {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

func isSentenceEnd(r rune) bool {
	return r == '.' || r == '!' || r == '?' ||
		r == '。' || r == '！' || r == '？' ||
		r == '；' || r == '\n'
}

// extractOverlap 从文本末尾提取 overlap 长度的文本
func extractOverlap(text string, overlapSize int) string {
	runes := []rune(text)
	if len(runes) <= overlapSize {
		return text
	}
	return string(runes[len(runes)-overlapSize:])
}

// splitByChars 按字符数切分（用于超长句子的兜底）
func splitByChars(text string, chunkSize, chunkOverlap int) []string {
	runes := []rune(text)
	if len(runes) <= chunkSize {
		return []string{text}
	}
	var chunks []string
	start := 0
	for start < len(runes) {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[start:end])
		chunks = append(chunks, strings.TrimSpace(chunk))
		if end >= len(runes) {
			break
		}
		start = end - chunkOverlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

// EstimateTokenCount 粗略估算 token 数量（中文按字符，英文按空格分词）
func EstimateTokenCount(text string) int {
	count := 0
	runes := []rune(text)
	inWord := false
	for _, r := range runes {
		if r > 0x4E00 && r < 0x9FFF {
			count++
			inWord = false
		} else if r == ' ' || r == '\n' || r == '\t' {
			inWord = false
		} else {
			if !inWord {
				count++
				inWord = true
			}
		}
	}
	return count
}

// ParseTextContent 从 io.Reader 解析文本内容（支持 txt、md；带 BOM 的 UTF-16 LE/BE）。
func ParseTextContent(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	if len(b) >= 2 {
		if b[0] == 0xFF && b[1] == 0xFE {
			s := decodeUTF16Payload(b[2:], binary.LittleEndian)
			return strings.TrimSpace(s), nil
		}
		if b[0] == 0xFE && b[1] == 0xFF {
			s := decodeUTF16Payload(b[2:], binary.BigEndian)
			return strings.TrimSpace(s), nil
		}
	}
	if !utf8.Valid(b) {
		b = bytes.ToValidUTF8(b, []byte(" "))
	}
	return strings.TrimSpace(string(b)), nil
}

func decodeUTF16Payload(b []byte, order binary.ByteOrder) string {
	if len(b) < 2 {
		return ""
	}
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = order.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u))
}

func updateRagDocumentIfProcessing(ctx context.Context, docID uint, u map[string]any) error {
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("id = ? AND status = ?", docID, "processing").Updates(u).Error
}

func effectiveChunkRankBoostFloor() float64 {
	f := global.LRAG_CONFIG.Rag.ChunkRankBoostPositionFloor
	if f <= 0 {
		return 0.35
	}
	if f > 1 {
		return 1
	}
	return f
}

// enrichMetadataRankBoostForChunk 若尚无 rank_boost：优先 chunk-rank-boost-by-position（按序号衰减）；否则 default-chunk-rank-boost。供 ApplyConfiguredRankBoostToScores 使用。
func enrichMetadataRankBoostForChunk(meta map[string]any, chunkIndex, chunkTotal int) {
	if meta == nil {
		return
	}
	if _, ok := meta["rank_boost"]; ok {
		return
	}
	cfg := global.LRAG_CONFIG.Rag
	if cfg.ChunkRankBoostByPosition && chunkTotal > 0 {
		floor := effectiveChunkRankBoostFloor()
		if chunkTotal == 1 {
			meta["rank_boost"] = 1.0
			return
		}
		idx := chunkIndex
		if idx < 0 {
			idx = 0
		}
		if idx >= chunkTotal {
			idx = chunkTotal - 1
		}
		t := float64(idx) / float64(chunkTotal-1)
		rb := 1.0 - t*(1.0-floor)
		meta["rank_boost"] = rb
		return
	}
	v := cfg.DefaultChunkRankBoost
	if v <= 0 {
		return
	}
	if v > 1 {
		v = 1
	}
	meta["rank_boost"] = v
}

func rankBoostFromMetadata(meta map[string]any) float64 {
	if meta == nil {
		return 0
	}
	v, ok := meta["rank_boost"]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint64:
		return float64(x)
	default:
		return 0
	}
}

// applyDocumentPriorityFloorToRankBoost 文档级 priority 作为 rank_boost 下限：rank_boost = max(当前, priority)。
func applyDocumentPriorityFloorToRankBoost(meta map[string]any, priority float64) {
	if meta == nil {
		return
	}
	p := priority
	if p <= 0 {
		return
	}
	if p > 1 {
		p = 1
	}
	if cur := rankBoostFromMetadata(meta); p > cur {
		meta["rank_boost"] = p
	}
}

// ProcessDocument 解析、切片、向量化文档并写入向量库；当 RetrieverType=pageindex 时构建 PageIndex 树
func ProcessDocument(ctx context.Context, doc *rag.RagDocument, content string, kb *rag.RagKnowledgeBase, userID ...uint) error {
	var uid uint
	if len(userID) > 0 {
		uid = userID[0]
	} else {
		uid = kb.OwnerID
	}
	var fresh rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).First(&fresh, doc.ID).Error; err != nil {
		return err
	}
	if fresh.Status != "processing" {
		return nil
	}
	*doc = fresh
	renewDocumentIndexingLease(ctx, doc.ID)

	if content == "" {
		if err := updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":      "completed",
			"chunk_count": 0,
		}); err != nil {
			return err
		}
		BumpRetrieveCacheEpochForKnowledgeBase(kb.ID)
		return nil
	}
	// PageIndex 推理检索：构建树结构并存储
	if strings.ToLower(kb.RetrieverType) == "pageindex" {
		return processDocumentPageIndex(ctx, doc, content, kb, uid)
	}

	// 根据知识库配置的切片方法进行切片
	cfg := ChunkConfigFromKB(kb)
	chunks := ChunkDocument(content, cfg)
	if len(chunks) == 0 {
		if err := updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":      "completed",
			"chunk_count": 0,
		}); err != nil {
			return err
		}
		BumpRetrieveCacheEpochForKnowledgeBase(kb.ID)
		return nil
	}

	renewDocumentIndexingLease(ctx, doc.ID)

	emb, err := resolveEmbeddingConfig(ctx, kb, uid)
	if err != nil {
		msg := "嵌入模型配置不存在，请先在模型配置管理中添加嵌入模型"
		_ = updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":    "failed",
			"error_msg": msg,
		})
		return err
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		_ = updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":    "failed",
			"error_msg": "创建嵌入模型失败",
		})
		return err
	}
	ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
	store, err := createVectorStoreFromKB(ctx, kb, embedder, ns, emb.Dimensions)
	if err != nil {
		_ = updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":    "failed",
			"error_msg": "创建向量存储失败",
		})
		return err
	}

	CleanupKnowledgeGraphLinksForDocument(ctx, kb, uid, doc.ID)
	_ = store.DeleteByMetadata(ctx, ns, "document_id", doc.ID)
	global.LRAG_DB.WithContext(ctx).Where("document_id = ?", doc.ID).Delete(&rag.RagChunk{})

	ragChunks := make([]rag.RagChunk, len(chunks))
	for i, c := range chunks {
		ragChunks[i] = rag.RagChunk{
			UUID:       uuid.New(),
			DocumentID: doc.ID,
			Content:    c,
			ChunkIndex: i,
		}
	}
	if len(ragChunks) > 0 {
		if err := global.LRAG_DB.WithContext(ctx).CreateInBatches(ragChunks, 100).Error; err != nil {
			global.LRAG_LOG.Error("切片持久化失败", zap.Uint("docID", doc.ID), zap.Error(err))
			_ = updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
				"status":    "failed",
				"error_msg": "切片持久化失败: " + err.Error(),
			})
			return err
		}
	}

	docs := make([]schema.Document, len(chunks))
	for i, c := range chunks {
		meta := map[string]any{
			"document_id":  doc.ID,
			"chunk_index":  i,
			"doc_name":     doc.Name,
			"rag_chunk_id": ragChunks[i].ID,
		}
		enrichMetadataRankBoostForChunk(meta, i, len(chunks))
		applyDocumentPriorityFloorToRankBoost(meta, doc.Priority)
		docs[i] = schema.Document{PageContent: c, Metadata: meta}
	}
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = ns },
	}
	renewDocumentIndexingLease(ctx, doc.ID)
	vectorIDs, err := store.AddDocuments(ctx, docs, opts...)
	if err != nil {
		_ = updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":    "failed",
			"error_msg": "向量化写入失败: " + err.Error(),
		})
		return err
	}
	for i := range ragChunks {
		vid := ""
		if i < len(vectorIDs) {
			vid = vectorIDs[i]
		}
		if err := global.LRAG_DB.WithContext(ctx).Model(&ragChunks[i]).Update("vector_store_id", vid).Error; err != nil {
			global.LRAG_LOG.Warn("回写切片向量 ID 失败", zap.Uint("chunkID", ragChunks[i].ID), zap.Error(err))
		}
	}

	if err := updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
		"status":      "completed",
		"chunk_count": len(chunks),
		"error_msg":   "",
	}); err != nil {
		return err
	}
	if kb.EnableKnowledgeGraph {
		ScheduleKnowledgeGraphExtraction(uid, kb.ID, doc.ID)
	}
	BumpRetrieveCacheEpochForKnowledgeBase(kb.ID)
	global.LRAG_LOG.Info("文档处理完成",
		zap.Uint("docID", doc.ID),
		zap.String("name", doc.Name),
		zap.Int("chunks", len(chunks)))
	return nil
}

// processDocumentPageIndex 为 PageIndex 检索构建树结构
func processDocumentPageIndex(ctx context.Context, doc *rag.RagDocument, content string, kb *rag.RagKnowledgeBase, userID uint) error {
	var fresh rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).First(&fresh, doc.ID).Error; err != nil {
		return err
	}
	if fresh.Status != "processing" {
		return nil
	}
	*doc = fresh
	renewDocumentIndexingLease(ctx, doc.ID)

	tree := pageindex.BuildTreeForPageIndex(doc.FileType, doc.Name, content)
	treeJSON, err := json.Marshal(tree)
	if err != nil {
		_ = updateRagDocumentIfProcessing(ctx, doc.ID, map[string]any{
			"status":    "failed",
			"error_msg": "序列化 PageIndex 树失败: " + err.Error(),
		})
		return err
	}

	// 将 PageIndex 节点也持久化到 rag_chunks，便于查看和编辑
	nodeList := pageindex.StructureToList(tree)
	CleanupKnowledgeGraphLinksForDocument(ctx, kb, userID, doc.ID)
	global.LRAG_DB.WithContext(ctx).Where("document_id = ?", doc.ID).Delete(&rag.RagChunk{})
	if len(nodeList) > 0 {
		ragChunks := make([]rag.RagChunk, len(nodeList))
		for i, node := range nodeList {
			ragChunks[i] = rag.RagChunk{
				UUID:       uuid.New(),
				DocumentID: doc.ID,
				Content:    node.Text,
				ChunkIndex: i,
			}
		}
		if err := global.LRAG_DB.WithContext(ctx).CreateInBatches(ragChunks, 100).Error; err != nil {
			global.LRAG_LOG.Error("PageIndex 切片持久化失败", zap.Uint("docID", doc.ID), zap.Error(err))
		} else {
			parentByChild := pageindex.BuildParentNormIDByChild(tree)
			normToID := make(map[string]uint)
			for i := 0; i < len(nodeList) && i < len(ragChunks); i++ {
				if nodeList[i] == nil {
					continue
				}
				normToID[pageindex.NormalizeNodeID(nodeList[i].NodeID)] = ragChunks[i].ID
			}
			for i := 0; i < len(ragChunks) && i < len(nodeList); i++ {
				if nodeList[i] == nil {
					continue
				}
				nid := pageindex.NormalizeNodeID(nodeList[i].NodeID)
				meta := map[string]any{"node_id": nid}
				if pnorm, ok := parentByChild[nid]; ok && pnorm != "" {
					if pid, ok2 := normToID[pnorm]; ok2 && pid != 0 {
						meta["parent_rag_chunk_id"] = pid
					}
				}
				b, mErr := json.Marshal(meta)
				if mErr != nil {
					continue
				}
				_ = global.LRAG_DB.WithContext(ctx).Model(&ragChunks[i]).Update("metadata", string(b)).Error
			}
			indexPageIndexChunksToVectorStore(ctx, doc, kb, userID, nodeList)
		}
	}

	updates := map[string]any{
		"status":               "completed",
		"chunk_count":          len(nodeList),
		"error_msg":            "",
		"page_index_structure": string(treeJSON),
	}
	if err := updateRagDocumentIfProcessing(ctx, doc.ID, updates); err != nil {
		return err
	}
	if kb.EnableKnowledgeGraph {
		ScheduleKnowledgeGraphExtraction(userID, kb.ID, doc.ID)
	}
	BumpRetrieveCacheEpochForKnowledgeBase(kb.ID)
	global.LRAG_LOG.Info("PageIndex 树构建完成",
		zap.Uint("docID", doc.ID),
		zap.String("name", doc.Name),
		zap.Int("nodes", len(nodeList)))
	return nil
}

// indexPageIndexChunksToVectorStore 将 PageIndex 节点切片写入向量库，对齐 references/ragflow 中「向量召回 + TOC 增强」的前半段。
func indexPageIndexChunksToVectorStore(ctx context.Context, doc *rag.RagDocument, kb *rag.RagKnowledgeBase, userID uint, nodeList []*pageindex.TreeNode) {
	if kb == nil || doc == nil || len(nodeList) == 0 {
		return
	}
	var persisted []rag.RagChunk
	if err := global.LRAG_DB.WithContext(ctx).Where("document_id = ?", doc.ID).Order("chunk_index ASC").Find(&persisted).Error; err != nil || len(persisted) == 0 {
		return
	}
	emb, err := resolveEmbeddingConfig(ctx, kb, userID)
	if err != nil {
		global.LRAG_LOG.Warn("PageIndex 向量索引跳过：嵌入配置不可用", zap.Uint("docID", doc.ID), zap.Error(err))
		return
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		global.LRAG_LOG.Warn("PageIndex 向量索引跳过：创建嵌入模型失败", zap.Uint("docID", doc.ID), zap.Error(err))
		return
	}
	ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
	store, err := createVectorStoreFromKB(ctx, kb, embedder, ns, emb.Dimensions)
	if err != nil || store == nil {
		global.LRAG_LOG.Warn("PageIndex 向量索引跳过：向量存储不可用", zap.Uint("docID", doc.ID), zap.Error(err))
		return
	}
	renewDocumentIndexingLease(ctx, doc.ID)
	_ = store.DeleteByMetadata(ctx, ns, "document_id", doc.ID)
	lim := len(persisted)
	if len(nodeList) < lim {
		lim = len(nodeList)
	}
	vecDocs := make([]schema.Document, lim)
	for i := 0; i < lim; i++ {
		nid := ""
		if nodeList[i] != nil {
			nid = pageindex.NormalizeNodeID(nodeList[i].NodeID)
		}
		meta := map[string]any{
			"document_id":  doc.ID,
			"chunk_index":  persisted[i].ChunkIndex,
			"doc_name":     doc.Name,
			"rag_chunk_id": persisted[i].ID,
			"node_id":      nid,
		}
		if strings.TrimSpace(persisted[i].Metadata) != "" {
			var extra map[string]any
			if err := json.Unmarshal([]byte(persisted[i].Metadata), &extra); err == nil {
				for k, v := range extra {
					if _, exists := meta[k]; !exists {
						meta[k] = v
					}
				}
			}
		}
		enrichMetadataRankBoostForChunk(meta, i, lim)
		applyDocumentPriorityFloorToRankBoost(meta, doc.Priority)
		vecDocs[i] = schema.Document{
			PageContent: strings.TrimSpace(persisted[i].Content),
			Metadata:    meta,
		}
	}
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = ns },
	}
	vectorIDs, err := store.AddDocuments(ctx, vecDocs, opts...)
	if err != nil {
		global.LRAG_LOG.Warn("PageIndex 向量写入失败", zap.Uint("docID", doc.ID), zap.Error(err))
		return
	}
	for i := 0; i < lim; i++ {
		vid := ""
		if i < len(vectorIDs) {
			vid = vectorIDs[i]
		}
		_ = global.LRAG_DB.WithContext(ctx).Model(&persisted[i]).Update("vector_store_id", vid).Error
	}
}
