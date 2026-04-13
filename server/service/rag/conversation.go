package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/llm"
	ragretriever "github.com/LightningRAG/LightningRAG/server/rag/providers/retriever"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	ragschema "github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/LightningRAG/LightningRAG/server/rag/tools"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ragCitationPrompt 对齐 references/LightRAG prompt.py 中 rag_response / naive_rag：仅用本消息 Context、Markdown、文末 ### References；[ID:n] 角标约定与前端一致。
const ragCitationPrompt = `Answer the user's question using only the **Context** in this message: if <knowledge_graph> is present, you may use it together with <context> chunks. Do not invent, assume, or infer facts that are not explicitly supported by that material. You may use general knowledge only to form fluent sentences, not to add new factual claims.

Citation rules (must follow; LightningRAG chunk indices):
- Each item in <context> is labeled "ID: 0", "ID: 1", … (zero-based). In the body of your answer use the same numbers in brackets, e.g. [ID:0], [ID:1]. Do not start numbering from 1.
- You may only cite IDs that appear in <context> in **this** user message; do not point citations at other spans.
- **Multi-turn chat**: [ID:n] from earlier turns applied only to that turn's context; do not mix them with this turn's <context>.
- When paraphrasing a span, add the matching [ID:n] at the end of the sentence (before final punctuation). Multiple citations in one sentence are allowed, e.g. [ID:0][ID:1]. Do not invent IDs; common knowledge needs no citation.

Formatting & language:
- Use Markdown (headings, lists, bold) when it improves clarity.
- Reply in the same language as the user's question unless system instructions specify otherwise.

References section (when you cited any chunk with [ID:n]):
- After the main answer, add a section with heading exactly ### References
- List at most 5 distinct cited chunks, one line each, e.g. - [ID:0] Short document title (from Source summary or document name in that chunk block)
- Do not add footnotes, commentary, or any text after the References section.

`

// ragNoMatchedChunksHint 在「已关联知识库且已执行检索但无切片」时注入用户侧；对齐 references/LightRAG prompt.py fail_response 中 [no-context] 机器可解析约定，减少模型编造引用
const ragNoMatchedChunksHint = `<rag_retrieval_status>
未从已关联知识库中检索到相关文档切片（结果为空或检索失败）。请基于常识回答，或明确说明当前知识库中没有匹配材料；不要编造 [ID:n] 引用或虚构文档内容。标签：[no-context]
</rag_retrieval_status>

User message:
`

// ragRefSourceLabel aligns with LightningRAG frontend Fig.n: index is 0-based; display uses (index+1) as human-friendly ordinal.
func ragRefSourceLabel(meta map[string]any, index int) string {
	var doc, title string
	if meta != nil {
		if v, ok := meta["doc_name"]; ok {
			doc = strings.TrimSpace(fmt.Sprintf("%v", v))
		}
		if v, ok := meta["title"]; ok {
			title = strings.TrimSpace(fmt.Sprintf("%v", v))
		}
	}
	parts := []string{fmt.Sprintf("[ID:%d] · ref #%d", index, index+1)}
	if doc != "" {
		parts = append(parts, "\""+doc+"\"")
	} else if title != "" {
		parts = append(parts, "\""+title+"\"")
	}
	if meta != nil {
		if v, ok := meta["chunk_index"]; ok {
			parts = append(parts, fmt.Sprintf("chunk #%v", v))
		}
	}
	return strings.Join(parts, " ")
}

// buildRAGContextPrompt 构建包含引用指令的 RAG 上下文；graphContext 非空时置于切片块之前（对齐 LightRAG 实体/关系上下文预算）
func buildRAGContextPrompt(docs []ragschema.Document, question, graphContext string) string {
	var sb strings.Builder
	sb.WriteString(ragCitationPrompt)
	if g := strings.TrimSpace(graphContext); g != "" {
		sb.WriteString("<knowledge_graph>\n")
		sb.WriteString(g)
		sb.WriteString("\n</knowledge_graph>\n\n")
	}
	// 与 LightRAG kg_query_context 中 Document Chunks 段说明一致：明确 reference_id 与引用角标 [ID:n] 的对应关系
	sb.WriteString("Document Chunks (each block is labeled ID: n; cite in your answer as [ID:n]):\n\n")
	sb.WriteString("<context>\n")
	for i, d := range docs {
		label := ragRefSourceLabel(d.Metadata, i)
		// Same chunk ID convention as kb_prompt: ID: 0, 1, 2…
		sb.WriteString(fmt.Sprintf("\nID: %d\n", i))
		sb.WriteString(fmt.Sprintf("(Use citation marker [ID:%d] only for the chunk body below.)\n", i))
		sb.WriteString(fmt.Sprintf("Source summary: %s\n", label))
		if d.Metadata != nil {
			if v, ok := d.Metadata["doc_name"]; ok {
				sb.WriteString(fmt.Sprintf("Document name: %v\n", v))
			}
			if v, ok := d.Metadata["title"]; ok && fmt.Sprintf("%v", v) != "" {
				t := fmt.Sprintf("%v", v)
				if d.Metadata["doc_name"] == nil || fmt.Sprintf("%v", d.Metadata["doc_name"]) != t {
					sb.WriteString(fmt.Sprintf("Title / section: %s\n", t))
				}
			}
		}
		sb.WriteString("Chunk body:\n")
		sb.WriteString(d.PageContent)
		sb.WriteString("\n")
	}
	sb.WriteString("\n</context>\n\n")
	sb.WriteString("User question:\n")
	sb.WriteString(question)
	out := sb.String()
	global.LRAG_LOG.Info("buildRAGContextPrompt 已生成（仅在有检索命中时调用）",
		zap.Int("切片数", len(docs)),
		zap.Int("promptLen", len(out)))
	return out
}

// contextOnlyAssistantMessage 与 LightningRAG only_need_context 一致：仅返回检索侧结果，不调用模型生成正文
const contextOnlyAssistantMessage = "（已按 context 模式仅返回检索引用，未调用模型生成回答。）"

// fetchRelevantDocumentsForKnowledgeBases 从指定知识库拉取与 query 相关的切片（对话 RAG 与「检索文档」页共用）
func fetchRelevantDocumentsForKnowledgeBases(ctx context.Context, kbIDs []uint, uid uint, llmInst interfaces.LLM, question string, topN int, session RetrieverSessionOptions) ([]ragschema.Document, error) {
	if topN <= 0 {
		return nil, fmt.Errorf("topN 无效")
	}
	if topN > EffectiveMaxRetrieveTopN() {
		topN = EffectiveMaxRetrieveTopN()
	}
	candidateN := resolveRetrieveCandidateCount(topN, session.RetrievePoolTopK)
	ttlSec := global.LRAG_CONFIG.Rag.RetrieveCacheTTLSeconds
	cacheKey := retrieveCacheKey(uid, kbIDs, question, topN, candidateN, session)
	if ttlSec > 0 {
		if docs, ok := loadRetrieveCache(cacheKey); ok {
			return cloneDocSlice(docs), nil
		}
	}
	v, err, _ := global.LRAG_Concurrency_Control.Do("lrag_ret:"+cacheKey, func() (interface{}, error) {
		if ttlSec > 0 {
			if docs, ok := loadRetrieveCache(cacheKey); ok {
				return docs, nil
			}
		}
		docs, ferr := fetchRelevantDocumentsUncached(ctx, kbIDs, uid, llmInst, question, topN, session, candidateN)
		if ferr != nil {
			return nil, ferr
		}
		if ttlSec > 0 && len(docs) > 0 {
			saveRetrieveCache(cacheKey, docs, ttlSec)
		}
		return docs, nil
	})
	if err != nil {
		return nil, err
	}
	docs, _ := v.([]ragschema.Document)
	return cloneDocSlice(docs), nil
}

func fetchRelevantDocumentsUncached(ctx context.Context, kbIDs []uint, uid uint, llmInst interfaces.LLM, question string, topN int, session RetrieverSessionOptions, candidateN int) ([]ragschema.Document, error) {
	retriever, err := CreateRetrieverForKnowledgeBases(ctx, kbIDs, uid, topN, session, llmInst)
	if err != nil {
		return nil, err
	}
	if retriever == nil {
		return nil, fmt.Errorf("检索器不可用")
	}
	docs, err := retriever.GetRelevantDocuments(ctx, question, candidateN)
	if err != nil {
		return nil, err
	}
	docs = ragretriever.DeduplicateRetrievedDocuments(docs)
	if len(docs) > topN {
		docs = docs[:topN]
	}
	return docs, nil
}

// ragDocumentsToRefMaps 将检索切片转为与对话引用一致的结构
func ragDocumentsToRefMaps(docs []ragschema.Document) []map[string]any {
	var refs []map[string]any
	for i, d := range docs {
		ref := map[string]any{
			"content":     d.PageContent,
			"score":       d.Score,
			"index":       i,
			"sourceLabel": ragRefSourceLabel(d.Metadata, i),
		}
		if d.Metadata != nil {
			if v, ok := d.Metadata["document_id"]; ok {
				ref["documentId"] = v
			}
			if v, ok := d.Metadata["doc_name"]; ok {
				ref["docName"] = v
			}
			if v, ok := d.Metadata["chunk_index"]; ok {
				ref["chunkIndex"] = v
			}
			if v, ok := d.Metadata["node_id"]; ok {
				ref["nodeId"] = v
			}
			if v, ok := d.Metadata["title"]; ok {
				ref["title"] = v
			}
		}
		refs = append(refs, ref)
	}
	return refs
}

// retrieveAndBuildRAGContext 从知识库检索并构建带引用指令的上下文，返回 refs 和 userContent
// displayQuestion 写入 <context> 中的用户问题；searchQuestion 用于向量/关键词检索（可含 hl/ll 关键词扩展）
// maxRagChunkTokens>0 时按粗估 token 限制切片正文总量（见 trimDocsToRagTokenBudget）
// maxEntityTokens/maxRelationTokens>0 时从命中切片关联图谱生成摘要并注入 prompt（见 LightRAG max_entity_tokens / max_relation_tokens）
func retrieveAndBuildRAGContext(ctx context.Context, kbIDs []uint, uid uint, llmInst interfaces.LLM, displayQuestion, searchQuestion string, topK int, session RetrieverSessionOptions, maxRagChunkTokens, maxEntityTokens, maxRelationTokens uint) ([]map[string]any, string) {
	topK = ClampRetrieveTopN(topK, DefaultConversationChunkTopKFromConfig())
	docs, rerr := fetchRelevantDocumentsForKnowledgeBases(ctx, kbIDs, uid, llmInst, searchQuestion, topK, session)
	if rerr != nil {
		global.LRAG_LOG.Warn("创建检索器或检索失败", zap.Any("kbIDs", kbIDs), zap.Error(rerr))
		return nil, displayQuestion
	}
	if len(docs) == 0 {
		global.LRAG_LOG.Warn("检索结果为空", zap.Any("kbIDs", kbIDs), zap.String("searchQuestion", searchQuestion))
		return nil, displayQuestion
	}
	docs = trimDocsToRagTokenBudget(docs, maxRagChunkTokens)
	if len(docs) == 0 {
		global.LRAG_LOG.Warn("RAG token 预算裁剪后无切片", zap.Any("kbIDs", kbIDs))
		return nil, displayQuestion
	}
	global.LRAG_LOG.Info("检索到文档切片", zap.Int("count", len(docs)), zap.Any("kbIDs", kbIDs))
	refs := ragDocumentsToRefMaps(docs)
	graphCtx := ""
	if maxEntityTokens > 0 || maxRelationTokens > 0 {
		cids := ChunkIDsFromRAGDocs(ctx, kbIDs, docs)
		if len(cids) > 0 {
			entMaps, relMaps := KnowledgeGraphMapsForChunkIDs(ctx, kbIDs, cids)
			graphCtx = FormatKnowledgeGraphPromptPrefix(entMaps, relMaps, maxEntityTokens, maxRelationTokens)
		}
	}
	userContent := buildRAGContextPrompt(docs, displayQuestion, graphCtx)
	return refs, userContent
}

// mergeGlobalKnowledgeBaseIDs 将全局共享知识库 ID 合并到用户知识库 ID 中（去重）
func mergeGlobalKnowledgeBaseIDs(ctx context.Context, userKbIDs []uint) []uint {
	globalIDs, err := systemModelService.GetGlobalKnowledgeBaseIDs(ctx)
	if err != nil || len(globalIDs) == 0 {
		return userKbIDs
	}
	seen := make(map[uint]bool, len(userKbIDs))
	for _, id := range userKbIDs {
		seen[id] = true
	}
	merged := make([]uint, len(userKbIDs))
	copy(merged, userKbIDs)
	for _, gid := range globalIDs {
		if !seen[gid] {
			merged = append(merged, gid)
			seen[gid] = true
		}
	}
	return merged
}

// fixOllamaProvider 当 provider 误存为 openai 但 baseURL 为空且模型名像 Ollama 时，自动修正为 ollama
func fixOllamaProvider(provider, baseURL, modelName string) string {
	if strings.ToLower(provider) != "openai" || baseURL != "" {
		return provider
	}
	m := strings.ToLower(modelName)
	ollamaModels := []string{"deepseek", "llama", "mistral", "qwen", "phi", "gemma", "codellama", "neural", "nous", "solar", "aya", "command", "dolphin"}
	for _, prefix := range ollamaModels {
		if strings.Contains(m, prefix) {
			return "ollama"
		}
	}
	return provider
}

// Create 创建对话
func (s *ConversationService) Create(ctx context.Context, uid uint, authorityId uint, req request.ConversationCreate) (*rag.RagConversation, error) {
	llmProviderID := req.LLMProviderID
	llmSource := req.LLMSource
	if llmProviderID == 0 {
		// 未指定模型时，按优先级解析默认：用户默认 > 角色默认
		if id, src, _, ok := llmProviderService.ResolveDefaultLLM(ctx, uid, authorityId, "chat"); ok {
			llmProviderID, llmSource = id, src
		} else {
			return nil, fmt.Errorf("请选择模型或先设置默认模型")
		}
	}
	if llmSource == "" {
		llmSource = "user"
	}
	enabledToolNames := ""
	if len(req.EnabledToolNames) > 0 {
		b, _ := json.Marshal(req.EnabledToolNames)
		enabledToolNames = string(b)
	}
	conv := &rag.RagConversation{
		UUID:             uuid.New(),
		UserID:           uid,
		Title:            req.Title,
		LLMProviderID:    llmProviderID,
		LLMSource:        llmSource,
		SourceType:       req.SourceType,
		SourceIDs:        req.SourceIDs,
		EnabledToolNames: enabledToolNames,
	}
	if conv.Title == "" {
		conv.Title = "新对话"
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

// Update 更新对话（如修改启用的工具）
func (s *ConversationService) Update(ctx context.Context, uid uint, req request.ConversationUpdate) error {
	var conv rag.RagConversation
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", req.ID, uid).First(&conv).Error; err != nil {
		return err
	}
	enabledToolNames := ""
	if len(req.EnabledToolNames) > 0 {
		b, _ := json.Marshal(req.EnabledToolNames)
		enabledToolNames = string(b)
	}
	return global.LRAG_DB.WithContext(ctx).Model(&conv).Update("enabled_tool_names", enabledToolNames).Error
}

// parseEnabledToolNames 解析对话启用的工具名称，空则返回 nil
func parseEnabledToolNames(s string) []string {
	if s == "" {
		return nil
	}
	var names []string
	if err := json.Unmarshal([]byte(s), &names); err != nil {
		return nil
	}
	return names
}

// keywordHistoryHint 将请求中的多轮摘录压缩为抽词用上下文（对齐 LightRAG 在关键词抽取中利用 history 的思想）
func keywordHistoryHint(items []request.ConversationHistoryItem) string {
	if len(items) == 0 {
		return ""
	}
	maxHistoryTurns := EffectiveConversationHistoryLimit()
	if len(items) > maxHistoryTurns {
		items = items[len(items)-maxHistoryTurns:]
	}
	const maxRunes = 1200
	var sb strings.Builder
	for _, it := range items {
		role := strings.TrimSpace(it.Role)
		c := strings.TrimSpace(it.Content)
		if c == "" {
			continue
		}
		sb.WriteString(role)
		sb.WriteString(": ")
		sb.WriteString(c)
		sb.WriteByte('\n')
	}
	s := sb.String()
	r := []rune(s)
	if len(r) > maxRunes {
		s = string(r[:maxRunes]) + "…"
	}
	return strings.TrimSpace(s)
}

// defaultSystemPrompt 创建会话时使用的系统角色提示，定义 AI 身份
const defaultSystemPrompt = "You are LightningRAG, a capable AI assistant. Answer questions, give professional guidance, and help users accomplish tasks. You may use the provided tools (e.g. web search) for up-to-date information. Respond in a friendly, professional manner."

// buildSystemPrompt 根据启用的工具构建系统提示；当启用 web_search 时强化联网搜索指令
func buildSystemPrompt(enabledNames []string) string {
	if len(enabledNames) == 0 {
		return defaultSystemPrompt
	}
	hasWebSearch := false
	for _, n := range enabledNames {
		if n == "web_search" {
			hasWebSearch = true
			break
		}
	}
	if !hasWebSearch {
		return defaultSystemPrompt
	}
	return defaultSystemPrompt + " Important: When the user clearly asks to search the web, get real-time information, or the latest news, call the web_search tool directly. Do not only say you will search without invoking the tool."
}

// buildSystemPromptForChat 在基础系统提示上叠加 LightningRAG 风格的 response_type / user_prompt / 回答语言（见 references/LightRAG lightrag_server addon_params.language）
func buildSystemPromptForChat(enabledNames []string, responseType, userPrompt, responseLanguage string) string {
	base := buildSystemPrompt(enabledNames)
	var parts []string
	if rt := strings.TrimSpace(responseType); rt != "" {
		parts = append(parts, "For this turn, format your answer as: "+rt+" (Markdown is welcome when it helps clarity).")
	}
	if up := strings.TrimSpace(userPrompt); up != "" {
		parts = append(parts, "Additional instructions for this turn:\n"+up)
	}
	if lang := strings.TrimSpace(responseLanguage); lang != "" {
		parts = append(parts, "For this turn, use "+lang+" for your reply unless the user explicitly requires another language.")
	}
	if len(parts) == 0 {
		return base
	}
	return base + "\n\n" + strings.Join(parts, "\n\n")
}

// formatMessageContentsForExport 将本轮将发给模型的消息序列化为可读文本（用于 only_need_prompt 类调试）
func formatMessageContentsForExport(msgs []interfaces.MessageContent) string {
	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString(strings.ToUpper(string(m.Role)))
		sb.WriteString(":\n")
		for _, p := range m.Parts {
			if t, ok := p.(interfaces.TextContent); ok {
				sb.WriteString(t.Text)
				sb.WriteByte('\n')
			}
		}
		sb.WriteString("\n---\n")
	}
	return sb.String()
}

// loadConversationHistory 加载对话历史消息（最近 N 条，按时间正序），用于多轮上下文
func loadConversationHistory(ctx context.Context, conversationID uint, limit int) ([]rag.RagMessage, error) {
	if limit <= 0 {
		limit = EffectiveConversationHistoryLimit()
	}
	var history []rag.RagMessage
	err := global.LRAG_DB.WithContext(ctx).Where("conversation_id = ?", conversationID).
		Order("created_at DESC").Limit(limit).Find(&history).Error
	if err != nil {
		return nil, err
	}
	// 反转使时间正序（从早到晚）
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
	return history, nil
}

// logMessagesForDebug 打印发送给大模型的多轮对话上下文，便于调试
func logMessagesForDebug(messages []interfaces.MessageContent) {
	const maxPreview = 200
	for i, m := range messages {
		text := ""
		for _, p := range m.Parts {
			if t, ok := p.(interfaces.TextContent); ok {
				text += t.Text
			}
		}
		preview := text
		if len(preview) > maxPreview {
			preview = preview[:maxPreview] + "..."
		}
		global.LRAG_LOG.Info("多轮上下文消息",
			zap.Int("index", i+1),
			zap.String("role", string(m.Role)),
			zap.Int("contentLen", len(text)),
			zap.String("contentPreview", preview))
	}
}

// historyToMessages 将存储的消息转为 LLM 接口所需格式
func historyToMessages(history []rag.RagMessage) []interfaces.MessageContent {
	msgs := make([]interfaces.MessageContent, 0, len(history))
	for _, m := range history {
		role := interfaces.MessageRoleHuman
		switch m.Role {
		case "assistant":
			role = interfaces.MessageRoleAssistant
		case "system":
			role = interfaces.MessageRoleSystem
		}
		text := m.Content
		if role == interfaces.MessageRoleAssistant {
			text = llm.StripAssistantReasoningMarkers(text)
		}
		msgs = append(msgs, interfaces.MessageContent{
			Role:  role,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: text}},
		})
	}
	return msgs
}

// resolvedLLMConfig 解析后的 LLM 配置
type resolvedLLMConfig struct {
	Provider             string
	ModelName            string
	BaseURL              string
	APIKey               string
	MaxContextTokens     uint
	SupportsDeepThinking bool
	SupportsToolCall     bool
}

// resolveLLMConfig 解析对话使用的模型：优先请求中的临时覆盖，否则用对话创建时的模型
func resolveLLMConfig(ctx context.Context, uid uint, conv *rag.RagConversation, overrideProviderID uint, overrideSource string) (*resolvedLLMConfig, error) {
	providerID := conv.LLMProviderID
	source := conv.LLMSource
	if overrideProviderID > 0 {
		providerID, source = overrideProviderID, overrideSource
		if source == "" {
			source = "user"
		}
	}
	if source == "" {
		source = "user"
	}
	cfg := &resolvedLLMConfig{}
	if source == "admin" {
		var adminLLM rag.RagLLMProvider
		if e := global.LRAG_DB.WithContext(ctx).Where("id = ?", providerID).First(&adminLLM).Error; e != nil {
			return nil, e
		}
		cfg.Provider, cfg.ModelName, cfg.BaseURL, cfg.APIKey = adminLLM.Name, adminLLM.ModelName, adminLLM.BaseURL, adminLLM.APIKey
		cfg.MaxContextTokens = adminLLM.MaxContextTokens
		cfg.SupportsDeepThinking = adminLLM.SupportsDeepThinking
		cfg.SupportsToolCall = adminLLM.SupportsToolCall
	} else {
		var userLLM rag.RagUserLLM
		if e := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", providerID, uid).First(&userLLM).Error; e != nil {
			return nil, e
		}
		cfg.Provider, cfg.ModelName, cfg.BaseURL, cfg.APIKey = userLLM.Provider, userLLM.ModelName, userLLM.BaseURL, userLLM.APIKey
		cfg.MaxContextTokens = userLLM.MaxContextTokens
		cfg.SupportsDeepThinking = userLLM.SupportsDeepThinking
		cfg.SupportsToolCall = userLLM.SupportsToolCall
	}
	cfg.Provider = fixOllamaProvider(cfg.Provider, cfg.BaseURL, cfg.ModelName)
	return cfg, nil
}

// estimateTokens 粗略估算文本的 token 数（约 4 字符/token，适用于中英文混合）
func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + 3) / 4
}

// truncateMessagesToFit 若 totalTokens 超过 maxContextTokens，从历史消息开头移除直到满足限制；保留系统提示和当前用户消息
func truncateMessagesToFit(messages []interfaces.MessageContent, maxContextTokens uint) []interfaces.MessageContent {
	if maxContextTokens == 0 || len(messages) <= 2 {
		return messages
	}
	var total int
	for _, m := range messages {
		for _, p := range m.Parts {
			if t, ok := p.(interfaces.TextContent); ok {
				total += estimateTokens(t.Text)
			}
		}
	}
	if total <= int(maxContextTokens) {
		return messages
	}
	// 保留 system(0)、最后一条 user 消息；从中间历史消息移除最老的
	maxTokens := int(maxContextTokens)
	systemTokens := 0
	for _, p := range messages[0].Parts {
		if t, ok := p.(interfaces.TextContent); ok {
			systemTokens += estimateTokens(t.Text)
		}
	}
	lastUserTokens := 0
	for _, p := range messages[len(messages)-1].Parts {
		if t, ok := p.(interfaces.TextContent); ok {
			lastUserTokens += estimateTokens(t.Text)
		}
	}
	allowed := maxTokens - systemTokens - lastUserTokens
	if allowed <= 0 {
		return append(messages[:1], messages[len(messages)-1])
	}
	// 从第 2 条开始，保留能塞进 allowed 的最近历史
	history := messages[1 : len(messages)-1]
	kept := make([]interfaces.MessageContent, 0, len(history))
	used := 0
	for i := len(history) - 1; i >= 0; i-- {
		t := 0
		for _, p := range history[i].Parts {
			if tc, ok := p.(interfaces.TextContent); ok {
				t += estimateTokens(tc.Text)
			}
		}
		if used+t > allowed {
			break
		}
		used += t
		kept = append([]interfaces.MessageContent{history[i]}, kept...)
	}
	return append(append(messages[:1], kept...), messages[len(messages)-1])
}

// Chat 对话，当关联知识库时进行检索增强；支持请求中 llmProviderId/llmSource 临时覆盖模型
func (s *ConversationService) Chat(ctx context.Context, uid uint, req request.ConversationChat) (any, error) {
	var conv rag.RagConversation
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", req.ConversationID, uid).First(&conv).Error; err != nil {
		return nil, err
	}
	cfg, err := resolveLLMConfig(ctx, uid, &conv, req.LLMProviderID, req.LLMSource)
	if err != nil {
		return nil, err
	}
	global.LRAG_LOG.Info("对话使用模型",
		zap.String("provider", cfg.Provider),
		zap.String("modelName", cfg.ModelName),
		zap.String("baseURL", cfg.BaseURL))
	llmInst, err := registry.CreateLLM(registry.LLMConfig{
		Provider:  cfg.Provider,
		ModelName: cfg.ModelName,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
	})
	if err != nil {
		return nil, err
	}
	if llmInst == nil {
		return nil, fmt.Errorf("不支持的模型提供商: %s，请确认 provider 为 openai 或 ollama", cfg.Provider)
	}

	// 检索增强：合并用户选择的知识库和全局共享知识库
	var refs []map[string]any
	userContent := req.Content
	kbIDs, parseErr := ParseKnowledgeBaseIDs(conv.SourceType, conv.SourceIDs)
	if parseErr != nil {
		global.LRAG_LOG.Warn("解析对话 sourceIds 失败，跳过检索增强", zap.String("sourceIDs", conv.SourceIDs), zap.Error(parseErr))
	}
	kbIDs = mergeGlobalKnowledgeBaseIDs(ctx, kbIDs)
	q, modeOv, onlyNeedContext := ResolveLightningRAGQueryModeAndQuestion(req.Content, req.QueryMode)
	effectiveMode := modeOv
	if effectiveMode == "" {
		effectiveMode = "kb_default"
	}
	global.LRAG_LOG.Info("对话关联知识库(非流式)",
		zap.Int("conversationId", int(req.ConversationID)),
		zap.String("sourceType", conv.SourceType),
		zap.String("sourceIDs", conv.SourceIDs),
		zap.Any("parsedKbIDs", kbIDs))
	session := RetrieverSessionFromLightningRAGParams(modeOv, req.TopK, req.EnableRerank, req.CosineThreshold, req.MinRerankScore)
	ApplyPageIndexTocEnhanceFromRequest(&session, request.EffectiveTocEnhance(req.TocEnhance, req.TocEnhanceRagflow))
	var searchQ string
	var resolvedHl, resolvedLl []string
	if len(kbIDs) > 0 && strings.TrimSpace(q) != "" {
		var combined, kgEnt, kgRel string
		combined, kgEnt, kgRel, resolvedHl, resolvedLl = PrepareLightningRAGSearchQueries(ctx, uid, llmInst, q, req.HlKeywords, req.LlKeywords, keywordHistoryHint(req.ConversationHistory))
		searchQ = combined
		session.KgEntitySearchQuery = kgEnt
		session.KgRelSearchQuery = kgRel
	} else {
		searchQ = AugmentQueryWithLightningRAGKeywords(q, req.HlKeywords, req.LlKeywords)
		resolvedHl = trimKeywordSlice(req.HlKeywords)
		resolvedLl = trimKeywordSlice(req.LlKeywords)
	}
	ragTopK := ClampRetrieveTopN(req.ChunkTopK, DefaultConversationChunkTopKFromConfig())
	ApplyDefaultConversationRetrievePoolIfNeeded(&session, ragTopK)
	if len(kbIDs) > 0 && strings.TrimSpace(q) != "" {
		global.LRAG_LOG.Info("检索增强启动(非流式)", zap.Any("mergedKbIDs", kbIDs), zap.String("retrievalQuery", q), zap.String("searchQuery", searchQ), zap.String("modeOverride", modeOv), zap.Int("chunkTopK", ragTopK))
		ragTok := EffectiveMaxRagContextTokens(req.MaxRagContextTokens)
		maxEt := EffectiveMaxEntityContextTokens(req.MaxEntityTokens)
		maxRt := EffectiveMaxRelationContextTokens(req.MaxRelationTokens)
		refs, userContent = retrieveAndBuildRAGContext(ctx, kbIDs, uid, llmInst, q, searchQ, ragTopK, session, ragTok, maxEt, maxRt)
		global.LRAG_LOG.Info("检索增强结果(非流式)", zap.Int("refsCount", len(refs)), zap.Int("userContentLen", len(userContent)))
		if len(refs) == 0 {
			global.LRAG_LOG.Info("检索未命中或检索器不可用，本轮未注入 RAG 上下文；buildRAGContextPrompt 未执行，前端无引用角标数据")
			userContent = ragNoMatchedChunksHint + strings.TrimSpace(q)
		}
	} else {
		st := strings.TrimSpace(strings.ToLower(conv.SourceType))
		if st == "files" {
			global.LRAG_LOG.Info("跳过检索增强(非流式)：sourceType=files 时当前对话接口仅对 knowledge_base 做向量检索与引用")
		} else {
			global.LRAG_LOG.Info("未关联知识库，跳过检索增强(非流式)；buildRAGContextPrompt 不会执行")
		}
	}

	apiRefs := ExposeReferencesForAPI(refs, req.IncludeReferences, req.IncludeChunkContent)

	// LightningRAG：*context 与 /context 仅返回检索上下文，不调用 LLM
	if onlyNeedContext && len(kbIDs) > 0 {
		userMsg := &rag.RagMessage{
			UUID:           uuid.New(),
			ConversationID: req.ConversationID,
			Role:           "user",
			Content:        req.Content,
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(userMsg).Error; err != nil {
			global.LRAG_LOG.Warn("failed to save user message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
		}
		refsJSON := ""
		if len(refs) > 0 {
			if b, err := json.Marshal(refs); err == nil {
				refsJSON = string(b)
			}
		}
		assistantMsg := &rag.RagMessage{
			UUID:           uuid.New(),
			ConversationID: req.ConversationID,
			Role:           "assistant",
			Content:        contextOnlyAssistantMessage,
			References:     refsJSON,
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(assistantMsg).Error; err != nil {
			global.LRAG_LOG.Warn("failed to save assistant message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
		}
		return map[string]any{
			"content":         contextOnlyAssistantMessage,
			"references":      apiRefs,
			"onlyNeedContext": true,
			"onlyNeedPrompt":  false,
			"retrievalMode":   effectiveMode,
			"retrievalQuery":  q,
			"searchQuery":     searchQ,
			"hlKeywords":      KeywordsForAPIResponse(resolvedHl),
			"llKeywords":      KeywordsForAPIResponse(resolvedLl),
		}, nil
	}

	// 多轮上下文：系统角色 + 历史消息 + 当前用户消息
	history, _ := loadConversationHistory(ctx, req.ConversationID, EffectiveConversationHistoryLimit())
	enabledNames := parseEnabledToolNames(conv.EnabledToolNames)
	messages := make([]interfaces.MessageContent, 0, len(history)+2)
	messages = append(messages, interfaces.MessageContent{
		Role:  interfaces.MessageRoleSystem,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: buildSystemPromptForChat(enabledNames, req.ResponseType, req.UserPrompt, req.ResponseLanguage)}},
	})
	messages = append(messages, historyToMessages(history)...)
	messages = append(messages, conversationHistoryItemsToMessages(req.ConversationHistory)...)
	messages = append(messages, interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: userContent}},
	})
	messages = truncateMessagesToFit(messages, effectiveTruncationBudget(cfg.MaxContextTokens, req.MaxTotalTokens))
	global.LRAG_LOG.Info("发送给大模型的消息列表", zap.Int("conversationId", int(req.ConversationID)), zap.Int("messageCount", len(messages)))
	logMessagesForDebug(messages)

	if req.OnlyNeedPrompt {
		return map[string]any{
			"onlyNeedPrompt":  true,
			"prompt":          formatMessageContentsForExport(messages),
			"references":      apiRefs,
			"retrievalMode":   effectiveMode,
			"retrievalQuery":  q,
			"searchQuery":     searchQ,
			"onlyNeedContext": false,
			"hlKeywords":      KeywordsForAPIResponse(resolvedHl),
			"llKeywords":      KeywordsForAPIResponse(resolvedLl),
		}, nil
	}

	// 工具调用循环（非流式）：仅当模型支持工具时使用用户选中的工具
	openAITools := tools.ToOpenAIToolsForNames(enabledNames)
	llmOpts := []interfaces.CallOption{}
	if cfg.SupportsToolCall && len(openAITools) > 0 {
		llmOpts = append(llmOpts, interfaces.WithTools(openAITools))
		global.LRAG_LOG.Info("对话已启用工具", zap.Strings("tools", enabledNames))
	} else if len(enabledNames) > 0 {
		global.LRAG_LOG.Warn("对话配置了工具但未生效", zap.Bool("supportsToolCall", cfg.SupportsToolCall), zap.Strings("enabledNames", enabledNames))
	}
	if cfg.SupportsDeepThinking && req.UseDeepThinking {
		llmOpts = append(llmOpts, interfaces.WithReasoningEffort("high"))
	}
	var fullContent string
	for round := 0; round < maxToolRounds; round++ {
		resp, rerr := llmInst.GenerateContent(ctx, messages, llmOpts...)
		if rerr != nil {
			return nil, rerr
		}
		if len(resp.Choices) == 0 {
			break
		}
		choice := resp.Choices[0]
		if len(choice.ToolCalls) > 0 {
			assistantMsg := interfaces.MessageContent{
				Role:      interfaces.MessageRoleAssistant,
				ToolCalls: choice.ToolCalls,
			}
			if tcText := strings.TrimSpace(llm.StripAssistantReasoningMarkers(choice.Content)); tcText != "" {
				assistantMsg.Parts = []interfaces.ContentPart{interfaces.TextContent{Text: tcText}}
			}
			messages = append(messages, assistantMsg)
			for _, tc := range choice.ToolCalls {
				toolCtx := tools.WithUserID(ctx, uid)
				result, execErr := tools.ExecuteTool(toolCtx, tc.Name, tc.Arguments)
				if execErr != nil {
					result = "Tool execution failed: " + execErr.Error()
				}
				toolResultMsg := interfaces.MessageContent{
					Role:       interfaces.MessageRoleTool,
					ToolCallID: tc.ID,
					ToolName:   tc.Name,
					Parts:      []interfaces.ContentPart{interfaces.TextContent{Text: result}},
				}
				messages = append(messages, toolResultMsg)
			}
			continue
		}
		fullContent = choice.Content
		break
	}

	userMsg := &rag.RagMessage{
		UUID:           uuid.New(),
		ConversationID: req.ConversationID,
		Role:           "user",
		Content:        req.Content,
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(userMsg).Error; err != nil {
		global.LRAG_LOG.Warn("failed to save user message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
	}
	refsJSON := ""
	if len(refs) > 0 {
		if b, err := json.Marshal(refs); err == nil {
			refsJSON = string(b)
		}
	}
	assistantMsg := &rag.RagMessage{
		UUID:           uuid.New(),
		ConversationID: req.ConversationID,
		Role:           "assistant",
		Content:        fullContent,
		References:     refsJSON,
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(assistantMsg).Error; err != nil {
		global.LRAG_LOG.Warn("failed to save assistant message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
	}
	return map[string]any{
		"content":         fullContent,
		"references":      apiRefs,
		"onlyNeedContext": false,
		"onlyNeedPrompt":  false,
		"retrievalMode":   effectiveMode,
		"retrievalQuery":  q,
		"searchQuery":     searchQ,
		"hlKeywords":      KeywordsForAPIResponse(resolvedHl),
		"llKeywords":      KeywordsForAPIResponse(resolvedLl),
	}, nil
}

// ChatStreamResult 流式对话结果，包含完整内容和引用
type ChatStreamResult struct {
	Content         string
	References      []map[string]any
	OnlyNeedContext bool
	OnlyNeedPrompt  bool
	RetrievalMode   string
	RetrievalQuery  string
	SearchQuery     string
	HlKeywords      []string
	LlKeywords      []string
}

// maxToolRounds 工具调用最大轮数，防止无限循环
const maxToolRounds = 5

// ChatStreamCallbacks 流式对话回调
type ChatStreamCallbacks struct {
	OnChunk      func(string)
	OnToolCall   func(name string, status string, result string) // status: "start"|"done"
	OnReferences func(refs []map[string]any)                     // 检索完成后、LLM 生成前，提前发送引用数据
}

// ChatStream 流式对话，每收到一块内容调用 opts.OnChunk(delta)；支持工具调用；支持请求中 llmProviderId/llmSource 临时覆盖模型
func (s *ConversationService) ChatStream(ctx context.Context, uid uint, req request.ConversationChat, opts ChatStreamCallbacks) (*ChatStreamResult, error) {
	var conv rag.RagConversation
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", req.ConversationID, uid).First(&conv).Error; err != nil {
		return nil, err
	}
	cfg, err := resolveLLMConfig(ctx, uid, &conv, req.LLMProviderID, req.LLMSource)
	if err != nil {
		return nil, err
	}
	llmInst, err := registry.CreateLLM(registry.LLMConfig{
		Provider:  cfg.Provider,
		ModelName: cfg.ModelName,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
	})
	if err != nil || llmInst == nil {
		return nil, fmt.Errorf("不支持的模型提供商: %s", cfg.Provider)
	}

	// 检索增强：合并用户选择的知识库和全局共享知识库
	var refs []map[string]any
	userContent := req.Content
	kbIDs, parseErr := ParseKnowledgeBaseIDs(conv.SourceType, conv.SourceIDs)
	if parseErr != nil {
		global.LRAG_LOG.Warn("解析对话 sourceIds 失败，跳过检索增强", zap.String("sourceIDs", conv.SourceIDs), zap.Error(parseErr))
	}
	kbIDs = mergeGlobalKnowledgeBaseIDs(ctx, kbIDs)
	q, modeOv, onlyNeedContext := ResolveLightningRAGQueryModeAndQuestion(req.Content, req.QueryMode)
	effectiveMode := modeOv
	if effectiveMode == "" {
		effectiveMode = "kb_default"
	}
	global.LRAG_LOG.Info("对话关联知识库(流式)",
		zap.Int("conversationId", int(req.ConversationID)),
		zap.String("sourceType", conv.SourceType),
		zap.String("sourceIDs", conv.SourceIDs),
		zap.Any("parsedKbIDs", kbIDs))
	session := RetrieverSessionFromLightningRAGParams(modeOv, req.TopK, req.EnableRerank, req.CosineThreshold, req.MinRerankScore)
	ApplyPageIndexTocEnhanceFromRequest(&session, request.EffectiveTocEnhance(req.TocEnhance, req.TocEnhanceRagflow))
	var searchQ string
	var resolvedHl, resolvedLl []string
	if len(kbIDs) > 0 && strings.TrimSpace(q) != "" {
		var combined, kgEnt, kgRel string
		combined, kgEnt, kgRel, resolvedHl, resolvedLl = PrepareLightningRAGSearchQueries(ctx, uid, llmInst, q, req.HlKeywords, req.LlKeywords, keywordHistoryHint(req.ConversationHistory))
		searchQ = combined
		session.KgEntitySearchQuery = kgEnt
		session.KgRelSearchQuery = kgRel
	} else {
		searchQ = AugmentQueryWithLightningRAGKeywords(q, req.HlKeywords, req.LlKeywords)
		resolvedHl = trimKeywordSlice(req.HlKeywords)
		resolvedLl = trimKeywordSlice(req.LlKeywords)
	}
	ragTopK := ClampRetrieveTopN(req.ChunkTopK, DefaultConversationChunkTopKFromConfig())
	ApplyDefaultConversationRetrievePoolIfNeeded(&session, ragTopK)
	if len(kbIDs) > 0 && strings.TrimSpace(q) != "" {
		global.LRAG_LOG.Info("检索增强启动(流式)", zap.Any("mergedKbIDs", kbIDs), zap.String("retrievalQuery", q), zap.String("searchQuery", searchQ), zap.String("modeOverride", modeOv), zap.Int("chunkTopK", ragTopK))
		ragTok := EffectiveMaxRagContextTokens(req.MaxRagContextTokens)
		maxEt := EffectiveMaxEntityContextTokens(req.MaxEntityTokens)
		maxRt := EffectiveMaxRelationContextTokens(req.MaxRelationTokens)
		refs, userContent = retrieveAndBuildRAGContext(ctx, kbIDs, uid, llmInst, q, searchQ, ragTopK, session, ragTok, maxEt, maxRt)
		global.LRAG_LOG.Info("检索增强结果(流式)", zap.Int("refsCount", len(refs)), zap.Int("userContentLen", len(userContent)))
		if len(refs) == 0 {
			global.LRAG_LOG.Info("检索未命中或检索器不可用，本轮未注入 RAG 上下文；buildRAGContextPrompt 未执行，前端无引用角标数据")
			userContent = ragNoMatchedChunksHint + strings.TrimSpace(q)
		}
	} else {
		st := strings.TrimSpace(strings.ToLower(conv.SourceType))
		if st == "files" {
			global.LRAG_LOG.Info("跳过检索增强(流式)：sourceType=files 时当前对话接口仅对 knowledge_base 做向量检索与引用")
		} else {
			global.LRAG_LOG.Info("未关联知识库，跳过检索增强(流式)；buildRAGContextPrompt 不会执行")
		}
	}
	apiRefs := ExposeReferencesForAPI(refs, req.IncludeReferences, req.IncludeChunkContent)
	if len(apiRefs) > 0 && opts.OnReferences != nil {
		opts.OnReferences(apiRefs)
	}

	if onlyNeedContext && len(kbIDs) > 0 {
		if opts.OnChunk != nil {
			opts.OnChunk(contextOnlyAssistantMessage)
		}
		userMsg := &rag.RagMessage{
			UUID:           uuid.New(),
			ConversationID: req.ConversationID,
			Role:           "user",
			Content:        req.Content,
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(userMsg).Error; err != nil {
			global.LRAG_LOG.Warn("failed to save user message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
		}
		streamRefsJSON := ""
		if len(refs) > 0 {
			if b, err := json.Marshal(refs); err == nil {
				streamRefsJSON = string(b)
			}
		}
		assistantMsg := &rag.RagMessage{
			UUID:           uuid.New(),
			ConversationID: req.ConversationID,
			Role:           "assistant",
			Content:        contextOnlyAssistantMessage,
			References:     streamRefsJSON,
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(assistantMsg).Error; err != nil {
			global.LRAG_LOG.Warn("failed to save assistant message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
		}
		return &ChatStreamResult{
			Content:         contextOnlyAssistantMessage,
			References:      apiRefs,
			OnlyNeedContext: true,
			OnlyNeedPrompt:  false,
			RetrievalMode:   effectiveMode,
			RetrievalQuery:  q,
			SearchQuery:     searchQ,
			HlKeywords:      KeywordsForAPIResponse(resolvedHl),
			LlKeywords:      KeywordsForAPIResponse(resolvedLl),
		}, nil
	}

	// 多轮上下文：系统角色 + 历史消息 + 当前用户消息
	history, _ := loadConversationHistory(ctx, req.ConversationID, EffectiveConversationHistoryLimit())
	enabledNames := parseEnabledToolNames(conv.EnabledToolNames)
	messages := make([]interfaces.MessageContent, 0, len(history)+2)
	messages = append(messages, interfaces.MessageContent{
		Role:  interfaces.MessageRoleSystem,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: buildSystemPromptForChat(enabledNames, req.ResponseType, req.UserPrompt, req.ResponseLanguage)}},
	})
	messages = append(messages, historyToMessages(history)...)
	messages = append(messages, conversationHistoryItemsToMessages(req.ConversationHistory)...)
	messages = append(messages, interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: userContent}},
	})
	messages = truncateMessagesToFit(messages, effectiveTruncationBudget(cfg.MaxContextTokens, req.MaxTotalTokens))
	global.LRAG_LOG.Info("发送给大模型的消息列表(流式)", zap.Int("conversationId", int(req.ConversationID)), zap.Int("messageCount", len(messages)))
	logMessagesForDebug(messages)

	if req.OnlyNeedPrompt {
		export := formatMessageContentsForExport(messages)
		if opts.OnChunk != nil {
			opts.OnChunk(export)
		}
		return &ChatStreamResult{
			Content:         export,
			References:      apiRefs,
			OnlyNeedContext: false,
			OnlyNeedPrompt:  true,
			RetrievalMode:   effectiveMode,
			RetrievalQuery:  q,
			SearchQuery:     searchQ,
			HlKeywords:      KeywordsForAPIResponse(resolvedHl),
			LlKeywords:      KeywordsForAPIResponse(resolvedLl),
		}, nil
	}

	// 工具调用循环：仅当模型支持工具时使用用户选中的工具
	openAITools := tools.ToOpenAIToolsForNames(enabledNames)
	llmOpts := []interfaces.CallOption{}
	if opts.OnChunk != nil {
		llmOpts = append(llmOpts, interfaces.WithStreamCallback(opts.OnChunk))
	}
	if cfg.SupportsToolCall && len(openAITools) > 0 {
		llmOpts = append(llmOpts, interfaces.WithTools(openAITools))
		global.LRAG_LOG.Info("对话已启用工具", zap.Strings("tools", enabledNames))
	} else if len(enabledNames) > 0 {
		global.LRAG_LOG.Warn("对话配置了工具但未生效", zap.Bool("supportsToolCall", cfg.SupportsToolCall), zap.Strings("enabledNames", enabledNames))
	}
	if cfg.SupportsDeepThinking && req.UseDeepThinking {
		llmOpts = append(llmOpts, interfaces.WithReasoningEffort("high"))
	}

	var fullContent string
	for round := 0; round < maxToolRounds; round++ {
		resp, err := llmInst.GenerateContent(ctx, messages, llmOpts...)
		if err != nil {
			return nil, err
		}
		if len(resp.Choices) == 0 {
			break
		}
		choice := resp.Choices[0]

		if len(choice.ToolCalls) > 0 {
			toolNames := make([]string, len(choice.ToolCalls))
			for i, tc := range choice.ToolCalls {
				toolNames[i] = tc.Name
			}
			global.LRAG_LOG.Info("收到模型工具调用", zap.Int("count", len(choice.ToolCalls)), zap.Strings("names", toolNames))
			// 追加 assistant 消息（含 tool_calls）
			assistantMsg := interfaces.MessageContent{
				Role:      interfaces.MessageRoleAssistant,
				ToolCalls: choice.ToolCalls,
			}
			if tcText := strings.TrimSpace(llm.StripAssistantReasoningMarkers(choice.Content)); tcText != "" {
				assistantMsg.Parts = []interfaces.ContentPart{interfaces.TextContent{Text: tcText}}
			}
			messages = append(messages, assistantMsg)

			// 执行每个工具调用
			for _, tc := range choice.ToolCalls {
				if opts.OnToolCall != nil {
					opts.OnToolCall(tc.Name, "start", "")
				}
				toolCtx := tools.WithUserID(ctx, uid)
				result, execErr := tools.ExecuteTool(toolCtx, tc.Name, tc.Arguments)
				if execErr != nil {
					result = "Tool execution failed: " + execErr.Error()
				}
				if opts.OnToolCall != nil {
					opts.OnToolCall(tc.Name, "done", result)
				}
				toolResultMsg := interfaces.MessageContent{
					Role:       interfaces.MessageRoleTool,
					ToolCallID: tc.ID,
					ToolName:   tc.Name,
					Parts:      []interfaces.ContentPart{interfaces.TextContent{Text: result}},
				}
				messages = append(messages, toolResultMsg)
			}
			continue
		}

		fullContent = choice.Content
		break
	}

	userMsg := &rag.RagMessage{
		UUID:           uuid.New(),
		ConversationID: req.ConversationID,
		Role:           "user",
		Content:        req.Content,
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(userMsg).Error; err != nil {
		global.LRAG_LOG.Warn("failed to save user message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
	}
	streamRefsJSON := ""
	if len(refs) > 0 {
		if b, err := json.Marshal(refs); err == nil {
			streamRefsJSON = string(b)
		}
	}
	assistantMsg := &rag.RagMessage{
		UUID:           uuid.New(),
		ConversationID: req.ConversationID,
		Role:           "assistant",
		Content:        fullContent,
		References:     streamRefsJSON,
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(assistantMsg).Error; err != nil {
		global.LRAG_LOG.Warn("failed to save assistant message", zap.Uint("conversationId", req.ConversationID), zap.Error(err))
	}
	return &ChatStreamResult{
		Content:         fullContent,
		References:      apiRefs,
		OnlyNeedContext: false,
		OnlyNeedPrompt:  false,
		RetrievalMode:   effectiveMode,
		RetrievalQuery:  q,
		SearchQuery:     searchQ,
		HlKeywords:      KeywordsForAPIResponse(resolvedHl),
		LlKeywords:      KeywordsForAPIResponse(resolvedLl),
	}, nil
}

// List 对话列表（排除 Agent 编排产生的对话）
func (s *ConversationService) List(ctx context.Context, uid uint, req request.ConversationList) ([]rag.RagConversation, int64, error) {
	var list []rag.RagConversation
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagConversation{}).
		Where("user_id = ?", uid).
		Where("source_type != ? OR source_type = '' OR source_type IS NULL", "agent")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("updated_at DESC").Order("id DESC").Scopes(req.Paginate()).Find(&list).Error
	return list, total, err
}

// Get 获取对话
func (s *ConversationService) Get(ctx context.Context, uid uint, id uint) (*rag.RagConversation, error) {
	var conv rag.RagConversation
	err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, uid).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// Delete 删除对话
func (s *ConversationService) Delete(ctx context.Context, uid uint, id uint) error {
	return global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, uid).Delete(&rag.RagConversation{}).Error
}

// ListMessages 获取对话消息列表
func (s *ConversationService) ListMessages(ctx context.Context, uid uint, req request.ConversationMessageList) ([]rag.RagMessage, int64, error) {
	// 校验对话归属
	var conv rag.RagConversation
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", req.ConversationID, uid).First(&conv).Error; err != nil {
		return nil, 0, err
	}
	var list []rag.RagMessage
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagMessage{}).Where("conversation_id = ?", req.ConversationID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at ASC").Scopes(req.Paginate()).Find(&list).Error
	return list, total, err
}
