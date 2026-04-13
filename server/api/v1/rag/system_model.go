package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/tools"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SystemModelApi struct{}

// ListAdminModels 列出管理员模型（系统全局模型）
func (a *SystemModelApi) ListAdminModels(c *gin.Context) {
	var req request.AdminModelList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := systemModelService.ListAdminModels(c.Request.Context(), req)
	if err != nil {
		global.LRAG_LOG.Error("获取管理员模型列表失败", zap.Error(err))
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

// CreateAdminModel 添加管理员模型
func (a *SystemModelApi) CreateAdminModel(c *gin.Context) {
	var req request.AdminModelCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	m, err := systemModelService.CreateAdminModel(c.Request.Context(), req)
	if err != nil {
		global.LRAG_LOG.Error("添加管理员模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(m, c)
}

// UpdateAdminModel 更新管理员模型
func (a *SystemModelApi) UpdateAdminModel(c *gin.Context) {
	var req request.AdminModelUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.UpdateAdminModel(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("更新管理员模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

// DeleteAdminModel 删除管理员模型
func (a *SystemModelApi) DeleteAdminModel(c *gin.Context) {
	var req request.AdminModelDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.DeleteAdminModel(c.Request.Context(), req.ID); err != nil {
		global.LRAG_LOG.Error("删除管理员模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// GetSystemDefaults 获取系统全局默认模型
func (a *SystemModelApi) GetSystemDefaults(c *gin.Context) {
	list, err := systemModelService.GetSystemDefaults(c.Request.Context())
	if err != nil {
		global.LRAG_LOG.Error("获取系统默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// SetSystemDefault 设置系统全局默认模型
func (a *SystemModelApi) SetSystemDefault(c *gin.Context) {
	var req request.SystemDefaultModelSet
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.SetSystemDefault(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("设置系统默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// ClearSystemDefault 清除系统全局默认模型
func (a *SystemModelApi) ClearSystemDefault(c *gin.Context) {
	var req request.SystemDefaultModelClear
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.ClearSystemDefault(c.Request.Context(), req.ModelType); err != nil {
		global.LRAG_LOG.Error("清除系统默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.cleared"), c)
}

// ListSystemWebSearchProviders 列出可用的网页搜索引擎（供系统配置选择）
func (a *SystemModelApi) ListSystemWebSearchProviders(c *gin.Context) {
	list := tools.ListWebSearchProviders()
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		schema := p.ConfigSchema()
		schemaArr := make([]map[string]any, 0, len(schema))
		for _, f := range schema {
			schemaArr = append(schemaArr, map[string]any{
				"key":         f.Key,
				"label":       f.Label,
				"required":    f.Required,
				"secret":      f.Secret,
				"placeholder": f.Placeholder,
			})
		}
		out = append(out, map[string]any{
			"id":           p.ID(),
			"displayName":  p.DisplayName(),
			"configSchema": schemaArr,
		})
	}
	response.OkWithData(out, c)
}

// GetSystemWebSearchConfig 获取系统全局默认互联网搜索配置
func (a *SystemModelApi) GetSystemWebSearchConfig(c *gin.Context) {
	provider, config, err := systemModelService.GetSystemWebSearchConfig(c.Request.Context())
	if err != nil {
		global.LRAG_LOG.Error("获取系统互联网搜索配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(map[string]any{"provider": provider, "config": config}, c)
}

// SetSystemWebSearchConfig 设置系统全局默认互联网搜索配置
func (a *SystemModelApi) SetSystemWebSearchConfig(c *gin.Context) {
	var req request.SystemWebSearchConfigSet
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.SetSystemWebSearchConfig(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("设置系统互联网搜索配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.save_success"), c)
}

// ClearSystemWebSearchConfig 清除系统全局默认互联网搜索配置
func (a *SystemModelApi) ClearSystemWebSearchConfig(c *gin.Context) {
	if err := systemModelService.ClearSystemWebSearchConfig(c.Request.Context()); err != nil {
		global.LRAG_LOG.Error("清除系统互联网搜索配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.cleared"), c)
}

// ========== 全局共享知识库 ==========

// ListGlobalKnowledgeBases 获取全局共享知识库列表
func (a *SystemModelApi) ListGlobalKnowledgeBases(c *gin.Context) {
	list, err := systemModelService.ListGlobalKnowledgeBases(c.Request.Context())
	if err != nil {
		global.LRAG_LOG.Error("获取全局知识库列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// SetGlobalKnowledgeBase 添加全局共享知识库
func (a *SystemModelApi) SetGlobalKnowledgeBase(c *gin.Context) {
	var req request.GlobalKnowledgeBaseSet
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.SetGlobalKnowledgeBase(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("设置全局知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// RemoveGlobalKnowledgeBase 移除全局共享知识库
func (a *SystemModelApi) RemoveGlobalKnowledgeBase(c *gin.Context) {
	var req request.GlobalKnowledgeBaseRemove
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := systemModelService.RemoveGlobalKnowledgeBase(c.Request.Context(), req.KnowledgeBaseID); err != nil {
		global.LRAG_LOG.Error("移除全局知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.remove_success"), c)
}

// ListAllKnowledgeBases 列出所有知识库（供管理员选择设为全局共享）
func (a *SystemModelApi) ListAllKnowledgeBases(c *gin.Context) {
	list, err := systemModelService.ListAllKnowledgeBases(c.Request.Context())
	if err != nil {
		global.LRAG_LOG.Error("获取知识库列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}
