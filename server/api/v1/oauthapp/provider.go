package oauthapp

import (
	"strconv"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	svc "github.com/LightningRAG/LightningRAG/server/service/oauthapp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OAuthProviderApi struct{}

func (a *OAuthProviderApi) Create(c *gin.Context) {
	var req systemReq.SysOAuthProviderCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := svc.SysOAuthProviderServiceApp.Create(req); err != nil {
		global.LRAG_LOG.Warn("oauth provider create", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.create_success"), c)
}

func (a *OAuthProviderApi) Update(c *gin.Context) {
	var req systemReq.SysOAuthProviderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := svc.SysOAuthProviderServiceApp.Update(req); err != nil {
		global.LRAG_LOG.Warn("oauth provider update", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.update_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

func (a *OAuthProviderApi) Delete(c *gin.Context) {
	idStr := c.Query("ID")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id64 == 0 {
		response.FailWithMessage(i18n.Msg(c, "validation.param_invalid"), c)
		return
	}
	if err := svc.SysOAuthProviderServiceApp.Delete(uint(id64)); err != nil {
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

func (a *OAuthProviderApi) DeleteByIds(c *gin.Context) {
	var ids request.IdsReq
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := svc.SysOAuthProviderServiceApp.DeleteByIds(ids); err != nil {
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

func (a *OAuthProviderApi) Find(c *gin.Context) {
	idStr := c.Query("ID")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id64 == 0 {
		response.FailWithMessage(i18n.Msg(c, "validation.param_invalid"), c)
		return
	}
	row, err := svc.SysOAuthProviderServiceApp.Find(uint(id64))
	if err != nil {
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(row, i18n.Msg(c, "common.fetch_success"), c)
}

func (a *OAuthProviderApi) List(c *gin.Context) {
	var pageInfo systemReq.SysOAuthProviderSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithError(c, err)
		return
	}
	if pageInfo.PageSize == 0 {
		pageInfo.PageSize = 10
	}
	if pageInfo.Page == 0 {
		pageInfo.Page = 1
	}
	list, total, err := svc.SysOAuthProviderServiceApp.List(pageInfo)
	if err != nil {
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

func (a *OAuthProviderApi) RegisteredKinds(c *gin.Context) {
	list := svc.SysOAuthProviderServiceApp.RegisteredKindsForAdmin()
	response.OkWithDetailed(list, i18n.Msg(c, "common.fetch_success"), c)
}
