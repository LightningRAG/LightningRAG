package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SettingsApi struct{}

// ========== 向量存储配置 ==========

func (s *SettingsApi) CreateVectorStoreConfig(c *gin.Context) {
	var req request.VectorStoreConfigCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	cfg, err := knowledgeBaseService.CreateVectorStoreConfig(c.Request.Context(), req)
	if err != nil {
		global.LRAG_LOG.Error("创建向量存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(cfg, c)
}

func (s *SettingsApi) UpdateVectorStoreConfig(c *gin.Context) {
	var req request.VectorStoreConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := knowledgeBaseService.UpdateVectorStoreConfig(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("更新向量存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

func (s *SettingsApi) DeleteVectorStoreConfig(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := knowledgeBaseService.DeleteVectorStoreConfig(c.Request.Context(), req.ID); err != nil {
		global.LRAG_LOG.Error("删除向量存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

func (s *SettingsApi) GetVectorStoreConfig(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	cfg, err := knowledgeBaseService.GetVectorStoreConfig(c.Request.Context(), req.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取向量存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(cfg, c)
}

func (s *SettingsApi) ListVectorStoreConfigsFull(c *gin.Context) {
	var req request.VectorStoreConfigList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := knowledgeBaseService.ListVectorStoreConfigsFull(c.Request.Context(), req)
	if err != nil {
		global.LRAG_LOG.Error("获取向量存储配置列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// ========== 文件存储配置 ==========

func (s *SettingsApi) CreateFileStorageConfig(c *gin.Context) {
	var req request.FileStorageConfigCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	cfg, err := knowledgeBaseService.CreateFileStorageConfig(c.Request.Context(), req)
	if err != nil {
		global.LRAG_LOG.Error("创建文件存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(cfg, c)
}

func (s *SettingsApi) UpdateFileStorageConfig(c *gin.Context) {
	var req request.FileStorageConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := knowledgeBaseService.UpdateFileStorageConfig(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("更新文件存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

func (s *SettingsApi) DeleteFileStorageConfig(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := knowledgeBaseService.DeleteFileStorageConfig(c.Request.Context(), req.ID); err != nil {
		global.LRAG_LOG.Error("删除文件存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

func (s *SettingsApi) GetFileStorageConfig(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	cfg, err := knowledgeBaseService.GetFileStorageConfig(c.Request.Context(), req.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取文件存储配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(cfg, c)
}

func (s *SettingsApi) ListFileStorageConfigsFull(c *gin.Context) {
	var req request.FileStorageConfigList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := knowledgeBaseService.ListFileStorageConfigsFull(c.Request.Context(), req)
	if err != nil {
		global.LRAG_LOG.Error("获取文件存储配置列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}
