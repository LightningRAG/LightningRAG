package rerank

// scoreRow 单条 rerank 结果与输入文档下标的对应关系（index 或 document_index）。
type scoreRow struct {
	Index  int
	DocIdx int
	Score  float64
}

// fillScoresForDocuments 将带下标的分数写回与输入文档等长的切片。
// 部分兼容网关省略 index（JSON 反序列化后均为 0），会把所有分数学到第 0 篇；
// 当且仅当「所有有效下标都落在同一篇、且结果条数与文档数一致」时，按响应数组顺序与文档顺序一一对应。
func fillScoresForDocuments(textsLen int, rows []scoreRow) []float32 {
	if textsLen <= 0 {
		return nil
	}
	scores := make([]float32, textsLen)
	uniq := make(map[int]struct{})
	for _, r := range rows {
		idx := -1
		if r.Index >= 0 && r.Index < textsLen {
			idx = r.Index
		} else if r.DocIdx >= 0 && r.DocIdx < textsLen {
			idx = r.DocIdx
		}
		if idx >= 0 {
			scores[idx] = float32(r.Score)
			uniq[idx] = struct{}{}
		}
	}
	if len(uniq) == 1 && textsLen > 1 && len(rows) == textsLen {
		var only int
		for k := range uniq {
			only = k
		}
		if only == 0 {
			for i, r := range rows {
				scores[i] = float32(r.Score)
			}
		}
	}
	return scores
}
