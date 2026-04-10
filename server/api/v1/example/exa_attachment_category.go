package example

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	common "github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/example"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AttachmentCategoryApi struct{}

// GetCategoryList
// @Tags      GetCategoryList
// @Summary   媒体库分类列表
// @Security  AttachmentCategory
// @Produce   application/json
// @Success   200   {object}  response.Response{data=example.ExaAttachmentCategory,msg=string}  "媒体库分类列表"
// @Router    /attachmentCategory/getCategoryList [get]
func (a *AttachmentCategoryApi) GetCategoryList(c *gin.Context) {
	res, err := attachmentCategoryService.GetCategoryList()
	if err != nil {
		global.LRAG_LOG.Error("获取分类列表失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "example.attachment_category_list_failed"), c)
		return
	}
	response.OkWithData(res, c)
}

// AddCategory
// @Tags      AddCategory
// @Summary   添加媒体库分类
// @Security  AttachmentCategory
// @accept    application/json
// @Produce   application/json
// @Param     data  body      example.ExaAttachmentCategory  true  "媒体库分类数据"
// @Success   200   {object}  response.Response{msg=string}   "添加媒体库分类"
// @Router    /attachmentCategory/addCategory [post]
func (a *AttachmentCategoryApi) AddCategory(c *gin.Context) {
	var req example.ExaAttachmentCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		global.LRAG_LOG.Error("参数错误!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "validation.param_invalid"), c)
		return
	}

	if err := attachmentCategoryService.AddCategory(&req); err != nil {
		global.LRAG_LOG.Error("创建/更新失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.upsert_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.upsert_success"), c)
}

// DeleteCategory
// @Tags      DeleteCategory
// @Summary   删除分类
// @Security  AttachmentCategory
// @accept    application/json
// @Produce   application/json
// @Param     data  body      common.GetById                true  "分类id"
// @Success   200   {object}  response.Response{msg=string}  "删除分类"
// @Router    /attachmentCategory/deleteCategory [post]
func (a *AttachmentCategoryApi) DeleteCategory(c *gin.Context) {
	var req common.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(i18n.Msg(c, "validation.param_invalid"), c)
		return
	}

	if req.ID == 0 {
		response.FailWithMessage(i18n.Msg(c, "validation.param_invalid"), c)
		return
	}

	if err := attachmentCategoryService.DeleteCategory(&req.ID); err != nil {
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}

	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}
