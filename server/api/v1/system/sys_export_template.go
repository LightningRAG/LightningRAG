package system

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/LightningRAG/LightningRAG/server/service"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 用于token一次性存储
var (
	exportTokenCache      = make(map[string]interface{})
	exportTokenExpiration = make(map[string]time.Time)
	tokenMutex            sync.RWMutex
)

// 五分钟检测窗口过期
func cleanupExpiredTokens() {
	for {
		time.Sleep(5 * time.Minute)
		tokenMutex.Lock()
		now := time.Now()
		for token, expiry := range exportTokenExpiration {
			if now.After(expiry) {
				delete(exportTokenCache, token)
				delete(exportTokenExpiration, token)
			}
		}
		tokenMutex.Unlock()
	}
}

func init() {
	go cleanupExpiredTokens()
}

type SysExportTemplateApi struct {
}

var sysExportTemplateService = service.ServiceGroupApp.SystemServiceGroup.SysExportTemplateService

// PreviewSQL 预览最终生成的SQL
// @Tags     SysExportTemplate
// @Summary  预览最终生成的SQL（不执行查询，仅返回SQL字符串）
// @Security ApiKeyAuth
// @accept   application/json
// @Produce  application/json
// @Param    templateID query string true  "导出模板ID"
// @Param    params     query string false "查询参数编码字符串，参考 ExportExcel 组件"
// @Success  200  {object}  response.Response{data=map[string]string} "获取成功"
// @Router   /sysExportTemplate/previewSQL [get]
func (sysExportTemplateApi *SysExportTemplateApi) PreviewSQL(c *gin.Context) {
	templateID := c.Query("templateID")
	if templateID == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.template_id_required"), c)
		return
	}

	// 直接复用导出接口的参数组织方式：使用 URL Query，其中 params 为内部编码的查询字符串
	queryParams := c.Request.URL.Query()

	if sqlPreview, err := sysExportTemplateService.PreviewSQL(templateID, queryParams); err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
	} else {
		response.OkWithData(gin.H{"sql": sqlPreview}, c)
	}
}

// CreateSysExportTemplate 创建导出模板
// @Tags SysExportTemplate
// @Summary 创建导出模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body system.SysExportTemplate true "创建导出模板"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sysExportTemplate/createSysExportTemplate [post]
func (sysExportTemplateApi *SysExportTemplateApi) CreateSysExportTemplate(c *gin.Context) {
	var sysExportTemplate system.SysExportTemplate
	err := c.ShouldBindJSON(&sysExportTemplate)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	verify := utils.Rules{
		"Name": {utils.NotEmpty()},
	}
	if err := utils.Verify(sysExportTemplate, verify); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := sysExportTemplateService.CreateSysExportTemplate(&sysExportTemplate); err != nil {
		global.LRAG_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.create_failed"), c)
	} else {
		response.OkWithMessage(i18n.Msg(c, "common.create_success"), c)
	}
}

// DeleteSysExportTemplate 删除导出模板
// @Tags SysExportTemplate
// @Summary 删除导出模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body system.SysExportTemplate true "删除导出模板"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sysExportTemplate/deleteSysExportTemplate [delete]
func (sysExportTemplateApi *SysExportTemplateApi) DeleteSysExportTemplate(c *gin.Context) {
	var sysExportTemplate system.SysExportTemplate
	err := c.ShouldBindJSON(&sysExportTemplate)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := sysExportTemplateService.DeleteSysExportTemplate(sysExportTemplate); err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
	} else {
		response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
	}
}

// DeleteSysExportTemplateByIds 批量删除导出模板
// @Tags SysExportTemplate
// @Summary 批量删除导出模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除导出模板"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"批量删除成功"}"
// @Router /sysExportTemplate/deleteSysExportTemplateByIds [delete]
func (sysExportTemplateApi *SysExportTemplateApi) DeleteSysExportTemplateByIds(c *gin.Context) {
	var IDS request.IdsReq
	err := c.ShouldBindJSON(&IDS)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := sysExportTemplateService.DeleteSysExportTemplateByIds(IDS); err != nil {
		global.LRAG_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.batch_delete_failed"), c)
	} else {
		response.OkWithMessage(i18n.Msg(c, "common.batch_delete_success"), c)
	}
}

// UpdateSysExportTemplate 更新导出模板
// @Tags SysExportTemplate
// @Summary 更新导出模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body system.SysExportTemplate true "更新导出模板"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sysExportTemplate/updateSysExportTemplate [put]
func (sysExportTemplateApi *SysExportTemplateApi) UpdateSysExportTemplate(c *gin.Context) {
	var sysExportTemplate system.SysExportTemplate
	err := c.ShouldBindJSON(&sysExportTemplate)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	verify := utils.Rules{
		"Name": {utils.NotEmpty()},
	}
	if err := utils.Verify(sysExportTemplate, verify); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := sysExportTemplateService.UpdateSysExportTemplate(sysExportTemplate); err != nil {
		global.LRAG_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.update_failed"), c)
	} else {
		response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
	}
}

// FindSysExportTemplate 用id查询导出模板
// @Tags SysExportTemplate
// @Summary 用id查询导出模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query system.SysExportTemplate true "用id查询导出模板"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sysExportTemplate/findSysExportTemplate [get]
func (sysExportTemplateApi *SysExportTemplateApi) FindSysExportTemplate(c *gin.Context) {
	var sysExportTemplate system.SysExportTemplate
	err := c.ShouldBindQuery(&sysExportTemplate)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if resysExportTemplate, err := sysExportTemplateService.GetSysExportTemplate(sysExportTemplate.ID); err != nil {
		global.LRAG_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.query_failed"), c)
	} else {
		response.OkWithData(gin.H{"resysExportTemplate": resysExportTemplate}, c)
	}
}

// GetSysExportTemplateList 分页获取导出模板列表
// @Tags SysExportTemplate
// @Summary 分页获取导出模板列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query systemReq.SysExportTemplateSearch true "分页获取导出模板列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sysExportTemplate/getSysExportTemplateList [get]
func (sysExportTemplateApi *SysExportTemplateApi) GetSysExportTemplateList(c *gin.Context) {
	var pageInfo systemReq.SysExportTemplateSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if list, total, err := sysExportTemplateService.GetSysExportTemplateInfoList(pageInfo); err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
	} else {
		response.OkWithDetailed(response.PageResult{
			List:     list,
			Total:    total,
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
		}, i18n.Msg(c, "common.fetch_success"), c)
	}
}

// ExportExcel 导出表格token
// @Tags SysExportTemplate
// @Summary 导出表格
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Router /sysExportTemplate/exportExcel [get]
func (sysExportTemplateApi *SysExportTemplateApi) ExportExcel(c *gin.Context) {
	templateID := c.Query("templateID")
	if templateID == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.template_id_required"), c)
		return
	}

	queryParams := c.Request.URL.Query()

	//创造一次性token
	token := utils.RandomString(32) // 随机32位

	// 记录本次请求参数
	exportParams := map[string]interface{}{
		"templateID":  templateID,
		"queryParams": queryParams,
	}

	// 参数保留记录完成鉴权
	tokenMutex.Lock()
	exportTokenCache[token] = exportParams
	exportTokenExpiration[token] = time.Now().Add(30 * time.Minute)
	tokenMutex.Unlock()

	// 生成一次性链接
	exportUrl := fmt.Sprintf("/sysExportTemplate/exportExcelByToken?token=%s", token)
	response.OkWithData(exportUrl, c)
}

// ExportExcelByToken 导出表格
// @Tags ExportExcelByToken
// @Summary 导出表格
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Router /sysExportTemplate/exportExcelByToken [get]
func (sysExportTemplateApi *SysExportTemplateApi) ExportExcelByToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.export_token_required"), c)
		return
	}

	// 获取token并且从缓存中剔除
	tokenMutex.RLock()
	exportParamsRaw, exists := exportTokenCache[token]
	expiry, _ := exportTokenExpiration[token]
	tokenMutex.RUnlock()

	if !exists || time.Now().After(expiry) {
		global.LRAG_LOG.Error("导出token无效或已过期!")
		response.FailWithMessage(i18n.Msg(c, "validation.export_token_invalid"), c)
		return
	}

	// 从token获取参数
	exportParams, ok := exportParamsRaw.(map[string]interface{})
	if !ok {
		global.LRAG_LOG.Error("解析导出参数失败!")
		response.FailWithMessage(i18n.Msg(c, "validation.parse_export_params_failed"), c)
		return
	}

	// 获取导出参数
	templateID := exportParams["templateID"].(string)
	queryParams := exportParams["queryParams"].(url.Values)

	// 清理一次性token
	tokenMutex.Lock()
	delete(exportTokenCache, token)
	delete(exportTokenExpiration, token)
	tokenMutex.Unlock()

	// 导出
	if file, name, err := sysExportTemplateService.ExportExcel(templateID, queryParams); err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name+utils.RandomString(6)+".xlsx"))
		c.Header("success", "true")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file.Bytes())
	}
}

// ExportTemplate 导出表格模板
// @Tags SysExportTemplate
// @Summary 导出表格模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Router /sysExportTemplate/exportTemplate [get]
func (sysExportTemplateApi *SysExportTemplateApi) ExportTemplate(c *gin.Context) {
	templateID := c.Query("templateID")
	if templateID == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.template_id_required"), c)
		return
	}

	// 创造一次性token
	token := utils.RandomString(32) // 随机32位

	// 记录本次请求参数
	exportParams := map[string]interface{}{
		"templateID": templateID,
		"isTemplate": true,
	}

	// 参数保留记录完成鉴权
	tokenMutex.Lock()
	exportTokenCache[token] = exportParams
	exportTokenExpiration[token] = time.Now().Add(30 * time.Minute)
	tokenMutex.Unlock()

	// 生成一次性链接
	exportUrl := fmt.Sprintf("/sysExportTemplate/exportTemplateByToken?token=%s", token)
	response.OkWithData(exportUrl, c)
}

// ExportTemplateByToken 通过token导出表格模板
// @Tags ExportTemplateByToken
// @Summary 通过token导出表格模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Router /sysExportTemplate/exportTemplateByToken [get]
func (sysExportTemplateApi *SysExportTemplateApi) ExportTemplateByToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.export_token_required"), c)
		return
	}

	// 获取token并且从缓存中剔除
	tokenMutex.RLock()
	exportParamsRaw, exists := exportTokenCache[token]
	expiry, _ := exportTokenExpiration[token]
	tokenMutex.RUnlock()

	if !exists || time.Now().After(expiry) {
		global.LRAG_LOG.Error("导出token无效或已过期!")
		response.FailWithMessage(i18n.Msg(c, "validation.export_token_invalid"), c)
		return
	}

	// 从token获取参数
	exportParams, ok := exportParamsRaw.(map[string]interface{})
	if !ok {
		global.LRAG_LOG.Error("解析导出参数失败!")
		response.FailWithMessage(i18n.Msg(c, "validation.parse_export_params_failed"), c)
		return
	}

	// 检查是否为模板导出
	isTemplate, _ := exportParams["isTemplate"].(bool)
	if !isTemplate {
		global.LRAG_LOG.Error("token类型错误!")
		response.FailWithMessage(i18n.Msg(c, "validation.token_type_wrong"), c)
		return
	}

	// 获取导出参数
	templateID := exportParams["templateID"].(string)

	// 清理一次性token
	tokenMutex.Lock()
	delete(exportTokenCache, token)
	delete(exportTokenExpiration, token)
	tokenMutex.Unlock()

	// 导出模板
	if file, name, err := sysExportTemplateService.ExportTemplate(templateID); err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name+"模板.xlsx"))
		c.Header("success", "true")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file.Bytes())
	}
}

// ImportExcel 导入表格
// @Tags SysImportTemplate
// @Summary 导入表格
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Router /sysExportTemplate/importExcel [post]
func (sysExportTemplateApi *SysExportTemplateApi) ImportExcel(c *gin.Context) {
	templateID := c.Query("templateID")
	if templateID == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.template_id_required"), c)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		global.LRAG_LOG.Error("文件获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "sys.export.file_fetch_failed"), c)
		return
	}
	if err := sysExportTemplateService.ImportExcel(templateID, file); err != nil {
		global.LRAG_LOG.Error(err.Error(), zap.Error(err))
		response.FailWithError(c, err)
	} else {
		response.OkWithMessage(i18n.Msg(c, "sys.version.import_success"), c)
	}
}
