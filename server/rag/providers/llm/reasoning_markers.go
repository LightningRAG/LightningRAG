package llm

import "strings"

// AssistantReasoningOpenTag / AssistantReasoningCloseTag 与前端 conversation 解析约定一致，
// 用于包装 API 返回的 reasoning_content / thinking 等字段。
const (
	AssistantReasoningOpenTag  = "<" + "think" + ">"
	AssistantReasoningCloseTag = "<" + "/" + "think" + ">"
	// AssistantOutputTruncationNotice 检测到输出因长度上限被截断时追加的提示（OpenAI/Azure/Anthropic/Ollama 等共用文案）。
	AssistantOutputTruncationNotice = "\n······\nThe reply was truncated because it hit the model context length limit."
)

// StripAssistantReasoningMarkers 移除成对的 think 标签及其中内容（与前端 stripAllThinkingBlocks 一致）。
// 用于多轮对话把历史 assistant 回填给模型时省 token、减少模型误读内部推理。
func StripAssistantReasoningMarkers(s string) string {
	if s == "" {
		return s
	}
	out := s
	open, close := AssistantReasoningOpenTag, AssistantReasoningCloseTag
	for {
		i0 := strings.Index(out, open)
		if i0 == -1 {
			break
		}
		afterOpen := i0 + len(open)
		i1 := strings.Index(out[afterOpen:], close)
		if i1 == -1 {
			break
		}
		i1 += afterOpen
		out = out[:i0] + out[i1+len(close):]
	}
	return out
}
