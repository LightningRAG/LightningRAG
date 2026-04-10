package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AutoCodeTemplateApi struct{}

// Preview
// @Tags      AutoCodeTemplate
// @Summary   预览创建后的代码
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.AutoCode                                      true  "预览创建代码"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "预览创建后的代码"
// @Router    /autoCode/preview [post]
func (a *AutoCodeTemplateApi) Preview(c *gin.Context) {
	var info request.AutoCode
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(info, utils.AutoCodeVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = info.Pretreatment()
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	info.PackageT = utils.FirstUpper(info.Package)
	autoCode, err := autoCodeTemplateService.Preview(c.Request.Context(), info)
	if err != nil {
		global.LRAG_LOG.Error(err.Error(), zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.preview_failed_detail", err.Error()), c)
	} else {
		response.OkWithDetailed(gin.H{"autoCode": autoCode}, i18n.Msg(c, "common.preview_success"), c)
	}
}

// Create
// @Tags      AutoCodeTemplate
// @Summary   自动代码模板
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.AutoCode  true  "创建自动代码"
// @Success   200   {string}  string                 "{"success":true,"data":{},"msg":"创建成功"}"
// @Router    /autoCode/createTemp [post]
func (a *AutoCodeTemplateApi) Create(c *gin.Context) {
	var info request.AutoCode
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(info, utils.AutoCodeVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = info.Pretreatment()
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = autoCodeTemplateService.Create(c.Request.Context(), info)
	if err != nil {
		global.LRAG_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithError(c, err)
	} else {
		response.OkWithMessage(i18n.Msg(c, "common.create_success"), c)
	}
}

// AddFunc
// @Tags      AddFunc
// @Summary   增加方法
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.AutoCode  true  "增加方法"
// @Success   200   {string}  string                 "{"success":true,"data":{},"msg":"创建成功"}"
// @Router    /autoCode/addFunc [post]
func (a *AutoCodeTemplateApi) AddFunc(c *gin.Context) {
	var info request.AutoFunc
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	var tempMap map[string]string
	if info.IsPreview {
		info.Router = "填充router"
		info.FuncName = "填充funcName"
		info.Method = "填充method"
		info.Description = "填充description"
		tempMap, err = autoCodeTemplateService.GetApiAndServer(info)
	} else {
		err = autoCodeTemplateService.AddFunc(info)
	}
	if err != nil {
		global.LRAG_LOG.Error("注入失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.inject_failed"), c)
	} else {
		if info.IsPreview {
			response.OkWithDetailed(tempMap, i18n.Msg(c, "common.inject_success"), c)
			return
		}
		response.OkWithMessage(i18n.Msg(c, "common.inject_success"), c)
	}
}
