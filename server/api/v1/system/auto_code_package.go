package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	common "github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
)

type AutoCodePackageApi struct{}

// Create
// @Tags      AutoCodePackage
// @Summary   创建package
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.SysAutoCodePackageCreate                                         true  "创建package"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "创建package成功"
// @Router    /autoCode/createPackage [post]
func (a *AutoCodePackageApi) Create(c *gin.Context) {
	var info request.SysAutoCodePackageCreate
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(info, utils.AutoPackageVerify); err != nil {
		response.FailWithError(c, err)
		return
	}
	if strings.Contains(info.PackageName, "\\") || strings.Contains(info.PackageName, "/") || strings.Contains(info.PackageName, "..") {
		response.FailWithMessage(i18n.Msg(c, "validation.package_name_invalid"), c)
		return
	} // PackageName可能导致路径穿越的问题 / 和 \ 都要防止
	err := autoCodePackageService.Create(c.Request.Context(), &info)
	if err != nil {
		global.LRAG_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.create_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.create_success"), c)
}

// Delete
// @Tags      AutoCode
// @Summary   删除package
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      common.GetById                                         true  "创建package"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "删除package成功"
// @Router    /autoCode/delPackage [post]
func (a *AutoCodePackageApi) Delete(c *gin.Context) {
	var info common.GetById
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err := autoCodePackageService.Delete(c.Request.Context(), info)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// All
// @Tags      AutoCodePackage
// @Summary   获取package
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=map[string]interface{},msg=string}  "创建package成功"
// @Router    /autoCode/getPackage [post]
func (a *AutoCodePackageApi) All(c *gin.Context) {
	data, err := autoCodePackageService.All(c.Request.Context())
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(gin.H{"pkgs": data}, i18n.Msg(c, "common.fetch_success"), c)
}

// Templates
// @Tags      AutoCodePackage
// @Summary   获取package
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=map[string]interface{},msg=string}  "创建package成功"
// @Router    /autoCode/getTemplates [get]
func (a *AutoCodePackageApi) Templates(c *gin.Context) {
	data, err := autoCodePackageService.Templates(c.Request.Context())
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(data, i18n.Msg(c, "common.fetch_success"), c)
}
