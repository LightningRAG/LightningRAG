package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	sysReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ApiTokenApi struct{}

// CreateApiToken 签发Token
func (s *ApiTokenApi) CreateApiToken(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"userId"`
		AuthorityID uint   `json:"authorityId"`
		Days        int    `json:"days"` // -1为永久, 其他为天数
		Remark      string `json:"remark"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}

	token := system.SysApiToken{
		UserID:      req.UserID,
		AuthorityID: req.AuthorityID,
		Remark:      req.Remark,
	}

	jwtStr, err := apiTokenService.CreateApiToken(token, req.Days)
	if err != nil {
		global.LRAG_LOG.Error("签发失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "sys.api_token.issue_failed", err.Error()), c)
		return
	}

	response.OkWithDetailed(gin.H{"token": jwtStr}, i18n.Msg(c, "sys.api_token.issue_success"), c)
}

// GetApiTokenList 获取列表
func (s *ApiTokenApi) GetApiTokenList(c *gin.Context) {
	var pageInfo sysReq.SysApiTokenSearch
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := apiTokenService.GetApiTokenList(pageInfo)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// DeleteApiToken 作废Token
func (s *ApiTokenApi) DeleteApiToken(c *gin.Context) {
	var req system.SysApiToken
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiTokenService.DeleteApiToken(req.ID)
	if err != nil {
		global.LRAG_LOG.Error("作废失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "sys.api_token.revoke_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "sys.api_token.revoke_success"), c)
}
