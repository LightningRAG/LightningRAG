package rag

import (
	"strings"

	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/llm"
)

// maxConversationHistoryItems 单次请求附加 history 条数上限，防止恶意超大 payload
const maxConversationHistoryItems = 40

// conversationHistoryItemsToMessages 将 LightningRAG 风格 conversation_history 转为 LLM 消息（仅参与多轮上下文，不参与检索）
func conversationHistoryItemsToMessages(items []request.ConversationHistoryItem) []interfaces.MessageContent {
	if len(items) == 0 {
		return nil
	}
	n := len(items)
	if n > maxConversationHistoryItems {
		n = maxConversationHistoryItems
	}
	out := make([]interfaces.MessageContent, 0, n)
	for i := 0; i < n; i++ {
		it := items[i]
		role := strings.TrimSpace(strings.ToLower(it.Role))
		content := strings.TrimSpace(it.Content)
		if role == "" || content == "" {
			continue
		}
		var mr interfaces.MessageRole
		switch role {
		case "assistant":
			mr = interfaces.MessageRoleAssistant
		case "system":
			mr = interfaces.MessageRoleSystem
		case "user", "human":
			mr = interfaces.MessageRoleHuman
		default:
			mr = interfaces.MessageRoleHuman
		}
		if mr == interfaces.MessageRoleAssistant {
			content = llm.StripAssistantReasoningMarkers(content)
		}
		out = append(out, interfaces.MessageContent{
			Role:  mr,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: content}},
		})
	}
	return out
}

// effectiveTruncationBudget 合并模型 MaxContextTokens 与单次请求的 maxTotalTokens（对齐 LightningRAG QueryParam.max_total_tokens）
func effectiveTruncationBudget(modelMax uint, requestMax *uint) uint {
	if requestMax == nil || *requestMax == 0 {
		return modelMax
	}
	if modelMax == 0 {
		return *requestMax
	}
	if *requestMax < modelMax {
		return *requestMax
	}
	return modelMax
}
