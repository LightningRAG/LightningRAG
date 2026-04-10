package oauthapp

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	svc "github.com/LightningRAG/LightningRAG/server/service/oauthapp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OAuthSettingApi struct{}

func (a *OAuthSettingApi) Get(c *gin.Context) {
	row, err := svc.SysOAuthSettingServiceApp.GetAdmin()
	if err != nil {
		global.LRAG_LOG.Warn("oauth setting get", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(row, i18n.Msg(c, "common.fetch_success"), c)
}

func (a *OAuthSettingApi) Update(c *gin.Context) {
	var req systemReq.SysOAuthSettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := svc.SysOAuthSettingServiceApp.Update(req); err != nil {
		global.LRAG_LOG.Warn("oauth setting update", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.update_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}
