package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
	"go.uber.org/zap"
)

// chunksMapsFromDocs 构建 LightningRAG query_data 风格的 chunks 列表
func chunksMapsFromDocs(docs []ragschema.Document, includeContent bool) []map[string]any {
	out := make([]map[string]any, 0, len(docs))
	for i, d := range docs {
		m := map[string]any{
			"index": i,
			"score": d.Score,
		}
		if includeContent {
			m["content"] = d.PageContent
		}
		if d.Metadata != nil {
			m["metadata"] = d.Metadata
		}
		out = append(out, m)
	}
	return out
}

// QueryData 对齐 references/LightRAG /query/data：仅结构化检索，不调用 LLM；entities/relationships 由检索命中切片关联的图谱填充（无图谱或无关数据时为空数组）
func (s *ConversationService) QueryData(ctx context.Context, uid uint, req request.ConversationQueryData) (map[string]any, error) {
	qin := strings.TrimSpace(req.Query)
	if len(qin) < 3 {
		return nil, fmt.Errorf("query 至少需要 3 个字符")
	}
	var conv rag.RagConversation
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", req.ConversationID, uid).First(&conv).Error; err != nil {
		return nil, err
	}
	cfg, err := resolveLLMConfig(ctx, uid, &conv, req.LLMProviderID, req.LLMSource)
	if err != nil {
		return nil, err
	}
	llmInst, err := registry.CreateLLM(registry.LLMConfig{
		Provider:  cfg.Provider,
		ModelName: cfg.ModelName,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
	})
	if err != nil {
		return nil, err
	}
	if llmInst == nil {
		return nil, fmt.Errorf("不支持的模型提供商: %s", cfg.Provider)
	}

	kbIDs, parseErr := ParseKnowledgeBaseIDs(conv.SourceType, conv.SourceIDs)
	if parseErr != nil {
		global.LRAG_LOG.Warn("queryData 解析 sourceIds 失败", zap.String("sourceIDs", conv.SourceIDs), zap.Error(parseErr))
	}
	kbIDs = mergeGlobalKnowledgeBaseIDs(ctx, kbIDs)
	q, modeOv, _ := ResolveLightningRAGQueryModeAndQuestion(qin, req.QueryMode)
	effectiveMode := modeOv
	if effectiveMode == "" {
		effectiveMode = "kb_default"
	}
	session := RetrieverSessionFromLightningRAGParams(modeOv, req.TopK, req.EnableRerank, req.CosineThreshold, req.MinRerankScore)
	ApplyPageIndexTocEnhanceFromRequest(&session, request.EffectiveTocEnhance(req.TocEnhance, req.TocEnhanceRagflow))
	combined, kgEnt, kgRel, resolvedHl, resolvedLl := PrepareLightningRAGSearchQueries(ctx, uid, llmInst, q, req.HlKeywords, req.LlKeywords, keywordHistoryHint(req.ConversationHistory))
	session.KgEntitySearchQuery = kgEnt
	session.KgRelSearchQuery = kgRel
	searchQ := combined
	ragTopK := ClampRetrieveTopN(req.ChunkTopK, DefaultConversationChunkTopKFromConfig())
	ApplyDefaultConversationRetrievePoolIfNeeded(&session, ragTopK)

	var docs []ragschema.Document
	if len(kbIDs) > 0 && strings.TrimSpace(q) != "" {
		var rerr error
		docs, rerr = fetchRelevantDocumentsForKnowledgeBases(ctx, kbIDs, uid, llmInst, searchQ, ragTopK, session)
		if rerr != nil {
			return nil, rerr
		}
		maxRag := EffectiveMaxRagContextTokens(req.MaxRagContextTokens)
		docs = trimDocsToRagTokenBudget(docs, maxRag)
	}

	includeChunk := true
	if req.IncludeChunkContent != nil && !*req.IncludeChunkContent {
		includeChunk = false
	}
	chunks := chunksMapsFromDocs(docs, includeChunk)
	refs := ragDocumentsToRefMaps(docs)
	refsOut := ExposeReferencesForAPI(refs, req.IncludeReferences, req.IncludeChunkContent)

	chunkIDs := ChunkIDsFromRAGDocs(ctx, kbIDs, docs)
	entMaps, relMaps := KnowledgeGraphMapsForChunkIDs(ctx, kbIDs, chunkIDs)
	entitiesAny := make([]any, 0, len(entMaps))
	for _, m := range entMaps {
		entitiesAny = append(entitiesAny, m)
	}
	relsAny := make([]any, 0, len(relMaps))
	for _, m := range relMaps {
		relsAny = append(relsAny, m)
	}

	data := map[string]any{
		"entities":        entitiesAny,
		"relationships":   relsAny,
		"chunks":          chunks,
		"references":      refsOut,
		"contextPrompt":   "",
		"displayQuestion": q,
	}
	if len(docs) > 0 {
		maxEt := EffectiveMaxEntityContextTokens(req.MaxEntityTokens)
		maxRt := EffectiveMaxRelationContextTokens(req.MaxRelationTokens)
		graphP := ""
		if maxEt > 0 || maxRt > 0 {
			graphP = FormatKnowledgeGraphPromptPrefix(entMaps, relMaps, maxEt, maxRt)
		}
		data["contextPrompt"] = buildRAGContextPrompt(docs, q, graphP)
	}

	meta := map[string]any{
		"retrievalMode":    effectiveMode,
		"retrievalQuery":   q,
		"searchQuery":      searchQ,
		"hlKeywords":       KeywordsForAPIResponse(resolvedHl),
		"llKeywords":       KeywordsForAPIResponse(resolvedLl),
		"chunkCount":       len(docs),
		"knowledgeBaseIds": kbIDs,
	}
	return map[string]any{
		"status":   "success",
		"message":  "检索完成",
		"data":     data,
		"metadata": meta,
	}, nil
}
