package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	"github.com/LightningRAG/LightningRAG/server/rag/tools"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LLMProviderApi struct{}

// ListAvailableProviders 列出指定场景类型可用的提供商（用于添加模型时的下拉选项）
func (l *LLMProviderApi) ListAvailableProviders(c *gin.Context) {
	var req request.LLMProviderListAvailableReq
	_ = c.ShouldBindJSON(&req)
	opts := registry.ListProvidersByScenario(req.ScenarioType, req.ScenarioTypes)
	response.OkWithData(opts, c)
}

// ListProviders 列出用户可用的 LLM 提供商（管理员配置+用户自己的），可按 scenarioType 过滤，并标记默认模型
func (l *LLMProviderApi) ListProviders(c *gin.Context) {
	var req request.LLMProviderListReq
	_ = c.ShouldBindJSON(&req)
	uid := utils.GetUserID(c)
	authorityId := utils.GetUserAuthorityId(c)
	list, err := llmProviderService.ListProviders(c.Request.Context(), uid, authorityId, req.ScenarioType)
	if err != nil {
		global.LRAG_LOG.Error("获取模型列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// ListUserModels 列出用户自定义模型
func (l *LLMProviderApi) ListUserModels(c *gin.Context) {
	uid := utils.GetUserID(c)
	list, err := llmProviderService.ListUserModels(c.Request.Context(), uid)
	if err != nil {
		global.LRAG_LOG.Error("获取用户模型列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// AddUserModel 添加用户自定义模型
func (l *LLMProviderApi) AddUserModel(c *gin.Context) {
	var req request.LLMProviderAddUser
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := llmProviderService.AddUserModel(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("添加用户模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.add_success"), c)
}

// UpdateUserModel 更新用户自定义模型
func (l *LLMProviderApi) UpdateUserModel(c *gin.Context) {
	var req request.LLMProviderUpdateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := llmProviderService.UpdateUserModel(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("更新用户模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

// DeleteUserModel 删除用户自定义模型
func (l *LLMProviderApi) DeleteUserModel(c *gin.Context) {
	var req request.LLMProviderDeleteUser
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := llmProviderService.DeleteUserModel(c.Request.Context(), uid, req.ID); err != nil {
		global.LRAG_LOG.Error("删除用户模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// SetAuthorityDefaultLLM 设置角色默认模型（需管理员权限）
func (l *LLMProviderApi) SetAuthorityDefaultLLM(c *gin.Context) {
	var req request.SetAuthorityDefaultLLMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := llmProviderService.SetAuthorityDefaultLLM(c.Request.Context(), req); err != nil {
		global.LRAG_LOG.Error("设置角色默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// ClearAuthorityDefaultLLM 清除角色某类型默认模型
func (l *LLMProviderApi) ClearAuthorityDefaultLLM(c *gin.Context) {
	var req request.ClearAuthorityDefaultLLMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := llmProviderService.ClearAuthorityDefaultLLM(c.Request.Context(), req.AuthorityId, req.ModelType); err != nil {
		global.LRAG_LOG.Error("清除角色默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.cleared"), c)
}

// GetAuthorityDefaultLLMs 获取角色各类型默认模型
func (l *LLMProviderApi) GetAuthorityDefaultLLMs(c *gin.Context) {
	var req request.GetAuthorityDefaultLLMsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	list, err := llmProviderService.GetAuthorityDefaultLLMs(c.Request.Context(), req.AuthorityId)
	if err != nil {
		global.LRAG_LOG.Error("获取角色默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// SetUserDefaultLLM 设置用户默认模型（用户为自己设置）
func (l *LLMProviderApi) SetUserDefaultLLM(c *gin.Context) {
	var req request.SetUserDefaultLLMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := llmProviderService.SetUserDefaultLLM(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("设置用户默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// GetUserDefaultLLMs 获取用户各类型默认模型
func (l *LLMProviderApi) GetUserDefaultLLMs(c *gin.Context) {
	uid := utils.GetUserID(c)
	list, err := llmProviderService.GetUserDefaultLLMs(c.Request.Context(), uid)
	if err != nil {
		global.LRAG_LOG.Error("获取用户默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// ClearUserDefaultLLM 清除用户某类型默认模型
func (l *LLMProviderApi) ClearUserDefaultLLM(c *gin.Context) {
	var req request.ClearUserDefaultLLMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := llmProviderService.ClearUserDefaultLLM(c.Request.Context(), uid, req.ModelType); err != nil {
		global.LRAG_LOG.Error("清除用户默认模型失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.cleared"), c)
}

// ListWebSearchProviders 列出可用的网页搜索引擎（供前端配置选择）
func (l *LLMProviderApi) ListWebSearchProviders(c *gin.Context) {
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

// GetWebSearchConfig 获取当前用户互联网搜索配置
func (l *LLMProviderApi) GetWebSearchConfig(c *gin.Context) {
	uid := utils.GetUserID(c)
	provider, config, useSystemDefault, err := llmProviderService.GetWebSearchConfig(c.Request.Context(), uid)
	if err != nil {
		global.LRAG_LOG.Error("获取互联网搜索配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(map[string]any{"provider": provider, "config": config, "useSystemDefault": useSystemDefault}, c)
}

// SetWebSearchConfig 设置当前用户互联网搜索配置
func (l *LLMProviderApi) SetWebSearchConfig(c *gin.Context) {
	var req request.WebSearchConfigSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := llmProviderService.SetWebSearchConfig(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("设置互联网搜索配置失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.save_success"), c)
}
