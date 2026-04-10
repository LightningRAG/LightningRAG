package llm

import "strings"

const (
	anthropicThinkingMinBudget     = 1024  // API 下限
	anthropicThinkingMinAnswerRoom = 2048  // 为正文预留，避免 max_tokens 过小导致仅输出思考
	anthropicThinkingMaxCap        = 200000 // 低于各模型实际上限即可
)

// anthropicThinkingBlockFromReasoningEffort 生成 Messages / Bedrock 共用的 thinking 对象（type=enabled + budget_tokens）。
// 若 effort 为空或为关闭语义则返回 nil；必要时上调 *maxTokens 以满足 budget + 正文空间。
func anthropicThinkingBlockFromReasoningEffort(effort string, maxTokens *int) map[string]any {
	e := strings.ToLower(strings.TrimSpace(effort))
	if e == "" || e == "disabled" || e == "none" || e == "off" {
		return nil
	}
	var budget int
	switch e {
	case "low":
		budget = 2048
	case "medium":
		budget = 8192
	case "high":
		budget = 20_000
	default:
		budget = 4096
	}
	if budget < anthropicThinkingMinBudget {
		budget = anthropicThinkingMinBudget
	}
	need := budget + anthropicThinkingMinAnswerRoom
	if *maxTokens < need {
		*maxTokens = need
	}
	if *maxTokens > anthropicThinkingMaxCap {
		*maxTokens = anthropicThinkingMaxCap
	}
	if budget+anthropicThinkingMinAnswerRoom > *maxTokens {
		budget = *maxTokens - anthropicThinkingMinAnswerRoom
		if budget < anthropicThinkingMinBudget {
			budget = anthropicThinkingMinBudget
		}
	}
	return map[string]any{
		"type":          "enabled",
		"budget_tokens": budget,
	}
}
