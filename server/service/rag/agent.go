package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/agent/canvas"
	"github.com/LightningRAG/LightningRAG/server/agent/component"
	"github.com/LightningRAG/LightningRAG/server/agent/dsl"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentService struct{}

// agentRetrieverFactory 为 Agent 提供知识库检索器（finalTopN 与画布 Retrieval 节点的 top_n 一致，默认候选池按该值扩展）
func agentRetrieverFactory(ctx context.Context, kbIDs []uint, userID uint, finalTopN int, opts component.RetrieverFactoryOptions) (component.Retriever, error) {
	n := ClampRetrieveTopN(finalTopN, DefaultConversationChunkTopKFromConfig())
	session := RetrieverSessionOptions{}
	ApplyDefaultConversationRetrievePoolIfNeeded(&session, n)
	ApplyPageIndexTocEnhanceFromRequest(&session, opts.PageIndexTocEnhance)
	r, err := CreateRetrieverForKnowledgeBases(ctx, kbIDs, userID, n, session)
	if err != nil || r == nil {
		return nil, err
	}
	return &component.RAGRetrieverAdapter{R: r}, nil
}

// agentLLMConfigResolver 为 Agent 组件提供 LLM 配置解析（含 API Key），
// 按用户模型 → 管理员模型的顺序匹配 provider + modelName
func agentLLMConfigResolver(ctx context.Context, userID uint, provider, modelName string) (*component.LLMResolvedConfig, error) {
	var userLLMs []rag.RagUserLLM
	if err := global.LRAG_DB.WithContext(ctx).
		Where("user_id = ? AND provider = ? AND model_name = ? AND enabled = ?", userID, provider, modelName, true).
		Find(&userLLMs).Error; err == nil && len(userLLMs) > 0 {
		m := userLLMs[0]
		return &component.LLMResolvedConfig{
			Provider: m.Provider, ModelName: m.ModelName,
			BaseURL: m.BaseURL, APIKey: m.APIKey,
		}, nil
	}

	var adminLLMs []rag.RagLLMProvider
	if err := global.LRAG_DB.WithContext(ctx).
		Where("name = ? AND model_name = ? AND enabled = ?", provider, modelName, true).
		Find(&adminLLMs).Error; err == nil && len(adminLLMs) > 0 {
		m := adminLLMs[0]
		return &component.LLMResolvedConfig{
			Provider: m.Name, ModelName: m.ModelName,
			BaseURL: m.BaseURL, APIKey: m.APIKey,
		}, nil
	}

	return nil, fmt.Errorf("未找到匹配的 LLM 配置: %s/%s", provider, modelName)
}

// Run 运行 Agent 工作流（支持 agent_id 或 dsl）
func (s *AgentService) Run(ctx context.Context, uid uint, authorityID uint, agentID uint, dslJSON map[string]any, query string, conversationID uint, workflowGlobals map[string]any) (map[string]any, uint, error) {
	d, err := s.resolveDSL(ctx, uid, agentID, dslJSON)
	if err != nil {
		return nil, 0, err
	}
	history, convID, err := s.resolveConversation(ctx, uid, authorityID, agentID, conversationID)
	if err != nil {
		return nil, 0, err
	}
	c, err := canvas.New(d, uid, agentRetrieverFactory, agentLLMConfigResolver)
	if err != nil {
		return nil, 0, err
	}
	out, err := c.Run(ctx, canvas.RunInput{
		Query:           query,
		Files:           nil,
		UserID:          uid,
		History:         history,
		ConversationID:  convID,
		WorkflowGlobals: workflowGlobals,
	})
	if err != nil {
		return nil, 0, err
	}
	content := extractContent(out)
	if convID > 0 && content != "" {
		s.saveAgentMessages(ctx, convID, query, content)
	}
	return out, convID, nil
}

// RunStream 流式运行 Agent，支持多轮上下文
func (s *AgentService) RunStream(ctx context.Context, uid uint, authorityID uint, req request.AgentRun, onChunk func(string)) (*AgentRunStreamResult, error) {
	d, err := s.resolveDSL(ctx, uid, req.AgentID, req.DSL)
	if err != nil {
		return nil, err
	}
	history, convID, err := s.resolveConversation(ctx, uid, authorityID, req.AgentID, req.ConversationID)
	if err != nil {
		return nil, err
	}
	c, err := canvas.New(d, uid, agentRetrieverFactory, agentLLMConfigResolver)
	if err != nil {
		return nil, err
	}
	out, err := c.Run(ctx, canvas.RunInput{
		Query:           req.Query,
		Files:           nil,
		UserID:          uid,
		OnChunk:         onChunk,
		History:         history,
		ConversationID:  convID,
		WorkflowGlobals: req.WorkflowGlobals,
	})
	if err != nil {
		return nil, err
	}
	content := extractContent(out)
	if convID > 0 && content != "" {
		s.saveAgentMessages(ctx, convID, req.Query, content)
	}
	paused := false
	if out != nil {
		if v, ok := out["workflowPausedAtEntry"].(bool); ok {
			paused = v
		}
	}
	return &AgentRunStreamResult{Content: content, ConversationID: convID, Output: out, WorkflowPausedAtEntry: paused}, nil
}

// AgentRunStreamResult 流式运行结果
type AgentRunStreamResult struct {
	Content               string
	ConversationID        uint
	Output                map[string]any
	WorkflowPausedAtEntry bool
}

func (s *AgentService) resolveDSL(ctx context.Context, uid uint, agentID uint, dslJSON map[string]any) (*dsl.DSL, error) {
	var d dsl.DSL
	if agentID > 0 {
		var agent rag.RagAgent
		if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", agentID, uid).First(&agent).Error; err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(agent.DSL), &d); err != nil {
			return nil, err
		}
	} else if dslJSON != nil {
		data, err := json.Marshal(dslJSON)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &d); err != nil {
			return nil, err
		}
	} else {
		return nil, gorm.ErrRecordNotFound
	}
	return &d, nil
}

func (s *AgentService) resolveConversation(ctx context.Context, uid uint, authorityID uint, agentID uint, conversationID uint) ([]component.HistoryMessage, uint, error) {
	if conversationID > 0 {
		var conv rag.RagConversation
		if err := global.LRAG_DB.Where("id = ? AND user_id = ?", conversationID, uid).First(&conv).Error; err != nil {
			return nil, 0, err
		}
		history, _ := loadConversationHistory(ctx, conversationID, EffectiveConversationHistoryLimit())
		return ragMessagesToHistory(history), conversationID, nil
	}
	if agentID > 0 {
		conv, err := s.createAgentConversation(ctx, uid, authorityID, agentID)
		if err != nil {
			return nil, 0, err
		}
		return nil, conv.ID, nil
	}
	return nil, 0, nil
}

func (s *AgentService) createAgentConversation(ctx context.Context, uid uint, authorityID uint, agentID uint) (*rag.RagConversation, error) {
	llmProviderID, llmSource := uint(0), "user"
	if id, src, _, ok := llmProviderService.ResolveDefaultLLM(ctx, uid, authorityID, "chat"); ok {
		llmProviderID, llmSource = id, src
	}
	conv := &rag.RagConversation{
		UUID:          uuid.New(),
		UserID:        uid,
		Title:         "Agent 对话",
		LLMProviderID: llmProviderID,
		LLMSource:     llmSource,
		SourceType:    "agent",
		SourceIDs:     strconv.FormatUint(uint64(agentID), 10),
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

func ragMessagesToHistory(msgs []rag.RagMessage) []component.HistoryMessage {
	out := make([]component.HistoryMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, component.HistoryMessage{Role: m.Role, Content: m.Content})
	}
	return out
}

func extractContent(out map[string]any) string {
	if out == nil {
		return ""
	}
	if v, ok := out["content"].(string); ok {
		return v
	}
	if arr, ok := out["content"].([]any); ok && len(arr) > 0 {
		var sb strings.Builder
		for i, a := range arr {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(strings.TrimSpace(strings.TrimPrefix(fmt.Sprint(a), " ")))
		}
		return sb.String()
	}
	return ""
}

func (s *AgentService) saveAgentMessages(ctx context.Context, conversationID uint, userContent, assistantContent string) {
	_ = global.LRAG_DB.WithContext(ctx).Create(&rag.RagMessage{
		UUID:           uuid.New(),
		ConversationID: conversationID,
		Role:           "user",
		Content:        userContent,
	}).Error
	_ = global.LRAG_DB.WithContext(ctx).Create(&rag.RagMessage{
		UUID:           uuid.New(),
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        assistantContent,
	}).Error
}

// Create 创建 Agent
func (s *AgentService) Create(ctx context.Context, uid uint, req request.AgentCreate) (*rag.RagAgent, error) {
	dslJSON, err := json.Marshal(req.DSL)
	if err != nil {
		return nil, err
	}
	agent := &rag.RagAgent{
		OwnerID: uid,
		Name:    req.Name,
		Desc:    req.Desc,
		DSL:     string(dslJSON),
	}
	if err := global.LRAG_DB.Create(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

// List 列表
func (s *AgentService) List(ctx context.Context, uid uint, req request.AgentList) (list []rag.RagAgent, total int64, err error) {
	db := global.LRAG_DB.Model(&rag.RagAgent{}).Where("owner_id = ?", uid)
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	if err = db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Get 获取单个
func (s *AgentService) Get(ctx context.Context, uid uint, id uint) (*rag.RagAgent, error) {
	var agent rag.RagAgent
	if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", id, uid).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

// Update 更新（支持部分更新：仅传 dsl 时只更新 dsl，传 name 时同时更新 name/desc）
func (s *AgentService) Update(ctx context.Context, uid uint, req request.AgentUpdate) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
		updates["desc"] = req.Desc
	}
	if req.DSL != nil {
		dslJSON, err := json.Marshal(req.DSL)
		if err != nil {
			return err
		}
		updates["dsl"] = string(dslJSON)
	}
	if len(updates) == 0 {
		return nil
	}
	return global.LRAG_DB.Model(&rag.RagAgent{}).Where("id = ? AND owner_id = ?", req.ID, uid).Updates(updates).Error
}

// Delete 删除
func (s *AgentService) Delete(ctx context.Context, uid uint, id uint) error {
	return global.LRAG_DB.Where("id = ? AND owner_id = ?", id, uid).Delete(&rag.RagAgent{}).Error
}

// CreateFromTemplate 从模板创建
func (s *AgentService) CreateFromTemplate(ctx context.Context, uid uint, req request.AgentCreateFromTemplate) (*rag.RagAgent, error) {
	d, err := s.LoadTemplate(req.TemplateName)
	if err != nil {
		return nil, err
	}
	dslJSON, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	agent := &rag.RagAgent{
		OwnerID: uid,
		Name:    req.Name,
		Desc:    req.Desc,
		DSL:     string(dslJSON),
	}
	if err := global.LRAG_DB.Create(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

// LoadTemplateDSL 加载模板 DSL（供 API 返回）
func (s *AgentService) LoadTemplateDSL(ctx context.Context, name string) (map[string]any, error) {
	d, err := s.LoadTemplate(name)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	var dslMap map[string]any
	if err := json.Unmarshal(data, &dslMap); err != nil {
		return nil, err
	}
	return dslMap, nil
}

// ListTemplates 列出内置模板（从工作目录或 server/agent/templates 读取）
func (s *AgentService) ListTemplates() ([]map[string]any, error) {
	dirs := []string{
		filepath.Join("server", "agent", "templates"),
		filepath.Join("agent", "templates"),
		"templates",
	}
	var entries []os.DirEntry
	var dir string
	for _, d := range dirs {
		if e, err := os.ReadDir(d); err == nil {
			entries = e
			dir = d
			break
		}
	}
	if entries == nil {
		return []map[string]any{}, nil
	}
	var list []map[string]any
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var t map[string]any
		if err := json.Unmarshal(raw, &t); err != nil {
			continue
		}
		templateName := strings.TrimSuffix(e.Name(), ".json")
		if _, ok := t["templateName"]; !ok {
			t["templateName"] = templateName
		}
		if _, hasTitle := t["title"]; !hasTitle {
			t["title"] = templateName
		}
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool {
		ci, _ := list[i]["category"].(string)
		cj, _ := list[j]["category"].(string)
		if ci != cj {
			if ci == "" {
				return false
			}
			if cj == "" {
				return true
			}
			return strings.Compare(ci, cj) < 0
		}
		ti, _ := list[i]["title"].(string)
		tj, _ := list[j]["title"].(string)
		return strings.Compare(ti, tj) < 0
	})
	return list, nil
}

// LoadTemplate 加载模板（支持 {title,description,dsl} 或纯 DSL 格式）
func (s *AgentService) LoadTemplate(name string) (*dsl.DSL, error) {
	if filepath.Ext(name) == "" {
		name = name + ".json"
	}
	dirs := []string{
		filepath.Join("server", "agent", "templates"),
		filepath.Join("agent", "templates"),
	}
	for _, d := range dirs {
		raw, err := os.ReadFile(filepath.Join(d, name))
		if err == nil {
			var wrapper map[string]any
			if err := json.Unmarshal(raw, &wrapper); err != nil {
				return nil, err
			}
			if dslVal, ok := wrapper["dsl"]; ok {
				data, err := json.Marshal(dslVal)
				if err != nil {
					return nil, err
				}
				var d dsl.DSL
				if err := json.Unmarshal(data, &d); err != nil {
					return nil, err
				}
				return &d, nil
			}
			var d dsl.DSL
			if err := json.Unmarshal(raw, &d); err != nil {
				return nil, err
			}
			return &d, nil
		}
	}
	return nil, os.ErrNotExist
}
