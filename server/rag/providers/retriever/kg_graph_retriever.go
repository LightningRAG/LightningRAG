package retriever

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// KgGraphRetriever 基于知识图谱实体或关系向量的检索（对齐 LightRAG local / global）
type KgGraphRetriever struct {
	store          interfaces.VectorStore
	baseNS         string
	kbID           uint
	numDocs        int
	useRels        bool // true: 关系向量（global）；false: 实体向量（local）
	retType        interfaces.RetrieverType
	searchOverride string  // 非空时忽略传入的 query，用于 LightRAG 式 ll/hl 分流检索
	scoreThreshold float32 // 实体/关系向量相似度下限，0 表示不过滤（对齐 QueryParam 类 cosine 阈值）
}

// WithScoreThreshold 设置图谱向量相似度下限；t<=0 时关闭过滤
func (r *KgGraphRetriever) WithScoreThreshold(t float32) *KgGraphRetriever {
	if t > 0 {
		r.scoreThreshold = t
	}
	return r
}

// NewKgLocalGraphRetriever 实体上下文检索（local）；searchOverride 一般为低层关键词增强问句
func NewKgLocalGraphRetriever(store interfaces.VectorStore, baseNS string, kbID uint, numDocs int, searchOverride string) *KgGraphRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	return &KgGraphRetriever{
		store:          store,
		baseNS:         baseNS,
		kbID:           kbID,
		numDocs:        numDocs,
		useRels:        false,
		retType:        interfaces.RetrieverTypeLocal,
		searchOverride: strings.TrimSpace(searchOverride),
	}
}

// NewKgGlobalGraphRetriever 关系/全局知识检索（global）；searchOverride 一般为高层关键词增强问句
func NewKgGlobalGraphRetriever(store interfaces.VectorStore, baseNS string, kbID uint, numDocs int, searchOverride string) *KgGraphRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	return &KgGraphRetriever{
		store:          store,
		baseNS:         baseNS,
		kbID:           kbID,
		numDocs:        numDocs,
		useRels:        true,
		retType:        interfaces.RetrieverTypeGlobal,
		searchOverride: strings.TrimSpace(searchOverride),
	}
}

// GetRelevantDocuments 先检索实体/关系向量，再展开到原文切片
func (r *KgGraphRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = r.numDocs
	}
	// 实体/关系向量：对齐 LightRAG QueryParam.top_k 量级并随最终条数放大，便于展开到足够切片
	fetch := lightragconst.WideSimilarityFetchK(n, 4)
	if r.scoreThreshold > 0 {
		fetch = lightragconst.WideSimilarityFetchK(n, 8)
	}
	ns := r.baseNS + "_kg_entity"
	if r.useRels {
		ns = r.baseNS + "_kg_rel"
	}
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = ns },
	}
	q := strings.TrimSpace(query)
	if r.searchOverride != "" {
		q = r.searchOverride
	}
	hits, err := r.store.SimilaritySearch(ctx, q, fetch, opts...)
	if err != nil {
		return nil, err
	}
	if r.scoreThreshold > 0 {
		filtered := make([]schema.Document, 0, len(hits))
		for _, d := range hits {
			if d.Score >= r.scoreThreshold {
				filtered = append(filtered, d)
			}
		}
		hits = filtered
	}
	if len(hits) == 0 && q != "" && !global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry {
		retryK := lightragconst.MaxSimilarityFetchK
		hits2, err2 := r.store.SimilaritySearch(ctx, q, retryK, opts...)
		if err2 != nil {
			return nil, err2
		}
		hits = hits2
	}
	return r.hitsToChunkDocuments(ctx, hits, n)
}

func (r *KgGraphRetriever) hitsToChunkDocuments(ctx context.Context, hits []schema.Document, limit int) ([]schema.Document, error) {
	chunkScore := make(map[uint]float32)
	db := global.LRAG_DB.WithContext(ctx)
	for _, h := range hits {
		var graphID uint
		if r.useRels {
			graphID = metadataAsUint(h.Metadata, "rag_kg_rel_id")
		} else {
			graphID = metadataAsUint(h.Metadata, "rag_kg_entity_id")
		}
		if graphID == 0 {
			continue
		}
		var cids []uint
		var q *gorm.DB
		if r.useRels {
			q = db.Model(&rag.RagKgRelationshipChunk{}).Where("relationship_id = ?", graphID).Order("chunk_id ASC")
		} else {
			q = db.Model(&rag.RagKgEntityChunk{}).Where("entity_id = ?", graphID).Order("chunk_id ASC")
		}
		if err := q.Pluck("chunk_id", &cids).Error; err != nil {
			continue
		}
		if mx := effectiveKgRetrieverRelatedChunkNumber(); mx > 0 && len(cids) > mx {
			cids = cids[:mx]
		}
		for _, cid := range cids {
			if prev, ok := chunkScore[cid]; !ok || h.Score > prev {
				chunkScore[cid] = h.Score
			}
		}
	}
	if len(chunkScore) == 0 {
		return nil, nil
	}
	type scored struct {
		id    uint
		score float32
	}
	var ranked []scored
	for id, s := range chunkScore {
		ranked = append(ranked, scored{id: id, score: s})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score > ranked[j].score
		}
		return ranked[i].id < ranked[j].id
	})
	// 原 limit*3 易在「多实体→多 chunk_id」时过早截断，导致最终条数远小于 limit（对齐 LightRAG related_chunk_number × 实体数 的规模）
	rankCap := limit * 12
	if minCap := lightragconst.DefaultTopK * 4; rankCap < minCap {
		rankCap = minCap
	}
	if rankCap > lightragconst.MaxSimilarityFetchK {
		rankCap = lightragconst.MaxSimilarityFetchK
	}
	if len(ranked) > rankCap {
		ranked = ranked[:rankCap]
	}
	ids := make([]uint, len(ranked))
	for i := range ranked {
		ids[i] = ranked[i].id
	}
	var chunks []rag.RagChunk
	if err := db.Where("id IN ?", ids).Find(&chunks).Error; err != nil {
		return nil, err
	}
	chunkByID := make(map[uint]rag.RagChunk, len(chunks))
	for _, c := range chunks {
		chunkByID[c.ID] = c
	}
	var docIDs []uint
	seenDoc := make(map[uint]bool)
	for _, c := range chunks {
		if !seenDoc[c.DocumentID] {
			seenDoc[c.DocumentID] = true
			docIDs = append(docIDs, c.DocumentID)
		}
	}
	var docsMeta []rag.RagDocument
	if len(docIDs) > 0 {
		_ = db.Where("id IN ? AND knowledge_base_id = ?", docIDs, r.kbID).Find(&docsMeta).Error
	}
	docName := make(map[uint]string)
	for _, d := range docsMeta {
		docName[d.ID] = d.Name
	}
	out := make([]schema.Document, 0, len(ranked))
	src := "kg_local"
	if r.useRels {
		src = "kg_global"
	}
	for _, sc := range ranked {
		ch, ok := chunkByID[sc.id]
		if !ok {
			continue
		}
		if _, ok := docName[ch.DocumentID]; !ok {
			continue
		}
		out = append(out, schema.Document{
			PageContent: ch.Content,
			Score:       sc.score,
			Metadata: map[string]any{
				"document_id":  ch.DocumentID,
				"chunk_index":  ch.ChunkIndex,
				"doc_name":     docName[ch.DocumentID],
				"rag_chunk_id": ch.ID,
				"kg_source":    src,
			},
		})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// RetrieverType 返回 local 或 global
func (r *KgGraphRetriever) RetrieverType() interfaces.RetrieverType {
	return r.retType
}

// KgHybridGraphRetriever 合并 local + global 图谱展开结果（对齐 LightRAG hybrid）
type KgHybridGraphRetriever struct {
	local   *KgGraphRetriever
	global  *KgGraphRetriever
	numDocs int
}

// NewKgHybridGraphRetriever 创建 hybrid 图谱检索器
func NewKgHybridGraphRetriever(local, global *KgGraphRetriever, numDocs int) *KgHybridGraphRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	return &KgHybridGraphRetriever{local: local, global: global, numDocs: numDocs}
}

// WithScoreThreshold 对 local/global 子检索器同时设置向量相似度下限
func (h *KgHybridGraphRetriever) WithScoreThreshold(t float32) *KgHybridGraphRetriever {
	if h == nil {
		return nil
	}
	h.local.WithScoreThreshold(t)
	h.global.WithScoreThreshold(t)
	return h
}

// GetRelevantDocuments 合并实体路与关系路切片，按分数去重截断
func (h *KgHybridGraphRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = h.numDocs
	}
	pool := lightragconst.WideSimilarityFetchK(n, 4)
	g, gctx := errgroup.WithContext(ctx)
	var a, b []schema.Document
	g.Go(func() error {
		var err error
		a, err = h.local.GetRelevantDocuments(gctx, query, pool)
		return err
	})
	g.Go(func() error {
		var err error
		b, err = h.global.GetRelevantDocuments(gctx, query, pool)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return mergeKgDocListsMaxScore(a, b, n), nil
}

// RetrieverType hybrid
func (h *KgHybridGraphRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypeHybrid
}

// KgMixRetriever 向量块检索 + 图谱 hybrid（对齐 LightRAG mix：图 + 向量）
type KgMixRetriever struct {
	vec     *VectorRetriever
	graph   *KgHybridGraphRetriever
	numDocs int
}

// NewKgMixRetriever 创建 mix 检索器
func NewKgMixRetriever(vec *VectorRetriever, graph *KgHybridGraphRetriever, numDocs int) *KgMixRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	return &KgMixRetriever{vec: vec, graph: graph, numDocs: numDocs}
}

// WithScoreThreshold 向量路与图谱路同时应用相似度下限
func (m *KgMixRetriever) WithScoreThreshold(t float32) *KgMixRetriever {
	if m == nil {
		return nil
	}
	m.vec.WithScoreThreshold(t)
	m.graph.WithScoreThreshold(t)
	return m
}

// GetRelevantDocuments 向量路与图谱 hybrid 按 FusionMix 轮询合并
func (m *KgMixRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = m.numDocs
	}
	vf := lightragconst.WideSimilarityFetchK(n, 4)
	gf := lightragconst.WideSimilarityFetchK(n, 4)
	g, gctx := errgroup.WithContext(ctx)
	var vecDocs, graphDocs []schema.Document
	g.Go(func() error {
		var err error
		vecDocs, err = m.vec.GetRelevantDocuments(gctx, query, vf)
		if err != nil {
			return err
		}
		if len(vecDocs) == 0 && strings.TrimSpace(query) != "" && !global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry {
			relax := NewVectorRetriever(m.vec.store, m.vec.namespace, lightragconst.MaxSimilarityFetchK)
			relax.reportType = m.vec.reportType
			relax.scoreThreshold = 0
			// 与纯向量 VectorEmptyRetry 一致：第二次请求用 top 量级 MaxSimilarityFetchK，避免仍按 vf 缩小候选池
			vecDocs, err = relax.GetRelevantDocuments(gctx, query, lightragconst.MaxSimilarityFetchK)
		}
		return err
	})
	g.Go(func() error {
		var err error
		graphDocs, err = m.graph.GetRelevantDocuments(gctx, query, gf)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return MergeFusionDocuments(vecDocs, graphDocs, FusionMix, n), nil
}

// RetrieverType mix
func (m *KgMixRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypeMix
}

// KgGlobalVectorMixRetriever 启用图谱时的 global 模式：普通切片向量 + 关系向量展开（对齐 references/LightRAG/lightrag/operate.py 中 mix 对 chunks_vdb 的 _get_vector_context 与 global 关系检索并存；纯关系路在无 chunk 向量参与时易与「未启用图谱时的 Global」语义召回脱节）
type KgGlobalVectorMixRetriever struct {
	vec     *VectorRetriever
	global  *KgGraphRetriever
	numDocs int
}

// NewKgGlobalVectorMixRetriever 创建 global+向量混合检索器
func NewKgGlobalVectorMixRetriever(vec *VectorRetriever, global *KgGraphRetriever, numDocs int) *KgGlobalVectorMixRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	return &KgGlobalVectorMixRetriever{vec: vec, global: global, numDocs: numDocs}
}

// WithScoreThreshold 向量路与关系路同时应用相似度下限
func (m *KgGlobalVectorMixRetriever) WithScoreThreshold(t float32) *KgGlobalVectorMixRetriever {
	if m == nil {
		return nil
	}
	m.vec.WithScoreThreshold(t)
	m.global.WithScoreThreshold(t)
	return m
}

// GetRelevantDocuments 切片向量与 global 关系展开按 FusionMix 合并（向量侧约 2/3 配额）
func (m *KgGlobalVectorMixRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = m.numDocs
	}
	vf := lightragconst.WideSimilarityFetchK(n, 4)
	gf := lightragconst.WideSimilarityFetchK(n, 4)
	g, gctx := errgroup.WithContext(ctx)
	var vecDocs, relDocs []schema.Document
	g.Go(func() error {
		var err error
		vecDocs, err = m.vec.GetRelevantDocuments(gctx, query, vf)
		if err != nil {
			return err
		}
		if len(vecDocs) == 0 && strings.TrimSpace(query) != "" && !global.LRAG_CONFIG.Rag.VectorSkipEmptyRetry {
			relax := NewVectorRetriever(m.vec.store, m.vec.namespace, lightragconst.MaxSimilarityFetchK)
			relax.reportType = m.vec.reportType
			relax.scoreThreshold = 0
			vecDocs, err = relax.GetRelevantDocuments(gctx, query, lightragconst.MaxSimilarityFetchK)
		}
		return err
	})
	g.Go(func() error {
		var err error
		relDocs, err = m.global.GetRelevantDocuments(gctx, query, gf)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return MergeFusionDocuments(vecDocs, relDocs, FusionMix, n), nil
}

// RetrieverType global
func (m *KgGlobalVectorMixRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypeGlobal
}

func mergeKgDocListsMaxScore(a, b []schema.Document, limit int) []schema.Document {
	best := make(map[string]schema.Document)
	absorb := func(d schema.Document) {
		k := DocumentDedupKey(d)
		old, ok := best[k]
		if !ok || d.Score > old.Score {
			best[k] = d
		}
	}
	for _, d := range a {
		absorb(d)
	}
	for _, d := range b {
		absorb(d)
	}
	out := make([]schema.Document, 0, len(best))
	for _, d := range best {
		out = append(out, d)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].PageContent < out[j].PageContent
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func metadataAsUint(m map[string]any, key string) uint {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
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
	case string:
		u, err := strconv.ParseUint(x, 10, 32)
		if err != nil {
			return 0
		}
		return uint(u)
	default:
		return 0
	}
}

func effectiveKgRetrieverRelatedChunkNumber() int {
	n := global.LRAG_CONFIG.Rag.KgRetrieverRelatedChunkNumber
	if n <= 0 {
		return 0
	}
	if n > 200 {
		return 200
	}
	return n
}
