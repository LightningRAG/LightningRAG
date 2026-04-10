package rag

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	llmprov "github.com/LightningRAG/LightningRAG/server/rag/providers/llm"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const kgExtractMaxRunesPerBatch = 5000
const kgExtractMaxChunksPerBatch = 8

// kgExtractSystemPrompt 与 callKgExtractLLMUncached / 补抽轮共用；实体类型集对齐 references/LightRAG/lightrag/constants.py DEFAULT_ENTITY_TYPES（并含 prompt 示例中的细类）
const kgExtractSystemPrompt = `你是知识图谱抽取助手。根据输入中多个 --- chunk_index=N --- 段落，为每个 chunk 抽取实体与二元关系。

实体 type 字段请从下列类型中选最合适的一项；若无匹配则用 Other：Person, Creature, Organization, Location, Event, Concept, Method, Content, Data, Artifact, NaturalObject, category, equipment, product

写作约束（对齐 LightRAG entity_extraction 提示）：第三人称客观描述；避免「本文」「我们」「该文章」等模糊指代；专有名词保持原文语言。

输出必须是单一 JSON 对象，不要 Markdown，不要解释。JSON 结构：
{"per_chunk":[{"chunk_index":整数,"entities":[{"name":"实体名","type":"类型","description":"简短描述"}],"relationships":[{"source":"源实体名","target":"目标实体名","keywords":"关键词,逗号分隔","description":"关系说明"}]}]}
实体名在同一文档内保持一致。关系视为无向语义：同一对实体不要因方向互换重复输出。仅使用输入中明确信息；若无关系则 relationships 可为 []。`

const kgExtractGleanUserPrompt = `请根据对话里上一条助手消息中的 JSON 抽取结果，重新对照用户消息里的原文。
仅输出**新发现的**实体与关系，或**需要修正**的实体与关系（与第一轮完全相同的单一 JSON 对象结构，不要 Markdown）。
不要重复输出上一轮已完整、正确的项。若无需补充，各 chunk 的 entities、relationships 使用空数组 []。`

type kgExtractBatch struct {
	Items []kgExtractBatchItem
}

type kgExtractBatchItem struct {
	ChunkIndex int    `json:"chunk_index"`
	Text       string `json:"text"`
}

type kgLLMExtractResult struct {
	PerChunk []kgLLMChunkExtract `json:"per_chunk"`
}

type kgLLMChunkExtract struct {
	ChunkIndex    int                 `json:"chunk_index"`
	Entities      []kgLLMEntity       `json:"entities"`
	Relationships []kgLLMRelationship `json:"relationships"`
}

type kgLLMEntity struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type kgLLMRelationship struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Keywords    string `json:"keywords"`
	Description string `json:"description"`
}

var reJSONFence = regexp.MustCompile("(?s)```(?:json)?\\s*([\\s\\S]*?)```")

// ExtractKnowledgeGraphForDocument 对已完成索引的文档做 LLM 抽取并写入图谱（异步任务中调用）
func ExtractKnowledgeGraphForDocument(ctx context.Context, uid uint, kb *rag.RagKnowledgeBase, documentID uint) error {
	if kb == nil || !kb.EnableKnowledgeGraph {
		return nil
	}
	db := global.LRAG_DB.WithContext(ctx)
	var chunks []rag.RagChunk
	if err := db.Where("document_id = ?", documentID).Order("chunk_index ASC").Find(&chunks).Error; err != nil {
		return err
	}
	if len(chunks) == 0 {
		return nil
	}
	llm := resolveExtractLLM(ctx, uid)
	if llm == nil {
		return fmt.Errorf("知识图谱抽取需要可用的对话模型，请配置管理员 LLM")
	}
	store, _, baseNS, err := kgOpenVectorStore(ctx, kb, uid)
	if err != nil {
		return err
	}

	docPriority := 0.0
	var docRow rag.RagDocument
	if err := db.Select("id", "priority").First(&docRow, documentID).Error; err == nil {
		docPriority = docRow.Priority
	}

	batches := buildKgExtractBatches(chunks)
	for _, b := range batches {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw, err := callKgExtractLLM(ctx, kb.ID, llm, b)
		if err != nil {
			global.LRAG_LOG.Warn("知识图谱批次抽取失败", zap.Uint("docID", documentID), zap.Error(err))
			continue
		}
		parsed, perr := parseKgExtractJSON(raw)
		if perr != nil {
			global.LRAG_LOG.Warn("知识图谱 JSON 解析失败", zap.Uint("docID", documentID), zap.Error(perr), zap.String("snippet", truncateRunes(raw, 200)))
			continue
		}
		if err := persistKgExtractResult(ctx, kb.ID, chunks, parsed, store, baseNS, llm, docPriority); err != nil {
			global.LRAG_LOG.Warn("知识图谱持久化失败", zap.Uint("docID", documentID), zap.Error(err))
			continue
		}
		if effectiveKgExtractMaxGleaning() > 0 && strings.TrimSpace(raw) != "" {
			userPl := kgExtractBatchUserPayload(b)
			asst := strings.TrimSpace(llmprov.StripAssistantReasoningMarkers(raw))
			if !kgGleaningFitsInputBudget(userPl, asst) {
				global.LRAG_LOG.Warn("知识图谱补抽跳过：预估输入超过 kg-extract-gleaning-max-input-tokens",
					zap.Uint("docID", documentID),
					zap.Int("maxApproxTokens", effectiveKgGleaningMaxInputTokens()))
				continue
			}
			gleanRaw, gerr := callKgExtractGleaningLLM(ctx, llm, b, raw)
			if gerr != nil {
				global.LRAG_LOG.Warn("知识图谱补抽失败", zap.Uint("docID", documentID), zap.Error(gerr))
				continue
			}
			gleanParsed, gperr := parseKgExtractJSON(gleanRaw)
			if gperr != nil {
				global.LRAG_LOG.Warn("知识图谱补抽 JSON 解析失败", zap.Uint("docID", documentID), zap.Error(gperr), zap.String("snippet", truncateRunes(gleanRaw, 200)))
				continue
			}
			if err := persistKgExtractResult(ctx, kb.ID, chunks, gleanParsed, store, baseNS, llm, docPriority); err != nil {
				global.LRAG_LOG.Warn("知识图谱补抽持久化失败", zap.Uint("docID", documentID), zap.Error(err))
			}
		}
	}
	return nil
}

// effectiveKgExtractMaxGleaning 是否启用补抽：>0 表示在首轮成功后追加一轮（与 LightRAG 单轮 gleaning 对齐，不链式多轮以控制成本）
func effectiveKgExtractMaxGleaning() int {
	g := global.LRAG_CONFIG.Rag.KgExtractMaxGleaning
	if g <= 0 {
		return 0
	}
	return 1
}

// effectiveKgGleaningMaxInputTokens 补抽输入粗算上限（与 LightRAG max_extract_input_tokens 同源思路）
func effectiveKgGleaningMaxInputTokens() int {
	n := global.LRAG_CONFIG.Rag.KgExtractGleaningMaxInputTokens
	if n <= 0 {
		return 20480
	}
	return n
}

// kgGleaningFitsInputBudget 用 estimateTokens（约 4 字符/token）粗算多轮消息总长度，防止补抽撑爆上下文
func kgGleaningFitsInputBudget(userPayload, strippedAssistant string) bool {
	maxTok := effectiveKgGleaningMaxInputTokens()
	approx := estimateTokens(kgExtractSystemPrompt) + estimateTokens(userPayload) +
		estimateTokens(strippedAssistant) + estimateTokens(kgExtractGleanUserPrompt) + 256
	return approx <= maxTok
}

func resolveExtractLLM(ctx context.Context, uid uint) interfaces.LLM {
	provider, modelName, baseURL, apiKey, ok := ResolveModelWithFallback(ctx, uid, 0, 0, "", interfaces.ModelTypeChat)
	if !ok {
		return nil
	}
	inst, err := registry.CreateLLM(registry.LLMConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
	if err != nil || inst == nil {
		return nil
	}
	return inst
}

func kgOpenVectorStore(ctx context.Context, kb *rag.RagKnowledgeBase, uid uint) (interfaces.VectorStore, interfaces.Embedder, string, error) {
	emb, err := resolveEmbeddingConfig(ctx, kb, uid)
	if err != nil {
		return nil, nil, "", err
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		return nil, nil, "", fmt.Errorf("创建嵌入模型失败")
	}
	ns := "kb_" + fmtUint(kb.ID)
	store, err := createVectorStoreFromKB(ctx, kb, embedder, ns, emb.Dimensions)
	if err != nil {
		return nil, nil, "", err
	}
	return store, embedder, ns, nil
}

func fmtUint(id uint) string {
	return fmt.Sprintf("%d", id)
}

func buildKgExtractBatches(chunks []rag.RagChunk) []kgExtractBatch {
	var out []kgExtractBatch
	var cur kgExtractBatch
	runes := 0
	for _, ch := range chunks {
		t := strings.TrimSpace(ch.Content)
		if t == "" {
			continue
		}
		n := utf8.RuneCountInString(t)
		if len(cur.Items) >= kgExtractMaxChunksPerBatch || (len(cur.Items) > 0 && runes+n > kgExtractMaxRunesPerBatch) {
			out = append(out, cur)
			cur = kgExtractBatch{}
			runes = 0
		}
		cur.Items = append(cur.Items, kgExtractBatchItem{ChunkIndex: ch.ChunkIndex, Text: t})
		runes += n
	}
	if len(cur.Items) > 0 {
		out = append(out, cur)
	}
	return out
}

func callKgExtractLLM(ctx context.Context, kbID uint, llm interfaces.LLM, batch kgExtractBatch) (string, error) {
	user := kgExtractBatchUserPayload(batch)
	key := kgExtractLLMCacheKey(kbID, user)
	ttlSec := global.LRAG_CONFIG.Rag.KgExtractLLMCacheTTLSeconds
	if ttlSec > 0 {
		if raw, ok := loadKgExtractLLMCache(key); ok {
			return raw, nil
		}
	}
	v, err, _ := global.LRAG_Concurrency_Control.Do("lrag_kgx:"+key, func() (interface{}, error) {
		if ttlSec > 0 {
			if raw, ok := loadKgExtractLLMCache(key); ok {
				return raw, nil
			}
		}
		raw, e := callKgExtractLLMUncached(ctx, llm, user)
		if e != nil {
			return "", e
		}
		if ttlSec > 0 {
			saveKgExtractLLMCache(key, raw, ttlSec)
		}
		return raw, nil
	})
	if err != nil {
		return "", err
	}
	s, _ := v.(string)
	return s, nil
}

func callKgExtractLLMUncached(ctx context.Context, llm interfaces.LLM, user string) (string, error) {
	msgs := []interfaces.MessageContent{
		{Role: interfaces.MessageRoleSystem, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: kgExtractSystemPrompt}}},
		{Role: interfaces.MessageRoleHuman, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: user}}},
	}
	resp, err := llm.GenerateContent(ctx, msgs, interfaces.WithTemperature(0.1), interfaces.WithMaxTokens(4096))
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty llm response")
	}
	return resp.Choices[0].Content, nil
}

// callKgExtractGleaningLLM 第二轮补抽：user → assistant(首轮 JSON) → user(补抽指令)；不经过首轮 LLM 缓存
func callKgExtractGleaningLLM(ctx context.Context, llm interfaces.LLM, batch kgExtractBatch, firstAssistantRaw string) (string, error) {
	user := kgExtractBatchUserPayload(batch)
	assistant := strings.TrimSpace(llmprov.StripAssistantReasoningMarkers(firstAssistantRaw))
	if assistant == "" {
		return "", fmt.Errorf("empty first extraction for gleaning")
	}
	return callKgExtractGleaningLLMUncached(ctx, llm, user, assistant)
}

func callKgExtractGleaningLLMUncached(ctx context.Context, llm interfaces.LLM, userPayload, firstAssistant string) (string, error) {
	msgs := []interfaces.MessageContent{
		{Role: interfaces.MessageRoleSystem, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: kgExtractSystemPrompt}}},
		{Role: interfaces.MessageRoleHuman, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: userPayload}}},
		{Role: interfaces.MessageRoleAssistant, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: firstAssistant}}},
		{Role: interfaces.MessageRoleHuman, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: kgExtractGleanUserPrompt}}},
	}
	resp, err := llm.GenerateContent(ctx, msgs, interfaces.WithTemperature(0.1), interfaces.WithMaxTokens(4096))
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty llm response")
	}
	return resp.Choices[0].Content, nil
}

func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func persistKgExtractResult(ctx context.Context, kbID uint, allChunks []rag.RagChunk, res *kgLLMExtractResult, store interfaces.VectorStore, baseNS string, llm interfaces.LLM, docPriority float64) error {
	chunkByIndex := make(map[int]uint)
	for _, ch := range allChunks {
		chunkByIndex[ch.ChunkIndex] = ch.ID
	}
	db := global.LRAG_DB.WithContext(ctx)
	for _, pc := range res.PerChunk {
		cid, ok := chunkByIndex[pc.ChunkIndex]
		if !ok || cid == 0 {
			continue
		}
		nameToID := make(map[string]uint)
		for _, e := range pc.Entities {
			ent, err := kgUpsertEntity(ctx, db, kbID, llm, e)
			if err != nil {
				continue
			}
			nameToID[normalizeKgName(e.Name)] = ent.ID
			kgLinkEntityChunk(db, ent.ID, cid)
			_ = kgSyncEntityEmbedding(ctx, store, baseNS, ent, docPriority)
		}
		for _, r := range pc.Relationships {
			sid, ok1 := nameToID[normalizeKgName(r.Source)]
			tid, ok2 := nameToID[normalizeKgName(r.Target)]
			if !ok1 || !ok2 || sid == tid {
				sid2, e1 := kgResolveEntityIDByName(db, kbID, r.Source)
				tid2, e2 := kgResolveEntityIDByName(db, kbID, r.Target)
				if e1 != nil || e2 != nil || sid2 == 0 || tid2 == 0 || sid2 == tid2 {
					continue
				}
				sid, tid = sid2, tid2
			}
			rel, err := kgUpsertRelationship(ctx, db, kbID, llm, sid, tid, r)
			if err != nil {
				continue
			}
			kgLinkRelChunk(db, rel.ID, cid)
			srcN, tgtN := kgRelationshipEndpointNames(db, rel)
			if srcN == "" {
				srcN = strings.TrimSpace(r.Source)
			}
			if tgtN == "" {
				tgtN = strings.TrimSpace(r.Target)
			}
			_ = kgSyncRelEmbedding(ctx, store, baseNS, rel, srcN, tgtN, docPriority)
		}
	}
	return nil
}

func normalizeKgName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func kgUpsertEntity(ctx context.Context, db *gorm.DB, kbID uint, llm interfaces.LLM, e kgLLMEntity) (*rag.RagKgEntity, error) {
	name := kgTruncateRunes(strings.TrimSpace(e.Name), effectiveKgMaxEntityNameRunes())
	if name == "" {
		return nil, fmt.Errorf("empty entity name")
	}
	norm := normalizeKgName(name)
	descIn := kgClampStoredDescription(e.Description)
	typeIn := kgClampEntityType(e.Type)
	var ent rag.RagKgEntity
	err := db.Where("knowledge_base_id = ? AND normalized_name = ?", kbID, norm).First(&ent).Error
	if err == nil {
		merged := kgClampStoredDescription(mergeKgDescription(ent.Description, descIn))
		display := strings.TrimSpace(ent.Name)
		if display == "" {
			display = name
		}
		merged = kgMaybeSummarizeEntityDescription(ctx, llm, display, merged)
		updates := map[string]any{}
		if merged != ent.Description {
			updates["description"] = merged
			ent.Description = merged
		}
		oldT := strings.TrimSpace(ent.EntityType)
		if typeIn != "" && !strings.EqualFold(typeIn, "other") {
			if oldT == "" || strings.EqualFold(oldT, "other") {
				updates["entity_type"] = typeIn
				ent.EntityType = typeIn
			}
		}
		if len(updates) > 0 {
			_ = db.Model(&ent).Updates(updates).Error
		}
		return &ent, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	ent = rag.RagKgEntity{
		KnowledgeBaseID: kbID,
		Name:            name,
		NormalizedName:  norm,
		EntityType:      typeIn,
		Description:     descIn,
	}
	if err := db.Create(&ent).Error; err != nil {
		return nil, err
	}
	return &ent, nil
}

func kgResolveEntityIDByName(db *gorm.DB, kbID uint, name string) (uint, error) {
	norm := normalizeKgName(name)
	if norm == "" {
		return 0, fmt.Errorf("empty")
	}
	var ent rag.RagKgEntity
	err := db.Where("knowledge_base_id = ? AND normalized_name = ?", kbID, norm).First(&ent).Error
	if err != nil {
		return 0, err
	}
	return ent.ID, nil
}

// kgUpsertRelationship 写入或复用关系；对齐 LightRAG 对无向语义边的处理：同时匹配 (src,tgt) 与 (tgt,src)，避免反向重复边。
func kgUpsertRelationship(ctx context.Context, db *gorm.DB, kbID uint, llm interfaces.LLM, src, tgt uint, r kgLLMRelationship) (*rag.RagKgRelationship, error) {
	if src == 0 || tgt == 0 || src == tgt {
		return nil, fmt.Errorf("invalid relationship endpoints")
	}
	var rel rag.RagKgRelationship
	err := db.Where("knowledge_base_id = ? AND source_entity_id = ? AND target_entity_id = ?", kbID, src, tgt).First(&rel).Error
	if err == nil {
		srcN, tgtN := kgRelationshipEndpointNames(db, &rel)
		if srcN == "" {
			srcN = strings.TrimSpace(r.Source)
		}
		if tgtN == "" {
			tgtN = strings.TrimSpace(r.Target)
		}
		kgMergeRelationshipContent(ctx, db, llm, &rel, r, srcN, tgtN)
		return &rel, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	err = db.Where("knowledge_base_id = ? AND source_entity_id = ? AND target_entity_id = ?", kbID, tgt, src).First(&rel).Error
	if err == nil {
		srcN, tgtN := kgRelationshipEndpointNames(db, &rel)
		if srcN == "" {
			srcN = strings.TrimSpace(r.Source)
		}
		if tgtN == "" {
			tgtN = strings.TrimSpace(r.Target)
		}
		kgMergeRelationshipContent(ctx, db, llm, &rel, r, srcN, tgtN)
		return &rel, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	rel = rag.RagKgRelationship{
		KnowledgeBaseID: kbID,
		SourceEntityID:  src,
		TargetEntityID:  tgt,
		Keywords:        kgClampStoredKeywords(r.Keywords),
		Description:     kgClampStoredDescription(r.Description),
	}
	if err := db.Create(&rel).Error; err != nil {
		return nil, err
	}
	return &rel, nil
}

func kgRelationshipEndpointNames(db *gorm.DB, rel *rag.RagKgRelationship) (srcName, tgtName string) {
	if rel == nil {
		return "", ""
	}
	var a, b rag.RagKgEntity
	if err := db.Select("name").Where("id = ?", rel.SourceEntityID).First(&a).Error; err != nil {
		return "", ""
	}
	if err := db.Select("name").Where("id = ?", rel.TargetEntityID).First(&b).Error; err != nil {
		return "", ""
	}
	return strings.TrimSpace(a.Name), strings.TrimSpace(b.Name)
}

func kgMergeRelationshipContent(ctx context.Context, db *gorm.DB, llm interfaces.LLM, rel *rag.RagKgRelationship, r kgLLMRelationship, srcDisplay, tgtDisplay string) {
	if rel == nil || rel.ID == 0 {
		return
	}
	mergedK := kgClampStoredKeywords(mergeKgCommaSeparatedUnique(rel.Keywords, r.Keywords))
	mergedD := kgClampStoredDescription(mergeKgDescription(rel.Description, r.Description))
	mergedD = kgMaybeSummarizeRelationshipDescription(ctx, llm, srcDisplay, tgtDisplay, mergedD)
	if mergedK == rel.Keywords && mergedD == rel.Description {
		return
	}
	rel.Keywords = mergedK
	rel.Description = mergedD
	_ = db.Model(rel).Updates(map[string]any{
		"keywords":    mergedK,
		"description": mergedD,
	}).Error
}

func mergeKgCommaSeparatedUnique(existing, add string) string {
	seen := make(map[string]struct{})
	var parts []string
	for _, s := range strings.Split(existing, ",") {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		parts = append(parts, t)
	}
	for _, s := range strings.Split(add, ",") {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		parts = append(parts, t)
	}
	return strings.Join(parts, ", ")
}

func mergeKgDescription(old, new string) string {
	new = strings.TrimSpace(new)
	if new == "" {
		return strings.TrimSpace(old)
	}
	old = strings.TrimSpace(old)
	if old == "" {
		return new
	}
	if strings.Contains(old, new) {
		return old
	}
	if strings.Contains(new, old) {
		return new
	}
	return old + "\n" + new
}

func kgLinkEntityChunk(db *gorm.DB, entityID, chunkID uint) {
	var n int64
	_ = db.Model(&rag.RagKgEntityChunk{}).Where("entity_id = ? AND chunk_id = ?", entityID, chunkID).Count(&n).Error
	if n > 0 {
		return
	}
	if err := db.Create(&rag.RagKgEntityChunk{EntityID: entityID, ChunkID: chunkID}).Error; err != nil {
		return
	}
	kgTrimEntityChunkLinksIfNeeded(db, entityID)
}

func kgLinkRelChunk(db *gorm.DB, relID, chunkID uint) {
	var n int64
	_ = db.Model(&rag.RagKgRelationshipChunk{}).Where("relationship_id = ? AND chunk_id = ?", relID, chunkID).Count(&n).Error
	if n > 0 {
		return
	}
	if err := db.Create(&rag.RagKgRelationshipChunk{RelationshipID: relID, ChunkID: chunkID}).Error; err != nil {
		return
	}
	kgTrimRelationshipChunkLinksIfNeeded(db, relID)
}

func kgSyncEntityEmbedding(ctx context.Context, store interfaces.VectorStore, baseNS string, ent *rag.RagKgEntity, docPriority float64) error {
	if store == nil || ent == nil {
		return nil
	}
	ns := baseNS + "_kg_entity"
	text := ent.Name + "\n" + ent.EntityType + "\n" + ent.Description
	if ent.VectorStoreID != "" {
		_ = store.DeleteByIDs(ctx, []string{ent.VectorStoreID})
	}
	meta := map[string]any{
		"rag_kg_entity_id": ent.ID,
	}
	enrichMetadataRankBoostForChunk(meta, 0, 1)
	applyDocumentPriorityFloorToRankBoost(meta, docPriority)
	doc := schema.Document{PageContent: text, Metadata: meta}
	opts := []interfaces.VectorStoreOption{func(o *interfaces.VectorStoreOptions) { o.Namespace = ns }}
	ids, err := store.AddDocuments(ctx, []schema.Document{doc}, opts...)
	if err != nil || len(ids) == 0 {
		return err
	}
	ent.VectorStoreID = ids[0]
	return global.LRAG_DB.WithContext(ctx).Model(ent).Update("vector_store_id", ent.VectorStoreID).Error
}

func kgSyncRelEmbedding(ctx context.Context, store interfaces.VectorStore, baseNS string, rel *rag.RagKgRelationship, srcName, tgtName string, docPriority float64) error {
	if store == nil || rel == nil {
		return nil
	}
	ns := baseNS + "_kg_rel"
	text := srcName + " -> " + tgtName + "\n" + rel.Keywords + "\n" + rel.Description
	if rel.VectorStoreID != "" {
		_ = store.DeleteByIDs(ctx, []string{rel.VectorStoreID})
	}
	meta := map[string]any{
		"rag_kg_rel_id": rel.ID,
	}
	enrichMetadataRankBoostForChunk(meta, 0, 1)
	applyDocumentPriorityFloorToRankBoost(meta, docPriority)
	doc := schema.Document{PageContent: text, Metadata: meta}
	opts := []interfaces.VectorStoreOption{func(o *interfaces.VectorStoreOptions) { o.Namespace = ns }}
	ids, err := store.AddDocuments(ctx, []schema.Document{doc}, opts...)
	if err != nil || len(ids) == 0 {
		return err
	}
	rel.VectorStoreID = ids[0]
	return global.LRAG_DB.WithContext(ctx).Model(rel).Update("vector_store_id", rel.VectorStoreID).Error
}

// ScheduleKnowledgeGraphExtraction 文档索引成功后异步触发图谱抽取
func ScheduleKnowledgeGraphExtraction(uid uint, kbID uint, documentID uint) {
	go func() {
		cctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()
		var kb rag.RagKnowledgeBase
		if err := global.LRAG_DB.WithContext(cctx).First(&kb, kbID).Error; err != nil {
			return
		}
		if !kb.EnableKnowledgeGraph {
			return
		}
		if err := ExtractKnowledgeGraphForDocument(cctx, uid, &kb, documentID); err != nil {
			global.LRAG_LOG.Warn("知识图谱抽取任务失败", zap.Uint("docID", documentID), zap.Error(err))
		}
	}()
}
