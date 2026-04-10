package retriever

import (
	"context"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"golang.org/x/sync/errgroup"
)

// MergeFusionDocuments 将两路已检索结果按 FusionHybrid / FusionMix 策略轮询合并去重（供 KgMix 等与 FusionRetriever 一致的组合逻辑复用）
func MergeFusionDocuments(vecDocs, secondDocs []schema.Document, strategy FusionStrategy, n int) []schema.Document {
	if n <= 0 {
		return nil
	}
	// key -> 在 out 中的下标；同一切片从两路出现时保留分数更高的版本（原先先到先得会丢掉后路的更高分）
	keyIndex := make(map[string]int)
	out := make([]schema.Document, 0, n)
	tryAppend := func(d schema.Document) bool {
		k := DocumentDedupKey(d)
		if idx, ok := keyIndex[k]; ok {
			if d.Score > out[idx].Score {
				out[idx] = d
			}
			return true
		}
		if len(out) >= n {
			return false
		}
		keyIndex[k] = len(out)
		out = append(out, d)
		return true
	}
	i, j := 0, 0
	vecBurst := 1
	if strategy == FusionMix {
		vecBurst = 2
	}
	pendingVec := vecBurst
	for len(out) < n && (i < len(vecDocs) || j < len(secondDocs)) {
		if pendingVec > 0 && i < len(vecDocs) {
			d := vecDocs[i]
			i++
			if tryAppend(d) {
				pendingVec--
			}
			continue
		}
		if j < len(secondDocs) {
			d := secondDocs[j]
			j++
			_ = tryAppend(d)
			pendingVec = vecBurst
			continue
		}
		for len(out) < n && i < len(vecDocs) {
			d := vecDocs[i]
			i++
			if !tryAppend(d) {
				continue
			}
		}
		break
	}
	for len(out) < n && j < len(secondDocs) {
		d := secondDocs[j]
		j++
		_ = tryAppend(d)
	}
	return out
}

// FusionStrategy 向量与关键词两路的配额策略（对应 LightningRAG hybrid / mix 的近似实现）
type FusionStrategy int

const (
	// FusionHybrid 约各占一半候选，再合并去重
	FusionHybrid FusionStrategy = iota
	// FusionMix 更侧重向量（约 2/3 向量、1/3 关键词）
	FusionMix
)

// fusionRagflowOpts 对齐 references/ragflow Dealer.search 中 FusionExpr("weighted_sum", topk, {"weights": "0.05,0.95"})
type fusionRagflowOpts struct {
	termWeight float32
	vecWeight  float32
	minFused   float32
	emptyRetry bool
}

func defaultFusionRagflowOpts() fusionRagflowOpts {
	return fusionRagflowOpts{
		termWeight: 0.05,
		vecWeight:  0.95,
		minFused:   0,
		emptyRetry: true,
	}
}

// FusionOption 可选调整 hybrid/mix 融合行为
type FusionOption func(*fusionRagflowOpts)

// WithFusionWeights 全文/向量权重（会按和为 1 归一化；与 Ragflow weights 字符串同语义）
func WithFusionWeights(termW, vecW float32) FusionOption {
	return func(o *fusionRagflowOpts) {
		sum := termW + vecW
		if sum > 0 {
			o.termWeight = termW / sum
			o.vecWeight = vecW / sum
		}
	}
}

// WithFusionMinScore 融合分下限，<=0 表示不过滤（归一化后融合分 ∈ [0,1]）
func WithFusionMinScore(min float32) FusionOption {
	return func(o *fusionRagflowOpts) {
		o.minFused = min
	}
}

// WithFusionEmptyRetry 零命中时是否用纯向量宽召回再试（Ragflow search 第二次放宽）
func WithFusionEmptyRetry(on bool) FusionOption {
	return func(o *fusionRagflowOpts) {
		o.emptyRetry = on
	}
}

// FusionRetriever 合并向量检索与关键词检索结果
type FusionRetriever struct {
	vec      *VectorRetriever
	kw       *KeywordRetriever
	strategy FusionStrategy
	numDocs  int
	retType  interfaces.RetrieverType
	ragflow  fusionRagflowOpts
}

// NewFusionRetriever 创建融合检索器；retType 为 hybrid 或 mix；opts 可覆盖 Ragflow 式加权融合参数
func NewFusionRetriever(vec *VectorRetriever, kw *KeywordRetriever, strategy FusionStrategy, numDocs int, retType interfaces.RetrieverType, opts ...FusionOption) *FusionRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	rf := defaultFusionRagflowOpts()
	for _, o := range opts {
		o(&rf)
	}
	return &FusionRetriever{
		vec:      vec,
		kw:       kw,
		strategy: strategy,
		numDocs:  numDocs,
		retType:  retType,
		ragflow:  rf,
	}
}

// WithVectorScoreThreshold 仅提高向量路的相似度下限；关键词路不受影响
func (f *FusionRetriever) WithVectorScoreThreshold(t float32) *FusionRetriever {
	if f.vec != nil && t > 0 {
		f.vec.WithScoreThreshold(t)
	}
	return f
}

// GetRelevantDocuments 并行拉取向量/全文候选，按 Ragflow 式加权融合分排序并截断至 numDocs。
// 零命中且开启 emptyRetry：先纯向量无阈值宽池；仍空则宽松全文（RelaxedKeywordSearch），对齐 Ragflow search 第二次放宽关键词侧。
func (f *FusionRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = f.numDocs
	}
	vecFetch := lightragconst.WideSimilarityFetchK(n, 4)
	kwFetch := vecFetch
	g, gctx := errgroup.WithContext(ctx)
	var vecDocs, kwDocs []schema.Document
	g.Go(func() error {
		var err error
		vecDocs, err = f.vec.GetRelevantDocuments(gctx, query, vecFetch)
		return err
	})
	g.Go(func() error {
		var err error
		kwDocs, err = f.kw.GetRelevantDocuments(gctx, query, kwFetch)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	out := mergeRagflowWeightedFusion(vecDocs, kwDocs, n, f.ragflow.termWeight, f.ragflow.vecWeight, f.ragflow.minFused)
	qtrim := strings.TrimSpace(query)
	if len(out) == 0 && f.ragflow.emptyRetry && qtrim != "" && f.vec != nil {
		relax := NewVectorRetriever(f.vec.store, f.vec.namespace, lightragconst.MaxSimilarityFetchK)
		relax.reportType = f.vec.reportType
		relax.scoreThreshold = 0
		var rerr error
		out, rerr = relax.GetRelevantDocuments(ctx, query, n)
		if rerr != nil {
			return nil, rerr
		}
	}
	if len(out) == 0 && f.ragflow.emptyRetry && qtrim != "" && f.kw != nil {
		kopts := []interfaces.VectorStoreOption{
			func(o *interfaces.VectorStoreOptions) { o.Namespace = f.kw.namespace },
			interfaces.WithRelaxedKeywordSearch(true),
		}
		kdocs, kerr := f.kw.store.KeywordSearch(ctx, qtrim, kwFetch, kopts...)
		if kerr != nil {
			return nil, kerr
		}
		out = kdocs
		if len(out) > n {
			out = out[:n]
		}
	}
	return out, nil
}

// RetrieverType 返回 hybrid 或 mix
func (f *FusionRetriever) RetrieverType() interfaces.RetrieverType {
	return f.retType
}
