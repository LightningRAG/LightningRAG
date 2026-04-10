package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginLogApi struct{}

func (s *LoginLogApi) DeleteLoginLog(c *gin.Context) {
	var loginLog system.SysLoginLog
	err := c.ShouldBindJSON(&loginLog)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = loginLogService.DeleteLoginLog(loginLog)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

func (s *LoginLogApi) DeleteLoginLogByIds(c *gin.Context) {
	var SDS request.IdsReq
	err := c.ShouldBindJSON(&SDS)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = loginLogService.DeleteLoginLogByIds(SDS)
	if err != nil {
		global.LRAG_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.batch_delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.batch_delete_success"), c)
}

func (s *LoginLogApi) FindLoginLog(c *gin.Context) {
	var loginLog system.SysLoginLog
	err := c.ShouldBindQuery(&loginLog)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	reLoginLog, err := loginLogService.GetLoginLog(loginLog.ID)
	if err != nil {
		global.LRAG_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.query_failed"), c)
		return
	}
	response.OkWithDetailed(reLoginLog, i18n.Msg(c, "common.query_success"), c)
}

func (s *LoginLogApi) GetLoginLogList(c *gin.Context) {
	var pageInfo systemReq.SysLoginLogSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := loginLogService.GetLoginLogInfoList(pageInfo)
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
