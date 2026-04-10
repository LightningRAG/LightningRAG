package rag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AgentApi struct{}

// Run 运行 Agent 工作流（支持 agent_id 或 dsl，支持多轮对话）
func (a *AgentApi) Run(ctx *gin.Context) {
	var req request.AgentRun
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	authorityID := utils.GetUserAuthorityId(ctx)
	out, convID, err := agentService.Run(ctx.Request.Context(), uid, authorityID, req.AgentID, req.DSL, req.Query, req.ConversationID, req.WorkflowGlobals)
	if err != nil {
		global.LRAG_LOG.Error("Agent 运行失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	result := map[string]any{}
	for k, v := range out {
		result[k] = v
	}
	if convID > 0 {
		result["conversationId"] = convID
	}
	response.OkWithData(result, ctx)
}

// RunStream 流式运行 Agent，SSE 输出，支持多轮对话
func (a *AgentApi) RunStream(ctx *gin.Context) {
	var req request.AgentRun
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	authorityID := utils.GetUserAuthorityId(ctx)

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	type streamEv struct {
		content               string
		done                  bool
		convID                uint
		err                   string
		workflowPausedAtEntry bool
	}
	ch := make(chan streamEv, 64)
	go func() {
		defer close(ch)
		result, err := agentService.RunStream(ctx.Request.Context(), uid, authorityID, req, func(delta string) {
			select {
			case ch <- streamEv{content: delta}:
			case <-ctx.Request.Context().Done():
			}
		})
		if err != nil {
			ch <- streamEv{err: err.Error()}
			return
		}
		convID := uint(0)
		if result != nil {
			convID = result.ConversationID
		}
		paused := false
		if result != nil && result.WorkflowPausedAtEntry {
			paused = true
		}
		ch <- streamEv{done: true, convID: convID, workflowPausedAtEntry: paused}
	}()

	ctx.Stream(func(w io.Writer) bool {
		select {
		case ev, ok := <-ch:
			if !ok {
				writeAgentSSE(ctx, w, map[string]any{"done": true})
				return false
			}
			if ev.err != "" {
				writeAgentSSE(ctx, w, map[string]any{"error": ev.err})
				return false
			}
			if ev.done {
				payload := map[string]any{"done": true}
				if ev.convID > 0 {
					payload["conversationId"] = ev.convID
				}
				if ev.workflowPausedAtEntry {
					payload["workflowPausedAtEntry"] = true
				}
				writeAgentSSE(ctx, w, payload)
				return false
			}
			if ev.content != "" {
				writeAgentSSE(ctx, w, map[string]any{"content": ev.content})
			}
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
}

func writeAgentSSE(c *gin.Context, w io.Writer, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", b)
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}
}

// ListTemplates 列出可导入的模板（POST 保持与现有 RAG 风格一致）
func (a *AgentApi) ListTemplates(ctx *gin.Context) {
	list, err := agentService.ListTemplates()
	if err != nil {
		global.LRAG_LOG.Error("获取模板列表失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(list, ctx)
}

// LoadTemplate 加载模板 DSL
func (a *AgentApi) LoadTemplate(ctx *gin.Context) {
	var req request.AgentLoadTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	dsl, err := agentService.LoadTemplateDSL(ctx.Request.Context(), req.TemplateName)
	if err != nil {
		global.LRAG_LOG.Error("加载模板失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(dsl, ctx)
}

// Create 创建 Agent
func (a *AgentApi) Create(ctx *gin.Context) {
	var req request.AgentCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	agent, err := agentService.Create(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("创建 Agent 失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(agent, ctx)
}

// List 列表
func (a *AgentApi) List(ctx *gin.Context) {
	var req request.AgentList
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	list, total, err := agentService.List(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取 Agent 列表失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total}, i18n.Msg(ctx, "common.fetch_success"), ctx)
}

// Get 获取单个
func (a *AgentApi) Get(ctx *gin.Context) {
	var req request.AgentGet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	agent, err := agentService.Get(ctx.Request.Context(), uid, req.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取 Agent 失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(agent, ctx)
}

// Update 更新
func (a *AgentApi) Update(ctx *gin.Context) {
	var req request.AgentUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	if err := agentService.Update(ctx.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("更新 Agent 失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithMessage(i18n.Msg(ctx, "common.update_success"), ctx)
}

// Delete 删除
func (a *AgentApi) Delete(ctx *gin.Context) {
	var req request.AgentDelete
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	if err := agentService.Delete(ctx.Request.Context(), uid, req.ID); err != nil {
		global.LRAG_LOG.Error("删除 Agent 失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithMessage(i18n.Msg(ctx, "common.delete_success"), ctx)
}

// CreateFromTemplate 从模板创建
func (a *AgentApi) CreateFromTemplate(ctx *gin.Context) {
	var req request.AgentCreateFromTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	agent, err := agentService.CreateFromTemplate(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("从模板创建 Agent 失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(agent, ctx)
}
