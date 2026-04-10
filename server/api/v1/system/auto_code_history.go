package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	common "github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	request "github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AutoCodeHistoryApi struct{}

// First
// @Tags      AutoCode
// @Summary   获取meta信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                                            true  "请求参数"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "获取meta信息"
// @Router    /autoCode/getMeta [post]
func (a *AutoCodeHistoryApi) First(c *gin.Context) {
	var info common.GetById
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	data, err := autoCodeHistoryService.First(c.Request.Context(), info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(gin.H{"meta": data}, i18n.Msg(c, "common.fetch_success"), c)
}

// Delete
// @Tags      AutoCode
// @Summary   删除回滚记录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                true  "请求参数"
// @Success   200   {object}  response.Response{msg=string}  "删除回滚记录"
// @Router    /autoCode/delSysHistory [post]
func (a *AutoCodeHistoryApi) Delete(c *gin.Context) {
	var info common.GetById
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = autoCodeHistoryService.Delete(c.Request.Context(), info)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// RollBack
// @Tags      AutoCode
// @Summary   回滚自动生成代码
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.SysAutoHistoryRollBack             true  "请求参数"
// @Success   200   {object}  response.Response{msg=string}  "回滚自动生成代码"
// @Router    /autoCode/rollback [post]
func (a *AutoCodeHistoryApi) RollBack(c *gin.Context) {
	var info request.SysAutoHistoryRollBack
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = autoCodeHistoryService.RollBack(c.Request.Context(), info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.rollback_success"), c)
}

// GetList
// @Tags      AutoCode
// @Summary   查询回滚记录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      common.PageInfo                                true  "请求参数"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "查询回滚记录,返回包括列表,总数,页码,每页数量"
// @Router    /autoCode/getSysHistory [post]
func (a *AutoCodeHistoryApi) GetList(c *gin.Context) {
	var info common.PageInfo
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := autoCodeHistoryService.GetList(c.Request.Context(), info)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     info.Page,
		PageSize: info.PageSize,
	}, "获取成功", c)
}
