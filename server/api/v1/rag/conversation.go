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
	"github.com/LightningRAG/LightningRAG/server/rag/tools"
	ragservice "github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ConversationApi struct{}

// Create 创建对话
func (c *ConversationApi) Create(ctx *gin.Context) {
	var req request.ConversationCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	authorityId := utils.GetUserAuthorityId(ctx)
	conv, err := conversationService.Create(ctx.Request.Context(), uid, authorityId, req)
	if err != nil {
		global.LRAG_LOG.Error("创建对话失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(conv, ctx)
}

// Chat 对话（非流式，兼容旧客户端）
func (c *ConversationApi) Chat(ctx *gin.Context) {
	var req request.ConversationChat
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	msg, err := conversationService.Chat(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("对话失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(msg, ctx)
}

// QueryData 纯检索结构化数据（对齐 LightningRAG /query/data），不写入消息、不调用 LLM 生成回答
func (c *ConversationApi) QueryData(ctx *gin.Context) {
	var req request.ConversationQueryData
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(ctx, "common.not_logged_in"), ctx)
		return
	}
	out, err := conversationService.QueryData(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("对话 queryData 检索失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(out, ctx)
}

// ChatStream 流式对话，SSE 输出
func (c *ConversationApi) ChatStream(ctx *gin.Context) {
	var req request.ConversationChat
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	type toolCallEvent struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName,omitempty"`
		Status      string `json:"status"`
		Result      string `json:"result,omitempty"`
	}
	type streamEvent struct {
		content         string
		toolCall        *toolCallEvent
		done            bool
		refs            []map[string]any
		err             string
		onlyNeedContext bool
		onlyNeedPrompt  bool
		retrievalMode   string
		retrievalQuery  string
		searchQuery     string
		hlKeywords      []string
		llKeywords      []string
	}
	ch := make(chan streamEvent, 64)
	go func() {
		defer close(ch)
		callbacks := ragservice.ChatStreamCallbacks{
			OnChunk: func(delta string) {
				select {
				case ch <- streamEvent{content: delta}:
				case <-ctx.Request.Context().Done():
				}
			},
			OnToolCall: func(name, status, result string) {
				displayName := tools.GetToolDisplayName(name)
				select {
				case ch <- streamEvent{toolCall: &toolCallEvent{Name: name, DisplayName: displayName, Status: status, Result: result}}:
				case <-ctx.Request.Context().Done():
				}
			},
			OnReferences: func(refs []map[string]any) {
				select {
				case ch <- streamEvent{refs: refs}:
				case <-ctx.Request.Context().Done():
				}
			},
		}
		result, err := conversationService.ChatStream(ctx.Request.Context(), uid, req, callbacks)
		if err != nil {
			ch <- streamEvent{err: err.Error()}
			return
		}
		refs := []map[string]any(nil)
		onlyCtx := false
		onlyPrompt := false
		retrievalMode := ""
		retrievalQuery := ""
		searchQuery := ""
		var hlKw, llKw []string
		if result != nil {
			refs = result.References
			onlyCtx = result.OnlyNeedContext
			onlyPrompt = result.OnlyNeedPrompt
			retrievalMode = result.RetrievalMode
			retrievalQuery = result.RetrievalQuery
			searchQuery = result.SearchQuery
			hlKw = result.HlKeywords
			llKw = result.LlKeywords
		}
		ch <- streamEvent{done: true, refs: refs, onlyNeedContext: onlyCtx, onlyNeedPrompt: onlyPrompt, retrievalMode: retrievalMode, retrievalQuery: retrievalQuery, searchQuery: searchQuery, hlKeywords: hlKw, llKeywords: llKw}
	}()

	ctx.Stream(func(w io.Writer) bool {
		select {
		case ev, ok := <-ch:
			if !ok {
				writeSSE(ctx, w, map[string]any{"done": true})
				return false
			}
			if ev.err != "" {
				writeSSE(ctx, w, map[string]any{"error": ev.err})
				return false
			}
			if ev.done {
				payload := map[string]any{"done": true}
				if len(ev.refs) > 0 {
					payload["references"] = ev.refs
				}
				if ev.onlyNeedContext {
					payload["onlyNeedContext"] = true
				}
				if ev.onlyNeedPrompt {
					payload["onlyNeedPrompt"] = true
				}
				if ev.retrievalMode != "" {
					payload["retrievalMode"] = ev.retrievalMode
				}
				if ev.retrievalQuery != "" {
					payload["retrievalQuery"] = ev.retrievalQuery
				}
				if ev.searchQuery != "" {
					payload["searchQuery"] = ev.searchQuery
				}
				hl := ev.hlKeywords
				if hl == nil {
					hl = []string{}
				}
				ll := ev.llKeywords
				if ll == nil {
					ll = []string{}
				}
				payload["hlKeywords"] = hl
				payload["llKeywords"] = ll
				writeSSE(ctx, w, payload)
				return false
			}
			if ev.content != "" {
				writeSSE(ctx, w, map[string]any{"content": ev.content})
			}
			if ev.toolCall != nil {
				writeSSE(ctx, w, map[string]any{"toolCall": ev.toolCall})
			}
			if len(ev.refs) > 0 && !ev.done {
				writeSSE(ctx, w, map[string]any{"references": ev.refs})
			}
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
}

func writeSSE(c *gin.Context, w io.Writer, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", b)
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}
}

// List 对话列表
func (c *ConversationApi) List(ctx *gin.Context) {
	var req request.ConversationList
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	list, total, err := conversationService.List(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取对话列表失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(ctx, "common.fetch_success"), ctx)
}

// Update 更新对话（如修改启用的工具）
func (c *ConversationApi) Update(ctx *gin.Context) {
	var req request.ConversationUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	if err := conversationService.Update(ctx.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("更新对话失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithMessage(i18n.Msg(ctx, "common.update_success"), ctx)
}

// Get 获取对话详情
func (c *ConversationApi) Get(ctx *gin.Context) {
	var req request.ConversationGet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	conv, err := conversationService.Get(ctx.Request.Context(), uid, req.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取对话失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithData(conv, ctx)
}

// Delete 删除对话
func (c *ConversationApi) Delete(ctx *gin.Context) {
	var req request.ConversationDelete
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	uid := utils.GetUserID(ctx)
	if err := conversationService.Delete(ctx.Request.Context(), uid, req.ID); err != nil {
		global.LRAG_LOG.Error("删除对话失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithMessage(i18n.Msg(ctx, "common.delete_success"), ctx)
}

// ListTools 获取对话可用的工具列表（供前端展示及后续扩展）
func (c *ConversationApi) ListTools(ctx *gin.Context) {
	list := tools.ListToolMeta()
	response.OkWithData(list, ctx)
}

// ListMessages 获取对话消息列表
func (c *ConversationApi) ListMessages(ctx *gin.Context) {
	var req request.ConversationMessageList
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 100
	}
	uid := utils.GetUserID(ctx)
	list, total, err := conversationService.ListMessages(ctx.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取消息列表失败", zap.Error(err))
		response.FailWithError(ctx, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(ctx, "common.fetch_success"), ctx)
}
