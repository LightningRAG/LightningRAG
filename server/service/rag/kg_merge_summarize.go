package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	llmprov "github.com/LightningRAG/LightningRAG/server/rag/providers/llm"
	"go.uber.org/zap"
)

// 对齐 references/LightRAG/lightrag/prompt.py PROMPTS["summarize_entity_descriptions"]（实体与关系共用），输出为简体中文摘要
const kgMergeSummarizeSystemPrompt = `你是知识图谱编辑。将「描述列表」中多条关于同一对象（实体或实体间关系）的文字合并为一段连贯、客观的第三人称说明。
要求：整合各条中的关键信息，不凭空添加事实；专有名词保持原文；若存在明显矛盾，可并列说明或标注不确定性。
只输出摘要正文，不要标题、Markdown、引号包裹或任何前后缀。`

// 单轮 LLM 最多处理的描述段数；超出则 map-reduce（对齐 references/LightRAG/lightrag/operate.py 多轮 _summarize_descriptions）
const (
	kgMergeSummarizeMaxSegmentsPerLLMCall = 8
	kgMergeSummarizeMapBatchSize          = 6
	kgMergeSummarizeMapReduceMaxRounds    = 16
)

func effectiveKgMergeSummarizeMinSegments() int {
	n := global.LRAG_CONFIG.Rag.KgMergeSummarizeMinSegments
	if n <= 0 {
		return 0
	}
	return n
}

func effectiveKgMergeSummaryTargetRunes() int {
	n := global.LRAG_CONFIG.Rag.KgMergeSummaryTargetRunes
	if n <= 0 {
		return 1200
	}
	if n > 8000 {
		return 8000
	}
	return n
}

// kgDescriptionSegmentsForSummarize 将合并后的 description 按换行拆成非空段（与 mergeKgDescription 的拼接方式一致）
func kgDescriptionSegmentsForSummarize(merged string) []string {
	merged = strings.TrimSpace(merged)
	if merged == "" {
		return nil
	}
	var out []string
	for _, line := range strings.Split(merged, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func kgMergeSummarizeUserPayload(objectKind, displayName string, segments []string) (string, error) {
	var lines []string
	for _, s := range segments {
		b, err := json.Marshal(map[string]string{"Description": s})
		if err != nil {
			return "", err
		}
		lines = append(lines, string(b))
	}
	target := effectiveKgMergeSummaryTargetRunes()
	return fmt.Sprintf(`---任务---
将下列 JSONL（每行一个 JSON 对象，字段 "Description"）中的描述合并为一段摘要。

---对象---
类型：%s
名称：%s

---描述列表---
%s

---要求---
摘要总长度建议不超过约 %d 个 Unicode 字符；语言以中文为主，专有名词可保留原文。`,
		objectKind,
		displayName,
		strings.Join(lines, "\n"),
		target,
	), nil
}

// kgSummarizeBatchWithCache 对一段描述列表做单次摘要（≥2 段）；开启 TTL 时按 batch 维度缓存
func kgSummarizeBatchWithCache(ctx context.Context, llm interfaces.LLM, objectKind, displayName string, batch []string) (string, error) {
	if len(batch) < 2 {
		return "", fmt.Errorf("batch too short")
	}
	user, err := kgMergeSummarizeUserPayload(objectKind, displayName, batch)
	if err != nil {
		return "", err
	}
	ttlSec := global.LRAG_CONFIG.Rag.KgMergeSummarizeLLMCacheTTLSeconds
	if ttlSec > 0 {
		key := kgMergeSummarizeCacheKey(objectKind, displayName, batch)
		if raw, ok := loadKgMergeSummarizeCache(key); ok {
			return raw, nil
		}
		out, err := kgSummarizeKGDescriptionSegmentsUncached(ctx, llm, user)
		if err == nil && out != "" {
			saveKgMergeSummarizeCache(key, out, ttlSec)
		}
		return out, err
	}
	return kgSummarizeKGDescriptionSegmentsUncached(ctx, llm, user)
}

func kgSummarizeKGDescriptionSegmentsUncached(ctx context.Context, llm interfaces.LLM, user string) (string, error) {
	msgs := []interfaces.MessageContent{
		{Role: interfaces.MessageRoleSystem, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: kgMergeSummarizeSystemPrompt}}},
		{Role: interfaces.MessageRoleHuman, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: user}}},
	}
	resp, err := llm.GenerateContent(ctx, msgs, interfaces.WithTemperature(0.2), interfaces.WithMaxTokens(2048))
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty llm response")
	}
	out := strings.TrimSpace(llmprov.StripAssistantReasoningMarkers(resp.Choices[0].Content))
	if out == "" {
		return "", fmt.Errorf("empty summary")
	}
	return out, nil
}

func kgSummarizeKGDescriptionSegmentsMapReduce(ctx context.Context, llm interfaces.LLM, objectKind, displayName string, segments []string) (string, error) {
	cur := segments
	rounds := 0
	for len(cur) > kgMergeSummarizeMaxSegmentsPerLLMCall {
		rounds++
		if rounds > kgMergeSummarizeMapReduceMaxRounds {
			return "", fmt.Errorf("kg summarize map-reduce exceeded max rounds")
		}
		var next []string
		for i := 0; i < len(cur); i += kgMergeSummarizeMapBatchSize {
			end := i + kgMergeSummarizeMapBatchSize
			if end > len(cur) {
				end = len(cur)
			}
			batch := cur[i:end]
			switch len(batch) {
			case 0:
				continue
			case 1:
				next = append(next, batch[0])
			default:
				s, err := kgSummarizeBatchWithCache(ctx, llm, objectKind, displayName, batch)
				if err != nil {
					return "", err
				}
				next = append(next, s)
			}
		}
		cur = next
		if len(cur) < 2 {
			if len(cur) == 1 {
				return cur[0], nil
			}
			return "", fmt.Errorf("kg summarize map-reduce produced empty list")
		}
	}
	return kgSummarizeBatchWithCache(ctx, llm, objectKind, displayName, cur)
}

func kgSummarizeKGDescriptionSegments(ctx context.Context, llm interfaces.LLM, objectKind, displayName string, segments []string) (string, error) {
	if llm == nil || len(segments) < 2 {
		return "", fmt.Errorf("invalid summarize input")
	}
	ttlSec := global.LRAG_CONFIG.Rag.KgMergeSummarizeLLMCacheTTLSeconds
	if ttlSec > 0 {
		fullKey := kgMergeSummarizeCacheKey(objectKind, displayName, segments)
		if raw, ok := loadKgMergeSummarizeCache(fullKey); ok {
			return raw, nil
		}
		out, err := kgSummarizeKGDescriptionSegmentsMapReduce(ctx, llm, objectKind, displayName, segments)
		if err == nil && out != "" {
			saveKgMergeSummarizeCache(fullKey, out, ttlSec)
		}
		return out, err
	}
	return kgSummarizeKGDescriptionSegmentsMapReduce(ctx, llm, objectKind, displayName, segments)
}

func kgMaybeSummarizeMergedDescription(ctx context.Context, llm interfaces.LLM, objectKind, displayName, merged string, warnField zap.Field) string {
	minSeg := effectiveKgMergeSummarizeMinSegments()
	if minSeg <= 0 || llm == nil {
		return merged
	}
	segs := kgDescriptionSegmentsForSummarize(merged)
	if len(segs) < minSeg {
		return merged
	}
	summary, err := kgSummarizeKGDescriptionSegments(ctx, llm, objectKind, displayName, segs)
	if err != nil {
		global.LRAG_LOG.Warn("知识图谱描述 LLM 摘要失败，保留合并原文",
			warnField,
			zap.Error(err))
		return merged
	}
	return kgClampStoredDescription(summary)
}

func kgMaybeSummarizeEntityDescription(ctx context.Context, llm interfaces.LLM, displayName, merged string) string {
	return kgMaybeSummarizeMergedDescription(ctx, llm, "实体", displayName, merged, zap.String("entity", displayName))
}

func kgMaybeSummarizeRelationshipDescription(ctx context.Context, llm interfaces.LLM, srcName, tgtName, merged string) string {
	label := strings.TrimSpace(srcName) + " — " + strings.TrimSpace(tgtName)
	label = strings.TrimSpace(label)
	if label == "—" {
		label = "关系"
	}
	return kgMaybeSummarizeMergedDescription(ctx, llm, "关系", label, merged, zap.String("relationship", label))
}
