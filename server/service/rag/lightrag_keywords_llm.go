package rag

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"go.uber.org/zap"
)

type llmKeywordJSON struct {
	HighLevel []string `json:"high_level_keywords"`
	LowLevel  []string `json:"low_level_keywords"`
}

var reJSONFenceKeywords = reJSONFence // 与 kg_extract 共用 ```json 围栏解析

type keywordExtractFlightResult struct {
	HL []string
	LL []string
}

// extractKeywordsWithLLM 对齐 LightRAG extract_keywords_only：从用户问句抽取高层/低层关键词；extraContext 可为多轮摘要供消歧
// 可选 TTL 缓存 + singleflight，减轻重复问句的并发与费用压力。
func extractKeywordsWithLLM(ctx context.Context, llm interfaces.LLM, userQuery, extraContext string) (hl, ll []string, err error) {
	q := strings.TrimSpace(userQuery)
	if llm == nil || q == "" {
		return nil, nil, nil
	}
	cacheKey := keywordExtractCacheKey(q, extraContext)
	ttlSec := global.LRAG_CONFIG.Rag.KeywordExtractCacheTTLSeconds
	if ttlSec > 0 {
		if h, l, ok := loadKeywordExtractCache(cacheKey); ok {
			return h, l, nil
		}
	}
	v, err, _ := global.LRAG_Concurrency_Control.Do("lrag_kw:"+cacheKey, func() (interface{}, error) {
		if ttlSec > 0 {
			if h, l, ok := loadKeywordExtractCache(cacheKey); ok {
				return &keywordExtractFlightResult{HL: h, LL: l}, nil
			}
		}
		h, l, e := doExtractKeywordsWithLLM(ctx, llm, q, extraContext)
		if e != nil {
			return nil, e
		}
		if ttlSec > 0 {
			saveKeywordExtractCache(cacheKey, h, l, ttlSec)
		}
		return &keywordExtractFlightResult{HL: h, LL: l}, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if v == nil {
		return nil, nil, nil
	}
	fr, _ := v.(*keywordExtractFlightResult)
	if fr == nil {
		return nil, nil, nil
	}
	return cloneKeywordSlice(fr.HL), cloneKeywordSlice(fr.LL), nil
}

func doExtractKeywordsWithLLM(ctx context.Context, llm interfaces.LLM, userQuery, extraContext string) (hl, ll []string, err error) {
	q := strings.TrimSpace(userQuery)
	system := `你是检索关键词抽取器。根据用户问题输出 JSON，不要 Markdown，不要解释。
字段：high_level_keywords（抽象主题、领域概念，字符串数组），low_level_keywords（具体实体、名称、术语，字符串数组）。
每侧 0～12 个短语，与问题语言一致；专有名词保持原文（对齐 LightRAG keywords_extraction）。
当问题有明确主题或实体时，high_level_keywords 与 low_level_keywords 均应至少各含 1 个短语；仅对无实质内容的寒暄（如「你好」）才允许两侧均为 []。
若提供对话摘录，仅用于理解指代与主题，关键词仍应主要服务于「主问题」。`
	user := q
	if xc := strings.TrimSpace(extraContext); xc != "" {
		user = "对话摘录（供理解上下文）：\n" + xc + "\n\n主问题：\n" + q
	}
	msgs := []interfaces.MessageContent{
		{Role: interfaces.MessageRoleSystem, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: system}}},
		{Role: interfaces.MessageRoleHuman, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: user}}},
	}
	resp, err := llm.GenerateContent(ctx, msgs, interfaces.WithTemperature(0.1), interfaces.WithMaxTokens(512))
	if err != nil {
		return nil, nil, err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return nil, nil, nil
	}
	raw := strings.TrimSpace(resp.Choices[0].Content)
	if m := reJSONFenceKeywords.FindStringSubmatch(raw); len(m) > 1 {
		raw = strings.TrimSpace(m[1])
	}
	var data llmKeywordJSON
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, nil, err
	}
	return trimKeywordSlice(data.HighLevel), trimKeywordSlice(data.LowLevel), nil
}

func trimKeywordSlice(in []string) []string {
	var out []string
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func cloneKeywordSlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}

// KeywordsForAPIResponse 将解析后的关键词转为稳定 JSON 数组（空则为 [] 而非 null）
func KeywordsForAPIResponse(s []string) []string {
	if len(s) == 0 {
		return []string{}
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}

// PrepareLightningRAGSearchQueries 生成合并检索串与图谱分支专用串（低层→实体向量，高层→关系向量；对齐 LightRAG ll_keywords_str / hl_keywords_str 分工）
// 返回值 resolvedHl、resolvedLl 为最终用于检索的高层/低层关键词（含自动抽取结果），与上游 raw_data 中关键词字段对应。
// keywordExtraContext 在开启自动抽词时并入 LLM 提示（如 conversation_history 摘要）
func PrepareLightningRAGSearchQueries(ctx context.Context, uid uint, llm interfaces.LLM, displayQ string, reqHl, reqLl []string, keywordExtraContext string) (combined, kgEntity, kgRel string, resolvedHl, resolvedLl []string) {
	q := strings.TrimSpace(displayQ)
	hl := trimKeywordSlice(reqHl)
	ll := trimKeywordSlice(reqLl)

	if global.LRAG_CONFIG.Rag.AutoExtractQueryKeywords && len(hl) == 0 && len(ll) == 0 && llm != nil && q != "" {
		h, l, err := extractKeywordsWithLLM(ctx, llm, q, keywordExtraContext)
		if err != nil {
			global.LRAG_LOG.Warn("查询关键词 LLM 抽取失败，使用原始问句检索", zap.Error(err))
		} else {
			hl, ll = h, l
		}
	}

	if len(hl) == 0 && len(ll) == 0 {
		return q, q, q, nil, nil
	}

	combined = AugmentQueryWithLightningRAGKeywords(q, hl, ll)
	// 图谱向量检索对齐 references/LightRAG/lightrag/operate.py _perform_kg_search：
	// _get_node_data 仅对 entities_vdb 使用 ll_keywords_str（", ".join(ll_keywords)），
	// _get_edge_data 仅对 relationships_vdb 使用 hl_keywords_str。
	// 若把完整问句与关键词拼接后嵌入，整句语义易与「实体名 / 关系描述」向量分布错位，global 模式尤其会拉回不相干关系及其切片。
	if len(ll) > 0 {
		kgEntity = strings.Join(ll, ", ")
	} else {
		kgEntity = q
	}
	if len(hl) > 0 {
		kgRel = strings.Join(hl, ", ")
	} else {
		kgRel = q
	}
	return combined, kgEntity, kgRel, cloneKeywordSlice(hl), cloneKeywordSlice(ll)
}
