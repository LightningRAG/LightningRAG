package llm

import "strings"

// ollamaThinkFromReasoningEffort 映射到 Ollama /api/chat 的 think 字段（bool 或 low/medium/high）。
func ollamaThinkFromReasoningEffort(effort string) (val any, set bool) {
	e := strings.ToLower(strings.TrimSpace(effort))
	if e == "" {
		return nil, false
	}
	switch e {
	case "disabled", "none", "off":
		return false, true
	case "low":
		return "low", true
	case "medium":
		return "medium", true
	case "high":
		return "high", true
	default:
		return true, true
	}
}
