package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthorityBtnApi struct{}

// GetAuthorityBtn
// @Tags      AuthorityBtn
// @Summary   获取权限按钮
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.SysAuthorityBtnReq                                      true  "菜单id, 角色id, 选中的按钮id"
// @Success   200   {object}  response.Response{data=response.SysAuthorityBtnRes,msg=string}  "返回列表成功"
// @Router    /authorityBtn/getAuthorityBtn [post]
func (a *AuthorityBtnApi) GetAuthorityBtn(c *gin.Context) {
	var req request.SysAuthorityBtnReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	res, err := authorityBtnService.GetAuthorityBtn(req)
	if err != nil {
		global.LRAG_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.query_failed"), c)
		return
	}
	response.OkWithDetailed(res, i18n.Msg(c, "common.query_success"), c)
}

// SetAuthorityBtn
// @Tags      AuthorityBtn
// @Summary   设置权限按钮
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.SysAuthorityBtnReq     true  "菜单id, 角色id, 选中的按钮id"
// @Success   200   {object}  response.Response{msg=string}  "返回列表成功"
// @Router    /authorityBtn/setAuthorityBtn [post]
func (a *AuthorityBtnApi) SetAuthorityBtn(c *gin.Context) {
	var req request.SysAuthorityBtnReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = authorityBtnService.SetAuthorityBtn(req)
	if err != nil {
		global.LRAG_LOG.Error("分配失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.allocate_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.allocate_success"), c)
}

// CanRemoveAuthorityBtn
// @Tags      AuthorityBtn
// @Summary   设置权限按钮
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{msg=string}  "删除成功"
// @Router    /authorityBtn/canRemoveAuthorityBtn [post]
func (a *AuthorityBtnApi) CanRemoveAuthorityBtn(c *gin.Context) {
	id := c.Query("id")
	err := authorityBtnService.CanRemoveAuthorityBtn(id)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}
