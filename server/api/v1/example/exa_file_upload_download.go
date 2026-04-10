package example

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/example"
	"github.com/LightningRAG/LightningRAG/server/model/example/request"
	exampleRes "github.com/LightningRAG/LightningRAG/server/model/example/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

type FileUploadAndDownloadApi struct{}

// UploadFile
// @Tags      ExaFileUploadAndDownload
// @Summary   上传文件示例
// @Security  ApiKeyAuth
// @accept    multipart/form-data
// @Produce   application/json
// @Param     file  formData  file                                                           true  "上传文件示例"
// @Success   200   {object}  response.Response{data=exampleRes.ExaFileResponse,msg=string}  "上传文件示例,返回包括文件详情"
// @Router    /fileUploadAndDownload/upload [post]
func (b *FileUploadAndDownloadApi) UploadFile(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	noSave := c.DefaultQuery("noSave", "0")
	_, header, err := c.Request.FormFile("file")
	classId, _ := strconv.Atoi(c.DefaultPostForm("classId", "0"))
	if err != nil {
		global.LRAG_LOG.Error("接收文件失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "rag.kb.receive_file_failed"), c)
		return
	}
	file, err = fileUploadAndDownloadService.UploadFile(header, noSave, classId) // 文件上传后拿到文件路径
	if err != nil {
		global.LRAG_LOG.Error("上传文件失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "example.upload_failed"), c)
		return
	}
	response.OkWithDetailed(exampleRes.ExaFileResponse{File: file}, i18n.Msg(c, "common.upload_success"), c)
}

// EditFileName 编辑文件名或者备注
func (b *FileUploadAndDownloadApi) EditFileName(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	err := c.ShouldBindJSON(&file)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = fileUploadAndDownloadService.EditFileName(file)
	if err != nil {
		global.LRAG_LOG.Error("编辑失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.edit_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.edit_success"), c)
}

// DeleteFile
// @Tags      ExaFileUploadAndDownload
// @Summary   删除文件
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      example.ExaFileUploadAndDownload  true  "传入文件里面id即可"
// @Success   200   {object}  response.Response{msg=string}     "删除文件"
// @Router    /fileUploadAndDownload/deleteFile [post]
func (b *FileUploadAndDownloadApi) DeleteFile(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	err := c.ShouldBindJSON(&file)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := fileUploadAndDownloadService.DeleteFile(file); err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// GetFileList
// @Tags      ExaFileUploadAndDownload
// @Summary   分页文件列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.ExaAttachmentCategorySearch                                        true  "页码, 每页大小, 分类id"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页文件列表,返回包括列表,总数,页码,每页数量"
// @Router    /fileUploadAndDownload/getFileList [post]
func (b *FileUploadAndDownloadApi) GetFileList(c *gin.Context) {
	var pageInfo request.ExaAttachmentCategorySearch
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := fileUploadAndDownloadService.GetFileRecordInfoList(pageInfo)
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
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// ImportURL
// @Tags      ExaFileUploadAndDownload
// @Summary   导入URL
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      example.ExaFileUploadAndDownload  true  "对象"
// @Success   200   {object}  response.Response{msg=string}     "导入URL"
// @Router    /fileUploadAndDownload/importURL [post]
func (b *FileUploadAndDownloadApi) ImportURL(c *gin.Context) {
	var file []example.ExaFileUploadAndDownload
	err := c.ShouldBindJSON(&file)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := fileUploadAndDownloadService.ImportURL(&file); err != nil {
		global.LRAG_LOG.Error("导入URL失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.import_url_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.import_url_success"), c)
}
