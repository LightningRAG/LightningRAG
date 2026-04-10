package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/pageindex"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/retriever"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// MultiRetriever 多知识库合并检索器
type MultiRetriever struct {
	retrievers []interfaces.Retriever
	numDocs    int
}

// NewMultiRetriever 创建多知识库检索器
func NewMultiRetriever(retrievers []interfaces.Retriever, numDocs int) *MultiRetriever {
	if numDocs <= 0 {
		numDocs = 6
	}
	return &MultiRetriever{retrievers: retrievers, numDocs: numDocs}
}

// normalizeRetrieverBucketScoresForMerge 将单库检索结果的 Score 归一化到 [0,1]，便于多库合并排序（不同向量库/重排模型分数尺度不一致；借鉴 LightRAG 多源融合时对可比性的需求）
func normalizeRetrieverBucketScoresForMerge(docs []schema.Document) {
	if len(docs) == 0 {
		return
	}
	if len(docs) == 1 {
		docs[0].Score = 1
		return
	}
	minS, maxS := docs[0].Score, docs[0].Score
	for i := 1; i < len(docs); i++ {
		s := docs[i].Score
		if s < minS {
			minS = s
		}
		if s > maxS {
			maxS = s
		}
	}
	if maxS <= minS {
		for i := range docs {
			docs[i].Score = 1
		}
		return
	}
	span := maxS - minS
	for i := range docs {
		docs[i].Score = (docs[i].Score - minS) / span
	}
}

// GetRelevantDocuments 合并多知识库检索结果，按分数排序
func (m *MultiRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = m.numDocs
	}
	if len(m.retrievers) == 0 {
		return nil, nil
	}
	// references/ragflow/rag/nlp/search.py：Dealer.retrieval 将多个 kb_id 一次性传入检索，向量 leg 默认 topk=1024，
	// 再在融合分数上排序后取 page_size（dialog.top_n）条。若此处按 ceil(n/k) 均分每库条数，每库仅拿到极少候选，
	// 合并去重后常明显少于 n，且不是「跨库全局」最优 topN。
	// 因此多库时每库均按完整 n 拉取候选，合并排序后再截断到 n（代价为并行查询量随库数线性增加，与 Ragflow 单次宽召回等价目标一致）。
	perKB := n
	if perKB < MinRetrieveTopN {
		perKB = MinRetrieveTopN
	}
	g, gctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	buckets := make([][]schema.Document, len(m.retrievers))
	for i, r := range m.retrievers {
		i, r := i, r
		g.Go(func() error {
			docs, err := r.GetRelevantDocuments(gctx, query, perKB)
			if err != nil {
				return nil
			}
			mu.Lock()
			buckets[i] = docs
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()

	multiKB := len(m.retrievers) > 1
	best := make(map[string]schema.Document)
	var keyOrder []string
	ordered := make(map[string]bool)
	for _, docs := range buckets {
		if multiKB {
			normalizeRetrieverBucketScoresForMerge(docs)
		}
		for _, d := range docs {
			key := retriever.DocumentDedupKey(d)
			if old, ok := best[key]; !ok || d.Score > old.Score {
				best[key] = d
			}
			if !ordered[key] {
				ordered[key] = true
				keyOrder = append(keyOrder, key)
			}
		}
	}
	all := make([]schema.Document, 0, len(keyOrder))
	for _, key := range keyOrder {
		all = append(all, best[key])
	}
	// 按 Score 降序；同分稳定次序，保证与 refs / prompt 编号一致、可复现
	sort.SliceStable(all, func(i, j int) bool {
		si, sj := all[i].Score, all[j].Score
		if si != sj {
			return si > sj
		}
		li, lj := len(all[i].PageContent), len(all[j].PageContent)
		if li != lj {
			return li > lj
		}
		return all[i].PageContent < all[j].PageContent
	})
	if len(all) > n {
		all = all[:n]
	}
	return all, nil
}

// RetrieverType 实现接口
func (m *MultiRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypeVector
}

// RetrieverSessionFromLightningRAGParams 从各 API 共有字段构造检索会话选项，避免 Chat、queryData、知识库检索等入口逻辑漂移
func RetrieverSessionFromLightningRAGParams(modeOverride string, poolTopK *int, enableRerank *bool, cosineThreshold *float32, minRerankScore *float32) RetrieverSessionOptions {
	s := RetrieverSessionOptions{
		ModeOverride:     NormalizeLightningRAGRetrieverMode(strings.TrimSpace(modeOverride)),
		RetrievePoolTopK: poolTopK,
		RerankOverride:   enableRerank,
		CosineThreshold:  cosineThreshold,
		MinRerankScore:   minRerankScore,
	}
	if s.CosineThreshold == nil {
		if d := EffectiveDefaultCosineThreshold(); d > 0 {
			x := d
			s.CosineThreshold = &x
		}
	}
	return s
}

// ApplyPageIndexTocEnhanceFromRequest 将请求体中的 tocEnhance（Ragflow 同名语义）写入检索会话；nil 表示不覆盖默认
func ApplyPageIndexTocEnhanceFromRequest(session *RetrieverSessionOptions, tocEnhance *bool) {
	if session == nil || tocEnhance == nil {
		return
	}
	v := *tocEnhance
	session.PageIndexTocEnhance = &v
}

// ApplyDefaultConversationRetrievePoolIfNeeded 对话与 queryData 在请求未带 topK 时套用 config.default-conversation-retrieve-pool-top-k
func ApplyDefaultConversationRetrievePoolIfNeeded(session *RetrieverSessionOptions, finalChunkTopK int) {
	if session == nil || session.RetrievePoolTopK != nil {
		return
	}
	pool := EffectiveDefaultConversationRetrievePoolTopK()
	if pool <= 0 {
		return
	}
	maxCand := EffectiveMaxRetrieveCandidateTopK()
	if pool > maxCand {
		pool = maxCand
	}
	if pool <= finalChunkTopK {
		return
	}
	p := pool
	session.RetrievePoolTopK = &p
}

// RetrieverSessionOptions 单次检索会话选项（借鉴 references/LightRAG QueryParam.mode、enable_rerank）
type RetrieverSessionOptions struct {
	// ModeOverride 非空时覆盖知识库 RetrieverType
	ModeOverride string
	// RerankOverride nil=按知识库 UseRerank；false=强制关闭 Rerank；true=已配置 Rerank 模型时启用（可覆盖知识库 UseRerank=false）
	RerankOverride *bool
	// CosineThreshold 向量相似度下限（与 pgvector/MySQL 等返回的 Score 一致，约 0~1）；nil 或 <=0 表示不过滤。请求未传时可能由 config.rag.default-cosine-threshold 填充（对齐 LightRAG）
	CosineThreshold *float32
	// MinRerankScore Rerank 后分数下限；nil 或 <=0 表示不过滤（分数尺度依赖 Rerank 提供商）
	MinRerankScore *float32
	// RetrievePoolTopK 对齐 LightningRAG QueryParam.top_k：扩大向量/融合/Rerank 前的候选池；最终注入条数仍由请求的 chunkTopK / topN 决定
	RetrievePoolTopK *int
	// KgEntitySearchQuery 非空时，知识图谱 local 路径用其做实体向量检索（通常为低层关键词 + 问句；对齐 LightRAG ll_keywords）
	KgEntitySearchQuery string
	// KgRelSearchQuery 非空时，知识图谱 global 路径用其做关系向量检索（通常为高层关键词 + 问句；对齐 LightRAG hl_keywords）
	KgRelSearchQuery string
	// PageIndexTocEnhance 对齐 references/ragflow 的 toc_enhance：PageIndex 知识库是否在向量命中后做 TOC 加权/补全。
	// nil 或未设置=启用混合（与当前默认一致）；显式 false=仅 TOC/Tree，不走向量+TOC。
	PageIndexTocEnhance *bool
}

// CreateRetrieverForKnowledgeBases 根据知识库 ID 列表创建检索器
// session.ModeOverride 与 LightningRAG API 的 mode 一致；空则使用各库配置
// llm 可选，用于 PageIndex 推理检索；为 nil 时 pageindex 知识库会尝试使用首个启用的管理员 LLM
func CreateRetrieverForKnowledgeBases(ctx context.Context, kbIDs []uint, userID uint, numDocs int, session RetrieverSessionOptions, llm ...interfaces.LLM) (interfaces.Retriever, error) {
	var llmInst interfaces.LLM
	if len(llm) > 0 {
		llmInst = llm[0]
	}
	if len(kbIDs) == 0 {
		return nil, nil
	}
	g, gctx := errgroup.WithContext(ctx)
	buckets := make([]interfaces.Retriever, len(kbIDs))
	var lastErr error
	var errMu sync.Mutex
	for i, kbID := range kbIDs {
		i, kbID := i, kbID
		g.Go(func() error {
			r, err := createRetrieverForKB(gctx, kbID, userID, numDocs, llmInst, session)
			if err != nil {
				errMu.Lock()
				lastErr = err
				errMu.Unlock()
				global.LRAG_LOG.Warn("创建知识库检索器失败", zap.Uint("kbID", kbID), zap.Error(err))
				return nil
			}
			buckets[i] = r
			return nil
		})
	}
	_ = g.Wait()
	var retrievers []interfaces.Retriever
	for _, r := range buckets {
		if r != nil {
			retrievers = append(retrievers, r)
		}
	}
	if len(retrievers) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("所有知识库检索器创建失败: %w", lastErr)
		}
		return nil, nil
	}
	if len(retrievers) == 1 {
		return retrievers[0], nil
	}
	return NewMultiRetriever(retrievers, numDocs), nil
}

// resolveVectorDimensions 解析向量维度：优先用配置值，其次 embedder 声明值，最后通过 trial embed 检测（如 Ollama）
func resolveVectorDimensions(ctx context.Context, embedder interfaces.Embedder, configuredDims int) int {
	if configuredDims > 0 {
		return configuredDims
	}
	if embedder != nil {
		if d := embedder.Dimensions(); d > 0 {
			return d
		}
		// 配置和 embedder 均未声明维度时，通过 trial embed 检测（Ollama 等）
		if vec, err := embedder.EmbedQuery(ctx, "test"); err == nil && len(vec) > 0 {
			return len(vec)
		}
	}
	return 1536 // 默认
}

// defaultVectorStoreProvider 根据系统配置的数据库类型返回默认向量存储提供商
func defaultVectorStoreProvider() string {
	dbType := strings.ToLower(global.LRAG_CONFIG.System.DbType)
	switch dbType {
	case "pgsql", "postgresql", "postgres":
		return "postgresql"
	case "mysql", "":
		return "mysql"
	default:
		return "mysql"
	}
}

// createVectorStoreFromKB 根据知识库的 VectorStoreID 创建向量存储，支持 mysql、postgresql、elasticsearch 等
func createVectorStoreFromKB(ctx context.Context, kb *rag.RagKnowledgeBase, embedder interfaces.Embedder, namespace string, vectorDims int) (interfaces.VectorStore, error) {
	vectorDims = resolveVectorDimensions(ctx, embedder, vectorDims)
	vsConfig := registry.VectorStoreConfig{
		Provider: defaultVectorStoreProvider(),
		Config:   nil,
	}
	if kb.VectorStoreID > 0 {
		var cfg rag.RagVectorStoreConfig
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", kb.VectorStoreID).First(&cfg).Error; err != nil {
			return nil, fmt.Errorf("向量存储配置不存在: %w", err)
		}
		vsConfig.Provider = cfg.Provider
		if cfg.Provider == "" {
			vsConfig.Provider = defaultVectorStoreProvider()
		}
		if cfg.Config != nil {
			vsConfig.Config = map[string]any(cfg.Config)
		}
	}
	store, err := registry.CreateVectorStore(vsConfig, embedder, namespace, vectorDims)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, fmt.Errorf("不支持的向量存储类型: %s", vsConfig.Provider)
	}
	return store, nil
}

func createRetrieverForKB(ctx context.Context, kbID uint, userID uint, numDocs int, llm interfaces.LLM, session RetrieverSessionOptions) (interfaces.Retriever, error) {
	var kb rag.RagKnowledgeBase
	// 先尝试用户自己的知识库，再尝试全局共享知识库（不限 owner）
	err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", kbID, userID).First(&kb).Error
	if err != nil {
		// 检查是否为全局共享知识库
		var gkb rag.RagGlobalKnowledgeBase
		if gerr := global.LRAG_DB.WithContext(ctx).Where("knowledge_base_id = ?", kbID).First(&gkb).Error; gerr != nil {
			return nil, err
		}
		if err2 := global.LRAG_DB.WithContext(ctx).Where("id = ?", kbID).First(&kb).Error; err2 != nil {
			return nil, err2
		}
	}
	rt := NormalizeLightningRAGRetrieverMode(session.ModeOverride)
	if rt == "" {
		rt = strings.ToLower(strings.TrimSpace(kb.RetrieverType))
		if rt == "" {
			rt = "vector"
		}
	}
	if rt == "naive" {
		rt = "vector"
	}
	// PageIndex 推理检索（与 Ragflow 一致：经 retrievalEnabledFilterRetriever 包装，并尽量启用「向量 + TOC」混合）
	if rt == "pageindex" {
		inner, err := createPageIndexRetriever(ctx, kbID, userID, numDocs, &kb, llm, session)
		if err != nil {
			return nil, err
		}
		return newRetrievalEnabledFilterRetriever(kbID, inner, numDocs), nil
	}
	// LightningRAG bypass：不拉取知识库上下文
	if rt == "bypass" {
		return newRetrievalEnabledFilterRetriever(kbID, retriever.NewBypassRetriever(), numDocs), nil
	}
	emb, err := resolveEmbeddingConfig(ctx, &kb, userID)
	if err != nil {
		return nil, fmt.Errorf("嵌入模型配置不存在: %w", err)
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		return nil, fmt.Errorf("创建嵌入模型失败: %v", err)
	}
	ns := "kb_" + strconv.FormatUint(uint64(kbID), 10)
	store, err := createVectorStoreFromKB(ctx, &kb, embedder, ns, emb.Dimensions)
	if err != nil {
		return nil, err
	}

	kgReady := kb.EnableKnowledgeGraph && KnowledgeGraphHasEntities(ctx, kb.ID)
	kgEntQ := strings.TrimSpace(session.KgEntitySearchQuery)
	kgRelQ := strings.TrimSpace(session.KgRelSearchQuery)

	var ret interfaces.Retriever
	switch rt {
	case "keyword":
		kr := retriever.NewKeywordRetriever(store, ns, numDocs)
		if EffectiveKeywordEmptyRetry() {
			ret = retriever.WrapKeywordEmptyRetry(kr)
		} else {
			ret = kr
		}
	case "local":
		if kgReady {
			ret = retriever.NewKgLocalGraphRetriever(store, ns, kb.ID, numDocs, kgEntQ)
		} else {
			kr := retriever.NewKeywordRetriever(store, ns, numDocs, interfaces.RetrieverTypeLocal)
			if EffectiveKeywordEmptyRetry() {
				ret = retriever.WrapKeywordEmptyRetry(kr)
			} else {
				ret = kr
			}
		}
	case "global":
		if kgReady {
			ret = retriever.NewKgGlobalVectorMixRetriever(
				retriever.NewVectorRetriever(store, ns, numDocs, interfaces.RetrieverTypeGlobal),
				retriever.NewKgGlobalGraphRetriever(store, ns, kb.ID, numDocs, kgRelQ),
				numDocs,
			)
		} else {
			ret = retriever.NewVectorRetriever(store, ns, numDocs, interfaces.RetrieverTypeGlobal)
		}
	case "hybrid":
		if kgReady {
			ret = retriever.NewKgHybridGraphRetriever(
				retriever.NewKgLocalGraphRetriever(store, ns, kb.ID, numDocs, kgEntQ),
				retriever.NewKgGlobalGraphRetriever(store, ns, kb.ID, numDocs, kgRelQ),
				numDocs,
			)
		} else {
			tw, vw := EffectiveHybridFusionWeights()
			ret = retriever.NewFusionRetriever(
				retriever.NewVectorRetriever(store, ns, numDocs),
				retriever.NewKeywordRetriever(store, ns, numDocs),
				retriever.FusionHybrid,
				numDocs,
				interfaces.RetrieverTypeHybrid,
				retriever.WithFusionWeights(tw, vw),
				retriever.WithFusionMinScore(EffectiveHybridFusionMinScore()),
				retriever.WithFusionEmptyRetry(EffectiveHybridFusionEmptyRetry()),
			)
		}
	case "mix":
		if kgReady {
			ret = retriever.NewKgMixRetriever(
				retriever.NewVectorRetriever(store, ns, numDocs),
				retriever.NewKgHybridGraphRetriever(
					retriever.NewKgLocalGraphRetriever(store, ns, kb.ID, numDocs, kgEntQ),
					retriever.NewKgGlobalGraphRetriever(store, ns, kb.ID, numDocs, kgRelQ),
					numDocs,
				),
				numDocs,
			)
		} else {
			tw, vw := EffectiveHybridFusionWeights()
			ret = retriever.NewFusionRetriever(
				retriever.NewVectorRetriever(store, ns, numDocs),
				retriever.NewKeywordRetriever(store, ns, numDocs),
				retriever.FusionMix,
				numDocs,
				interfaces.RetrieverTypeMix,
				retriever.WithFusionWeights(tw, vw),
				retriever.WithFusionMinScore(EffectiveHybridFusionMinScore()),
				retriever.WithFusionEmptyRetry(EffectiveHybridFusionEmptyRetry()),
			)
		}
	default:
		ret = retriever.NewVectorRetriever(store, ns, numDocs)
	}

	if ct := session.CosineThreshold; ct != nil && *ct > 0 {
		switch v := ret.(type) {
		case *retriever.VectorRetriever:
			v.WithScoreThreshold(*ct)
		case *retriever.FusionRetriever:
			v.WithVectorScoreThreshold(*ct)
		case *retriever.KgGraphRetriever:
			v.WithScoreThreshold(*ct)
		case *retriever.KgHybridGraphRetriever:
			v.WithScoreThreshold(*ct)
		case *retriever.KgMixRetriever:
			v.WithScoreThreshold(*ct)
		case *retriever.KgGlobalVectorMixRetriever:
			v.WithScoreThreshold(*ct)
		}
	}

	if vr, ok := ret.(*retriever.VectorRetriever); ok && EffectiveVectorEmptyRetry() {
		ret = retriever.WrapVectorEmptyRetry(vr)
	}

	wantRerank := kb.UseRerank && kb.RerankID > 0
	if session.RerankOverride != nil {
		if !*session.RerankOverride {
			wantRerank = false
		} else {
			wantRerank = kb.RerankID > 0
		}
	}
	if wantRerank {
		reranker, rerankErr := createRerankerFromKB(ctx, &kb, userID)
		if rerankErr != nil {
			global.LRAG_LOG.Warn("创建 Rerank 模型失败，将使用原始检索结果",
				zap.Uint("kbID", kbID), zap.Error(rerankErr))
		} else if reranker != nil {
			var minRS float32
			if session.MinRerankScore != nil && *session.MinRerankScore > 0 {
				minRS = *session.MinRerankScore
			}
			ret = retriever.NewRerankRetriever(ret, reranker, kb.RerankTopK, minRS)
		}
	}

	return newRetrievalEnabledFilterRetriever(kbID, ret, numDocs), nil
}

// createRerankerFromKB 根据知识库配置创建 Reranker
func createRerankerFromKB(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) (interfaces.Reranker, error) {
	if kb.RerankID == 0 {
		return nil, nil
	}
	source := kb.RerankSource
	if source == "" {
		source = "admin"
	}
	switch source {
	case "user":
		var m rag.RagUserLLM
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", kb.RerankID, userID).First(&m).Error; err != nil {
			return nil, fmt.Errorf("用户 Rerank 模型不存在(ID=%d): %w", kb.RerankID, err)
		}
		return registry.CreateRerank(registry.RerankConfig{
			Provider:  m.Provider,
			ModelName: m.ModelName,
			BaseURL:   m.BaseURL,
			APIKey:    m.APIKey,
		})
	default:
		var m rag.RagLLMProvider
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND enabled = ?", kb.RerankID, true).First(&m).Error; err != nil {
			return nil, fmt.Errorf("管理员 Rerank 模型不存在(ID=%d): %w", kb.RerankID, err)
		}
		return registry.CreateRerank(registry.RerankConfig{
			Provider:  m.Name,
			ModelName: m.ModelName,
			BaseURL:   m.BaseURL,
			APIKey:    m.APIKey,
		})
	}
}

// createPageIndexVectorLeg 为 PageIndex 知识库创建向量检索腿（与 references/ragflow 中先向量命中再 TOC 增强一致）
func createPageIndexVectorLeg(ctx context.Context, kb *rag.RagKnowledgeBase, kbID, userID uint, numDocs int, session RetrieverSessionOptions) *retriever.VectorRetriever {
	emb, err := resolveEmbeddingConfig(ctx, kb, userID)
	if err != nil {
		return nil
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		return nil
	}
	ns := "kb_" + strconv.FormatUint(uint64(kbID), 10)
	store, err := createVectorStoreFromKB(ctx, kb, embedder, ns, emb.Dimensions)
	if err != nil || store == nil {
		return nil
	}
	vr := retriever.NewVectorRetriever(store, ns, numDocs)
	if ct := session.CosineThreshold; ct != nil && *ct > 0 {
		vr.WithScoreThreshold(*ct)
	}
	return vr
}

func createPageIndexRetriever(ctx context.Context, kbID, userID uint, numDocs int, kb *rag.RagKnowledgeBase, llm interfaces.LLM, session RetrieverSessionOptions) (interfaces.Retriever, error) {
	var docs []rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("knowledge_base_id = ? AND status = ? AND retrieval_enabled = ?", kbID, "completed", true).
		Find(&docs).Error; err != nil {
		return nil, err
	}
	var docTrees []retriever.DocTree
	for i := range docs {
		d := &docs[i]
		tree, err := resolvePageIndexTreeForRetrieve(ctx, d, kb, userID)
		if err != nil {
			global.LRAG_LOG.Warn("解析或构建 PageIndex 树失败", zap.Uint("docID", d.ID), zap.Error(err))
			continue
		}
		if len(tree) == 0 {
			continue
		}
		nodeList := pageindex.StructureToList(tree)
		var chunks []rag.RagChunk
		_ = global.LRAG_DB.WithContext(ctx).Where("document_id = ?", d.ID).Order("chunk_index ASC").Find(&chunks).Error
		nodeMap := make(map[string]uint)
		lim := len(nodeList)
		if len(chunks) < lim {
			lim = len(chunks)
		}
		for j := 0; j < lim; j++ {
			if nodeList[j] == nil {
				continue
			}
			nodeMap[pageindex.NormalizeNodeID(nodeList[j].NodeID)] = chunks[j].ID
		}
		docTrees = append(docTrees, retriever.DocTree{
			DocID:            d.ID,
			DocName:          d.Name,
			Tree:             tree,
			NodeToRagChunkID: nodeMap,
		})
	}
	if len(docTrees) == 0 {
		return nil, fmt.Errorf("知识库无有效 PageIndex 文档，请上传 md/txt 并确保 RetrieverType 为 pageindex")
	}
	// 若未传入 LLM，按优先级解析：知识库配置 > 首个启用的管理员 LLM
	pureLLM := llm
	if pureLLM == nil {
		pureLLM = resolvePageIndexLLM(ctx, kb, userID)
	}
	if pureLLM == nil {
		return nil, fmt.Errorf("PageIndex 检索需要 LLM，请在知识库中配置 PageIndex LLM 或配置管理员 LLM")
	}
	pure := retriever.NewPageIndexRetriever(docTrees, pureLLM, numDocs)
	useHybrid := true
	if session.PageIndexTocEnhance != nil && !*session.PageIndexTocEnhance {
		useHybrid = false
	}
	if vr := createPageIndexVectorLeg(ctx, kb, kbID, userID, numDocs, session); vr != nil && useHybrid {
		// Ragflow：Dealer.retrieval 在 rerank 之后再跑 retrieval_by_toc；向量腿在 TOC 前套一层 Rerank（与 createRetrieverForKB 一致）
		vecBase := interfaces.Retriever(vr)
		if EffectiveVectorEmptyRetry() {
			vecBase = retriever.WrapVectorEmptyRetry(vr)
		}
		vecChain := vecBase
		wantRerank := kb.UseRerank && kb.RerankID > 0
		if session.RerankOverride != nil {
			if !*session.RerankOverride {
				wantRerank = false
			} else {
				wantRerank = kb.RerankID > 0
			}
		}
		if wantRerank {
			if reranker, rerankErr := createRerankerFromKB(ctx, kb, userID); rerankErr != nil {
				global.LRAG_LOG.Warn("PageIndex 向量腿：创建 Rerank 模型失败，跳过重排",
					zap.Uint("kbID", kbID), zap.Error(rerankErr))
			} else if reranker != nil {
				var minRS float32
				if session.MinRerankScore != nil && *session.MinRerankScore > 0 {
					minRS = *session.MinRerankScore
				}
				vecChain = retriever.NewRerankRetriever(vecBase, reranker, kb.RerankTopK, minRS)
			}
		}
		// Ragflow：retrieval_by_toc 使用租户默认 Chat；无会话传入模型时用 Chat 解析链，不用 KB 独占 PageIndex 模型覆盖
		tocLLM := llm
		if tocLLM == nil {
			tocLLM = resolveRagflowTocRelevanceLLM(ctx, userID)
		}
		if tocLLM == nil {
			tocLLM = pureLLM
		}
		return retriever.NewPageIndexRagflowRetriever(vecChain, pure, tocLLM), nil
	}
	return pure, nil
}

// resolvePageIndexTreeForRetrieve 优先使用已持久化的树；为空或损坏时按 references/PageIndex 思路从源文件重建（并回写 DB），否则用 rag_chunks 合成浅层树
func resolvePageIndexTreeForRetrieve(ctx context.Context, doc *rag.RagDocument, kb *rag.RagKnowledgeBase, userID uint) ([]pageindex.TreeNode, error) {
	if s := strings.TrimSpace(doc.PageIndexStructure); s != "" {
		var tree []pageindex.TreeNode
		if err := json.Unmarshal([]byte(s), &tree); err == nil && len(tree) > 0 {
			return tree, nil
		}
	}
	return fallbackPageIndexTreeForRetrieve(ctx, doc, kb, userID)
}

func fallbackPageIndexTreeForRetrieve(ctx context.Context, doc *rag.RagDocument, kb *rag.RagKnowledgeBase, userID uint) ([]pageindex.TreeNode, error) {
	if doc.StoragePath != "" {
		data, err := loadDocumentFileBytes(doc)
		if err == nil {
			content, err := parseDocumentContent(ctx, data, doc.FileType, doc.Name, userID, kb)
			if err == nil && strings.TrimSpace(content) != "" {
				tree := pageindex.BuildTreeForPageIndex(doc.FileType, doc.Name, content)
				if len(tree) > 0 {
					if b, mErr := json.Marshal(tree); mErr == nil {
						_ = global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
							Where("id = ? AND (COALESCE(page_index_structure,'') = '')", doc.ID).
							Update("page_index_structure", string(b)).Error
					}
					return tree, nil
				}
			}
		}
	}
	var chunks []rag.RagChunk
	if err := global.LRAG_DB.WithContext(ctx).Where("document_id = ?", doc.ID).Order("chunk_index ASC").Find(&chunks).Error; err != nil {
		return nil, err
	}
	var parts []string
	for _, c := range chunks {
		if t := strings.TrimSpace(c.Content); t != "" {
			parts = append(parts, t)
		}
	}
	tree := pageindex.BuildTreeFromTextChunks(doc.Name, parts)
	if len(tree) == 0 {
		return nil, nil
	}
	return tree, nil
}

// resolveRagflowTocRelevanceLLM 对齐 references/ragflow：retrieval_by_toc / relevant_chunks_with_toc 使用租户默认 Chat 模型（不读取知识库 PageIndexLLM 覆盖）。
func resolveRagflowTocRelevanceLLM(ctx context.Context, userID uint) interfaces.LLM {
	provider, modelName, baseURL, apiKey, ok := ResolveModelWithFallback(ctx, userID, 0, 0, "", interfaces.ModelTypeChat)
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

// resolvePageIndexLLM 解析 PageIndex 推理检索用 LLM：KB配置 → 用户默认 → 系统全局默认 → 首个管理员 LLM
func resolvePageIndexLLM(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) interfaces.LLM {
	var kbModelID uint
	var kbModelSource string
	if kb != nil {
		kbModelID = kb.PageIndexLLMID
		kbModelSource = kb.PageIndexLLMSource
	}
	provider, modelName, baseURL, apiKey, ok := ResolveModelWithFallback(ctx, userID, 0, kbModelID, kbModelSource, interfaces.ModelTypeChat)
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

// ParseKnowledgeBaseIDs 从 SourceIDs JSON 解析知识库 ID 列表
// 支持 ["1","2"] 或 ["kb:1","kb:2"] 格式
// 仅当 sourceType 为 knowledge_base（忽略大小写与首尾空格）时解析；files 等类型当前不走知识库向量检索链路
func ParseKnowledgeBaseIDs(sourceType, sourceIDs string) ([]uint, error) {
	st := strings.TrimSpace(strings.ToLower(sourceType))
	if st != "knowledge_base" || strings.TrimSpace(sourceIDs) == "" {
		return nil, nil
	}
	sourceIDs = strings.TrimSpace(sourceIDs)
	if sourceIDs == "[]" {
		return nil, nil
	}
	var raw []any
	if err := json.Unmarshal([]byte(sourceIDs), &raw); err != nil {
		return nil, err
	}
	var ids []uint
	for _, v := range raw {
		switch x := v.(type) {
		case string:
			s := strings.TrimPrefix(x, "kb:")
			n, err := strconv.ParseUint(s, 10, 32)
			if err == nil {
				ids = append(ids, uint(n))
			}
		case float64:
			if x >= 0 && x == float64(uint(x)) {
				ids = append(ids, uint(x))
			}
		}
	}
	return ids, nil
}
