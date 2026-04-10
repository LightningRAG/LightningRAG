package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DictionaryApi struct{}

// CreateSysDictionary
// @Tags      SysDictionary
// @Summary   创建SysDictionary
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDictionary           true  "SysDictionary模型"
// @Success   200   {object}  response.Response{msg=string}  "创建SysDictionary"
// @Router    /sysDictionary/createSysDictionary [post]
func (s *DictionaryApi) CreateSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	err := c.ShouldBindJSON(&dictionary)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = dictionaryService.CreateSysDictionary(dictionary)
	if err != nil {
		global.LRAG_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.create_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.create_success"), c)
}

// DeleteSysDictionary
// @Tags      SysDictionary
// @Summary   删除SysDictionary
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDictionary           true  "SysDictionary模型"
// @Success   200   {object}  response.Response{msg=string}  "删除SysDictionary"
// @Router    /sysDictionary/deleteSysDictionary [delete]
func (s *DictionaryApi) DeleteSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	err := c.ShouldBindJSON(&dictionary)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = dictionaryService.DeleteSysDictionary(dictionary)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// UpdateSysDictionary
// @Tags      SysDictionary
// @Summary   更新SysDictionary
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDictionary           true  "SysDictionary模型"
// @Success   200   {object}  response.Response{msg=string}  "更新SysDictionary"
// @Router    /sysDictionary/updateSysDictionary [put]
func (s *DictionaryApi) UpdateSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	err := c.ShouldBindJSON(&dictionary)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = dictionaryService.UpdateSysDictionary(&dictionary)
	if err != nil {
		global.LRAG_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.update_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

// FindSysDictionary
// @Tags      SysDictionary
// @Summary   用id查询SysDictionary
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     system.SysDictionary                                       true  "ID或字典英名"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "用id查询SysDictionary"
// @Router    /sysDictionary/findSysDictionary [get]
func (s *DictionaryApi) FindSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	err := c.ShouldBindQuery(&dictionary)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	sysDictionary, err := dictionaryService.GetSysDictionary(dictionary.Type, dictionary.ID, dictionary.Status)
	if err != nil {
		global.LRAG_LOG.Error("字典未创建或未开启!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "sys.dictionary.not_enabled"), c)
		return
	}
	response.OkWithDetailed(gin.H{"resysDictionary": sysDictionary}, i18n.Msg(c, "common.query_success"), c)
}

// GetSysDictionaryList
// @Tags      SysDictionary
// @Summary   分页获取SysDictionary列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.SysDictionarySearch                                    true  "字典 name 或者 type"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页获取SysDictionary列表,返回包括列表,总数,页码,每页数量"
// @Router    /sysDictionary/getSysDictionaryList [get]
func (s *DictionaryApi) GetSysDictionaryList(c *gin.Context) {
	var dictionary request.SysDictionarySearch
	err := c.ShouldBindQuery(&dictionary)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, err := dictionaryService.GetSysDictionaryInfoList(c, dictionary)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(list, i18n.Msg(c, "common.fetch_success"), c)
}

// ExportSysDictionary
// @Tags      SysDictionary
// @Summary   导出字典JSON（包含字典详情）
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     system.SysDictionary                                       true  "字典ID"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "导出字典JSON"
// @Router    /sysDictionary/exportSysDictionary [get]
func (s *DictionaryApi) ExportSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	err := c.ShouldBindQuery(&dictionary)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if dictionary.ID == 0 {
		response.FailWithMessage(i18n.Msg(c, "validation.dict_id_required"), c)
		return
	}
	exportData, err := dictionaryService.ExportSysDictionary(dictionary.ID)
	if err != nil {
		global.LRAG_LOG.Error("导出失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.export_failed"), c)
		return
	}
	response.OkWithDetailed(exportData, i18n.Msg(c, "common.export_success"), c)
}

// ImportSysDictionary
// @Tags      SysDictionary
// @Summary   导入字典JSON（包含字典详情）
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.ImportSysDictionaryRequest     true  "字典JSON数据"
// @Success   200   {object}  response.Response{msg=string}          "导入字典"
// @Router    /sysDictionary/importSysDictionary [post]
func (s *DictionaryApi) ImportSysDictionary(c *gin.Context) {
	var req request.ImportSysDictionaryRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = dictionaryService.ImportSysDictionary(req.Json)
	if err != nil {
		global.LRAG_LOG.Error("导入失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.import_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "sys.version.import_success"), c)
}
