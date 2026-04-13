package rag

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	ragmodel "github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	ragservice "github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ChannelConnectorApi struct{}

func connectorWebhookURL(connectorID uint) string {
	prefix := global.LRAG_CONFIG.System.RouterPrefix
	return prefix + "/open/channel/webhook/" + strconv.FormatUint(uint64(connectorID), 10)
}

func toConnectorMap(conn *ragmodel.RagChannelConnector) gin.H {
	if conn == nil {
		return gin.H{}
	}
	return gin.H{
		"id":               conn.ID,
		"ownerId":          conn.OwnerID,
		"authorityId":      conn.AuthorityID,
		"name":             conn.Name,
		"channel":          conn.Channel,
		"agentId":          conn.AgentID,
		"enabled":          conn.Enabled,
		"extra":            conn.Extra,
		"webhookUrl":       connectorWebhookURL(conn.ID),
		"webhookSecretSet": conn.WebhookSecret != "",
		"createdAt":        conn.CreatedAt,
		"updatedAt":        conn.UpdatedAt,
	}
}

// CreateChannelConnector 创建第三方渠道连接器
func (a *ChannelConnectorApi) CreateChannelConnector(c *gin.Context) {
	var req request.ChannelConnectorCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	authID := utils.GetUserAuthorityId(c)
	conn, secretOnce, err := channelConnectorService.CreateChannelConnector(c.Request.Context(), uid, authID, req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	data := toConnectorMap(conn)
	if secretOnce != "" {
		data["webhookSecret"] = secretOnce
	}
	response.OkWithData(data, c)
}

func (a *ChannelConnectorApi) UpdateChannelConnector(c *gin.Context) {
	var req request.ChannelConnectorUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := channelConnectorService.UpdateChannelConnector(c.Request.Context(), uid, req); err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

func (a *ChannelConnectorApi) ListChannelConnectors(c *gin.Context) {
	var req request.ChannelConnectorList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	list, total, err := channelConnectorService.ListChannelConnectors(c.Request.Context(), uid, req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	rows := make([]gin.H, 0, len(list))
	for i := range list {
		rows = append(rows, toConnectorMap(&list[i]))
	}
	response.OkWithDetailed(response.PageResult{
		List: rows, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

func (a *ChannelConnectorApi) GetChannelConnector(c *gin.Context) {
	var req request.ChannelConnectorGet
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	conn, err := channelConnectorService.GetChannelConnector(c.Request.Context(), uid, req.ID)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(toConnectorMap(conn), c)
}

// ListChannelConnectorTypes 已注册的第三方渠道类型（与 server/channel 各适配器 init Register 一致）
func (a *ChannelConnectorApi) ListChannelConnectorTypes(c *gin.Context) {
	response.OkWithData(channel.RegisteredKinds(), c)
}

func (a *ChannelConnectorApi) DeleteChannelConnector(c *gin.Context) {
	var req request.ChannelConnectorDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := channelConnectorService.DeleteChannelConnector(c.Request.Context(), uid, req.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// ListChannelOutboundQueue 出站重试队列（SendReply / Discord 延迟编辑失败入队）
func (a *ChannelConnectorApi) ListChannelOutboundQueue(c *gin.Context) {
	var req request.ChannelOutboundList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	} else if req.PageSize > 100 {
		req.PageSize = 100
	}
	uid := utils.GetUserID(c)
	list, total, err := channelConnectorService.ListChannelOutboundQueue(c.Request.Context(), uid, req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// DeleteChannelOutboundRow 删除一条出站重试任务
func (a *ChannelConnectorApi) DeleteChannelOutboundRow(c *gin.Context) {
	var req request.ChannelOutboundDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := channelConnectorService.DeleteChannelOutboundRow(c.Request.Context(), uid, req.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// RunChannelOutboundOnce 立即执行一轮出站重试（与定时任务相同逻辑；适合 channel-outbound-poll-seconds=-1 时手动触发）
func (a *ChannelConnectorApi) RunChannelOutboundOnce(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	n, err := channelConnectorService.ProcessChannelOutboundQueue(ctx, 0)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(gin.H{"processed": n}, c)
}

// OpenChannelWebhook 公开 Webhook（鉴权：X-Webhook-Secret）
func (a *ChannelConnectorApi) OpenChannelWebhook(c *gin.Context) {
	idStr := c.Param("connectorId")
	cid64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || cid64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connector id"})
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
		return
	}
	secret := c.GetHeader("X-Webhook-Secret")
	meta := channel.WebhookHTTPMeta{Method: c.Request.Method, Query: c.Request.URL.Query(), Headers: c.Request.Header}
	out, err := channelConnectorService.ProcessWebhook(c.Request.Context(), uint(cid64), secret, raw, meta)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "connector not found"})
		case errors.Is(err, ragservice.ErrWebhookSecretMismatch):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case errors.Is(err, ragservice.ErrConnectorDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "connector disabled"})
		default:
			global.LRAG_LOG.Warn("OpenChannelWebhook failed", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		}
		return
	}
	if len(out.ImmediateBody) > 0 {
		ct := out.ImmediateContentType
		if ct == "" {
			ct = "application/json; charset=utf-8"
		}
		c.Data(out.ImmediateStatus, ct, out.ImmediateBody)
		if out.DeferredAfter != nil {
			fn := out.DeferredAfter
			go func() {
				defer func() {
					if r := recover(); r != nil {
						global.LRAG_LOG.Error("channel webhook deferred panic", zap.Any("recover", r))
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
				defer cancel()
				if err := fn(ctx); err != nil {
					global.LRAG_LOG.Warn("channel webhook deferred", zap.Error(err))
				}
			}()
		}
		return
	}
	if len(out.FinalBody) > 0 {
		ct := out.FinalContentType
		if ct == "" {
			ct = "application/xml; charset=utf-8"
		}
		c.Data(http.StatusOK, ct, out.FinalBody)
		return
	}
	c.JSON(http.StatusOK, out.JSONResponse)
}

// OpenChannelWebhookGet 微信公众号服务器配置 URL 验证（GET + echostr）
func (a *ChannelConnectorApi) OpenChannelWebhookGet(c *gin.Context) {
	idStr := c.Param("connectorId")
	cid64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || cid64 == 0 {
		c.String(http.StatusBadRequest, "invalid connector id")
		return
	}
	out, err := channelConnectorService.ProcessOpenChannelWebhookGet(c.Request.Context(), uint(cid64), c.Request.URL.Query())
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.String(http.StatusNotFound, "not found")
		case errors.Is(err, ragservice.ErrWebhookSecretMismatch):
			c.String(http.StatusUnauthorized, "unauthorized")
		case errors.Is(err, ragservice.ErrConnectorDisabled):
			c.String(http.StatusForbidden, "disabled")
		case errors.Is(err, ragservice.ErrChannelGETVerifyNotSupported):
			c.String(http.StatusBadRequest, "GET verify not supported for this connector")
		default:
			global.LRAG_LOG.Warn("OpenChannelWebhookGet failed", zap.Error(err))
			c.String(http.StatusBadRequest, "bad request")
		}
		return
	}
	ct := out.ImmediateContentType
	if ct == "" {
		ct = "text/plain; charset=utf-8"
	}
	c.Data(out.ImmediateStatus, ct, out.ImmediateBody)
}
