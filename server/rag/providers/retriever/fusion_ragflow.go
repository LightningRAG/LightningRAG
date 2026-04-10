package retriever

import (
	"sort"

	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// minMaxNormalizeScores 将每条文档的 Score 线性归一化到 [0,1]（单条或同分时全 1），便于与 Ragflow FusionExpr weighted_sum 尺度可比。
func minMaxNormalizeScores(docs []schema.Document) []float32 {
	if len(docs) == 0 {
		return nil
	}
	out := make([]float32, len(docs))
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
		for i := range out {
			out[i] = 1
		}
		return out
	}
	span := maxS - minS
	for i := range docs {
		out[i] = (docs[i].Score - minS) / span
	}
	return out
}

// mergeRagflowWeightedFusion 对齐 references/ragflow/rag/nlp/search.py：全文与向量两路分数在 [0,1] 归一化后按 termW/vecW 加权求和，再截断到 finalN。
// minFused>0 时在归一化融合分上过滤弱命中（近似 Ragflow retrieval 的 similarity_threshold，尺度为融合分而非原始余弦）。
func mergeRagflowWeightedFusion(vecDocs, kwDocs []schema.Document, finalN int, termW, vecW, minFused float32) []schema.Document {
	vecN := minMaxNormalizeScores(vecDocs)
	kwN := minMaxNormalizeScores(kwDocs)

	type ent struct {
		vecN, kwN  float32
		hasV, hasK bool
		doc        schema.Document
	}
	m := make(map[string]*ent, len(vecDocs)+len(kwDocs))
	for i, d := range vecDocs {
		k := DocumentDedupKey(d)
		e := m[k]
		if e == nil {
			e = &ent{}
			m[k] = e
		}
		v := vecN[i]
		if !e.hasV || v > e.vecN {
			e.vecN, e.hasV = v, true
			e.doc = d
		}
	}
	for i, d := range kwDocs {
		k := DocumentDedupKey(d)
		e := m[k]
		if e == nil {
			e = &ent{}
			m[k] = e
		}
		kk := kwN[i]
		if !e.hasK || kk > e.kwN {
			e.kwN, e.hasK = kk, true
		}
		if !e.hasV {
			e.doc = d
		}
	}

	out := make([]schema.Document, 0, len(m))
	for _, e := range m {
		var vn, kn float32
		if e.hasV {
			vn = e.vecN
		}
		if e.hasK {
			kn = e.kwN
		}
		fused := vecW*vn + termW*kn
		if minFused > 0 && fused < minFused {
			continue
		}
		e.doc.Score = fused
		out = append(out, e.doc)
	}
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
	if finalN > 0 && len(out) > finalN {
		out = out[:finalN]
	}
	return out
}
