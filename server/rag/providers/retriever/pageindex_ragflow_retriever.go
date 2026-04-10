package retriever

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/pageindex"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// PageIndexRagflowRetriever 对齐 references/ragflow/rag/nlp/search.py 的 retrieval_by_toc：
// 先向量宽召回，再按文档聚合相似度选出主文档，对该文档的 PageIndex 树做 TOC 相关性打分，对已向量化切片加分（similarity += toc），并补全未命中的高相关切片。
type PageIndexRagflowRetriever struct {
	vector interfaces.Retriever
	pure   *PageIndexRetriever
	// tocLLM 用于 TocRelevanceSearch；对齐 references/ragflow 中 retrieval_by_toc 使用的租户默认 Chat，可与 pure.llm（PageIndex/树推理）分离。
	tocLLM interfaces.LLM
}

// NewPageIndexRagflowRetriever 创建 Ragflow 式 PageIndex 混合检索器；tocLLM 可为 nil，此时回退为 pure.llm。
func NewPageIndexRagflowRetriever(vector interfaces.Retriever, pure *PageIndexRetriever, tocLLM interfaces.LLM) *PageIndexRagflowRetriever {
	return &PageIndexRagflowRetriever{vector: vector, pure: pure, tocLLM: tocLLM}
}

func (r *PageIndexRagflowRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypePageIndex
}

func (r *PageIndexRagflowRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 && r.pure != nil {
		n = r.pure.numDocs
	}
	if n <= 0 {
		n = lightragconst.DefaultChunkTopK
	}
	if r.vector == nil || r.pure == nil {
		if r.pure != nil {
			return r.pure.GetRelevantDocuments(ctx, query, n)
		}
		return nil, nil
	}

	vecDocs, err := r.vector.GetRelevantDocuments(ctx, query, n)
	if err != nil || len(vecDocs) == 0 {
		return r.pure.GetRelevantDocuments(ctx, query, n)
	}

	sums := map[uint]float32{}
	for _, d := range vecDocs {
		did, ok := metadataUintFromMeta(d.Metadata, "document_id")
		if !ok {
			continue
		}
		sums[did] += d.Score
	}
	// 对齐 retrieval_by_toc：按 doc 聚合 similarity 取最大；分数并列时取较小 document_id，避免 Go map 迭代顺序导致非确定主文档
	var bestDoc uint
	var bestSum float32 = -1
	for id, s := range sums {
		if s > bestSum || (s == bestSum && (bestDoc == 0 || id < bestDoc)) {
			bestSum = s
			bestDoc = id
		}
	}
	if bestDoc == 0 {
		return truncateDocsByScore(vecDocs, n), nil
	}

	var dt *DocTree
	for i := range r.pure.docTrees {
		if r.pure.docTrees[i].DocID == bestDoc {
			dt = &r.pure.docTrees[i]
			break
		}
	}
	if dt == nil || len(dt.Tree) == 0 {
		return truncateDocsByScore(vecDocs, n), nil
	}
	llmToc := r.tocLLM
	if llmToc == nil && r.pure != nil {
		llmToc = r.pure.llm
	}
	if llmToc == nil {
		return truncateDocsByScore(vecDocs, n), nil
	}

	entries := pageindex.CapTOCEntries(pageindex.FlattenTreeToTOC(dt.Tree))
	if len(entries) == 0 {
		return truncateDocsByScore(vecDocs, n), nil
	}

	scored, err := pageindex.TocRelevanceSearch(ctx, llmToc, query, entries)
	if err != nil || len(scored) == 0 {
		return truncateDocsByScore(vecDocs, n), nil
	}
	if dt.NodeToRagChunkID == nil || len(dt.NodeToRagChunkID) == 0 {
		return truncateDocsByScore(vecDocs, n), nil
	}

	boostLists := map[uint][]float64{}
	for _, sn := range scored {
		// Ragflow：每个 TOC 项的 ids 可含多枚 chunk；树形索引下将分数施加到该节点子树内全部 rag_chunk（与 task_executor 中按 chunk 区间填充 ids 同义）
		nodeIDs := pageindex.CollectDescendantNodeIDs(dt.Tree, sn.NodeID)
		if len(nodeIDs) == 0 {
			nodeIDs = []string{pageindex.NormalizeNodeID(sn.NodeID)}
		}
		for _, nid := range nodeIDs {
			rid, ok := dt.NodeToRagChunkID[nid]
			if !ok || rid == 0 {
				continue
			}
			boostLists[rid] = append(boostLists[rid], float64(sn.Score))
		}
	}
	idBoost := make(map[uint]float32, len(boostLists))
	for rid, xs := range boostLists {
		var s float64
		for _, x := range xs {
			s += x
		}
		idBoost[rid] = float32(s / float64(len(xs)))
	}
	// references/ragflow relevant_chunks_with_toc：返回的 (chunk_id, score) 对数量上限为 top_n * 2，不随向量宽召回放大
	idBoost = capTocBoostEntries(idBoost, ragflowTocBoostPairCap(n))

	chunkByID := batchRagChunksByIDs(ctx, vecDocs)

	out := make([]schema.Document, 0, len(vecDocs)+len(idBoost))
	seenChunk := make(map[uint]struct{}, len(vecDocs))

	for _, d := range vecDocs {
		did, _ := metadataUintFromMeta(d.Metadata, "document_id")
		rid, hasRid := metadataUintFromMeta(d.Metadata, "rag_chunk_id")
		nd := d
		if did == bestDoc && hasRid {
			if b, ok := idBoost[rid]; ok {
				nd.Score = d.Score + b
			}
			seenChunk[rid] = struct{}{}
		}
		meta := cloneMeta(nd.Metadata)
		if meta == nil {
			meta = map[string]any{}
		}
		if hasRid {
			if ch, ok := chunkByID[rid]; ok {
				mergeChunkMetadataFromDB(ch, meta)
			}
		}
		meta["pageindex_mode"] = "vector_toc_boost"
		nd.Metadata = meta
		out = append(out, nd)
	}

	for rid, boost := range idBoost {
		if _, ok := seenChunk[rid]; ok {
			continue
		}
		var ch rag.RagChunk
		if err := global.LRAG_DB.WithContext(ctx).First(&ch, rid).Error; err != nil {
			continue
		}
		if ch.DocumentID != bestDoc {
			continue
		}
		suppMeta := map[string]any{
			"document_id":    bestDoc,
			"doc_name":       dt.DocName,
			"rag_chunk_id":   ch.ID,
			"chunk_index":    ch.ChunkIndex,
			"pageindex_mode": "toc_supplement",
		}
		mergeChunkMetadataFromDB(&ch, suppMeta)
		out = append(out, schema.Document{
			PageContent: ch.Content,
			Score:       boost,
			Metadata:    suppMeta,
		})
	}

	out = MergeRetrievalByChildren(ctx, out)

	sort.SliceStable(out, func(i, j int) bool {
		si, sj := out[i].Score, out[j].Score
		if si != sj {
			return si > sj
		}
		li, lj := len(out[i].PageContent), len(out[j].PageContent)
		if li != lj {
			return li > lj
		}
		return out[i].PageContent < out[j].PageContent
	})
	if len(out) > n {
		out = out[:n]
	}
	return out, nil
}

// ragflowTocBoostPairCap 对齐 references/ragflow：relevant_chunks_with_toc(..., topn*2)；
// 外层 n 可能为宽池条数时，用 DefaultChunkTopK 封顶，避免把「最终上下文条数」误放大到宽召回量级。
func ragflowTocBoostPairCap(requestedN int) int {
	top := requestedN
	if top > lightragconst.DefaultChunkTopK {
		top = lightragconst.DefaultChunkTopK
	}
	pairCap := 2 * top
	if pairCap < 12 {
		pairCap = 12
	}
	if pairCap > 256 {
		pairCap = 256
	}
	return pairCap
}

func capTocBoostEntries(m map[uint]float32, max int) map[uint]float32 {
	if max <= 0 || len(m) <= max {
		return m
	}
	type kv struct {
		id uint
		s  float32
	}
	list := make([]kv, 0, len(m))
	for id, s := range m {
		list = append(list, kv{id, s})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].s != list[j].s {
			return list[i].s > list[j].s
		}
		return list[i].id < list[j].id
	})
	out := make(map[uint]float32, max)
	for i := 0; i < max && i < len(list); i++ {
		out[list[i].id] = list[i].s
	}
	return out
}

func truncateDocsByScore(docs []schema.Document, n int) []schema.Document {
	if n <= 0 || len(docs) <= n {
		return docs
	}
	cp := make([]schema.Document, len(docs))
	copy(cp, docs)
	sort.SliceStable(cp, func(i, j int) bool {
		if cp[i].Score != cp[j].Score {
			return cp[i].Score > cp[j].Score
		}
		return len(cp[i].PageContent) > len(cp[j].PageContent)
	})
	return cp[:n]
}

func cloneMeta(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func metadataUintFromMeta(meta map[string]any, key string) (uint, bool) {
	if meta == nil {
		return 0, false
	}
	raw, ok := meta[key]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return uint(v), true
	case float32:
		return uint(v), true
	case int:
		return uint(v), true
	case int64:
		return uint(v), true
	case uint:
		return v, true
	case uint64:
		return uint(v), true
	case json.Number:
		u64, err := v.Int64()
		if err != nil || u64 < 0 {
			return 0, false
		}
		return uint(u64), true
	case string:
		u, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(u), true
	default:
		return 0, false
	}
}
