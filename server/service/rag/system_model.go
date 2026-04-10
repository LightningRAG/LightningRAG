package rag

import (
	"context"
	"errors"
	"slices"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"gorm.io/gorm"
)

// SystemModelService 系统全局模型管理服务
type SystemModelService struct{}

// ================== 管理员模型 CRUD (RagLLMProvider) ==================

// ListAdminModels 列出管理员模型（分页、按场景过滤）
func (s *SystemModelService) ListAdminModels(ctx context.Context, req request.AdminModelList) ([]rag.RagLLMProvider, int64, error) {
	var list []rag.RagLLMProvider
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagLLMProvider{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if req.PageSize > 0 {
		db = db.Scopes(req.Paginate())
	}
	if err := db.Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if req.ScenarioType != "" {
		var filtered []rag.RagLLMProvider
		for _, m := range list {
			if slices.Contains(m.ModelTypes, req.ScenarioType) {
				filtered = append(filtered, m)
			}
		}
		return filtered, int64(len(filtered)), nil
	}
	return list, total, nil
}

// CreateAdminModel 添加管理员模型
func (s *SystemModelService) CreateAdminModel(ctx context.Context, req request.AdminModelCreate) (*rag.RagLLMProvider, error) {
	modelTypes := req.ModelTypes
	if len(modelTypes) == 0 {
		modelTypes = []string{"chat"}
	}
	m := &rag.RagLLMProvider{
		Name:                 req.Name,
		ModelName:            req.ModelName,
		ModelTypes:           modelTypes,
		BaseURL:              req.BaseURL,
		APIKey:               req.APIKey,
		MaxContextTokens:     req.MaxContextTokens,
		SupportsDeepThinking: req.SupportsDeepThinking,
		SupportsToolCall:     req.SupportsToolCall,
		ShareScope:           req.ShareScope,
		Enabled:              req.Enabled,
	}
	if m.ShareScope == "" {
		m.ShareScope = "all"
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// UpdateAdminModel 更新管理员模型
func (s *SystemModelService) UpdateAdminModel(ctx context.Context, req request.AdminModelUpdate) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ModelName != "" {
		updates["model_name"] = req.ModelName
	}
	if len(req.ModelTypes) > 0 {
		updates["model_types"] = req.ModelTypes
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	updates["max_context_tokens"] = req.MaxContextTokens
	updates["supports_deep_thinking"] = req.SupportsDeepThinking
	updates["supports_tool_call"] = req.SupportsToolCall
	if req.ShareScope != "" {
		updates["share_scope"] = req.ShareScope
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	// ModelTypes 需用 Select 才能更新 JSON 字段
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagLLMProvider{}).Where("id = ?", req.ID).
		Select("Name", "ModelName", "ModelTypes", "BaseURL", "APIKey", "MaxContextTokens", "SupportsDeepThinking", "SupportsToolCall", "ShareScope", "Enabled").
		Updates(&rag.RagLLMProvider{
			Name:                 req.Name,
			ModelName:            req.ModelName,
			ModelTypes:           req.ModelTypes,
			BaseURL:              req.BaseURL,
			APIKey:               req.APIKey,
			MaxContextTokens:     req.MaxContextTokens,
			SupportsDeepThinking: req.SupportsDeepThinking,
			SupportsToolCall:     req.SupportsToolCall,
			ShareScope:           req.ShareScope,
			Enabled:              req.Enabled != nil && *req.Enabled,
		}).Error
}

// DeleteAdminModel 删除管理员模型
func (s *SystemModelService) DeleteAdminModel(ctx context.Context, id uint) error {
	return global.LRAG_DB.WithContext(ctx).Where("id = ?", id).Delete(&rag.RagLLMProvider{}).Error
}

// ================== 系统全局默认模型 ==================

// GetSystemDefaults 获取所有系统全局默认模型
func (s *SystemModelService) GetSystemDefaults(ctx context.Context) ([]map[string]any, error) {
	var list []rag.RagSystemDefaultModel
	if err := global.LRAG_DB.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(list))
	for _, d := range list {
		item := map[string]any{
			"modelType":     d.ModelType,
			"llmProviderId": d.LLMProviderID,
		}
		// 补充模型详情
		var provider rag.RagLLMProvider
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", d.LLMProviderID).First(&provider).Error; err == nil {
			item["providerName"] = provider.Name
			item["modelName"] = provider.ModelName
		}
		result = append(result, item)
	}
	return result, nil
}

// SetSystemDefault 设置系统全局默认模型
func (s *SystemModelService) SetSystemDefault(ctx context.Context, req request.SystemDefaultModelSet) error {
	var existing rag.RagSystemDefaultModel
	err := global.LRAG_DB.WithContext(ctx).Where("model_type = ?", req.ModelType).First(&existing).Error
	if err == nil {
		return global.LRAG_DB.WithContext(ctx).Model(&existing).Update("llm_provider_id", req.LLMProviderID).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Create(&rag.RagSystemDefaultModel{
		ModelType:     req.ModelType,
		LLMProviderID: req.LLMProviderID,
	}).Error
}

// ClearSystemDefault 清除系统全局默认模型
func (s *SystemModelService) ClearSystemDefault(ctx context.Context, modelType string) error {
	return global.LRAG_DB.WithContext(ctx).Where("model_type = ?", modelType).Delete(&rag.RagSystemDefaultModel{}).Error
}

// ================== 系统全局默认互联网搜索配置 ==================

// GetSystemWebSearchConfig 获取系统全局默认互联网搜索配置
func (s *SystemModelService) GetSystemWebSearchConfig(ctx context.Context) (provider string, config map[string]string, err error) {
	var cfg rag.RagSystemDefaultWebSearchConfig
	err = global.LRAG_DB.WithContext(ctx).First(&cfg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || cfg.Provider == "" {
		return "", nil, nil
	}
	configMap := make(map[string]string)
	if cfg.Config != nil {
		for k, v := range cfg.Config {
			if str, ok := v.(string); ok {
				configMap[k] = str
			}
		}
	}
	return cfg.Provider, configMap, nil
}

// SetSystemWebSearchConfig 设置系统全局默认互联网搜索配置
func (s *SystemModelService) SetSystemWebSearchConfig(ctx context.Context, req request.SystemWebSearchConfigSet) error {
	config := make(map[string]any)
	for k, v := range req.Config {
		config[k] = v
	}
	var existing rag.RagSystemDefaultWebSearchConfig
	err := global.LRAG_DB.WithContext(ctx).First(&existing).Error
	if err == nil {
		return global.LRAG_DB.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"provider": req.Provider,
			"config":   config,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Create(&rag.RagSystemDefaultWebSearchConfig{
		Provider: req.Provider,
		Config:   config,
	}).Error
}

// ClearSystemWebSearchConfig 清除系统全局默认互联网搜索配置
func (s *SystemModelService) ClearSystemWebSearchConfig(ctx context.Context) error {
	return global.LRAG_DB.WithContext(ctx).Where("1 = 1").Delete(&rag.RagSystemDefaultWebSearchConfig{}).Error
}

// ================== 模型回退解析 ==================

// ResolveModelWithFallback 完整回退链解析模型：KB配置 → 用户默认 → 角色默认 → 系统全局默认
// 返回 provider, modelName, baseURL, apiKey, ok
func ResolveModelWithFallback(ctx context.Context, uid uint, authorityId uint, kbModelID uint, kbModelSource string, modelType string) (provider, modelName, baseURL, apiKey string, ok bool) {
	// 1. KB 指定的模型
	if kbModelID > 0 {
		p, m, b, a, found := resolveByIDAndSource(ctx, uid, kbModelID, kbModelSource)
		if found {
			return p, m, b, a, true
		}
	}

	// 2. 用户默认模型
	var userDef rag.RagUserDefaultLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND model_type = ?", uid, modelType).First(&userDef).Error; err == nil {
		p, m, b, a, found := resolveByIDAndSource(ctx, uid, userDef.LLMProviderID, userDef.LLMSource)
		if found {
			return p, m, b, a, true
		}
	}

	// 3. 角色默认模型
	if authorityId > 0 {
		var authDef rag.RagAuthorityDefaultLLM
		if err := global.LRAG_DB.WithContext(ctx).Where("authority_id = ? AND model_type = ?", authorityId, modelType).First(&authDef).Error; err == nil {
			p, m, b, a, found := resolveByIDAndSource(ctx, uid, authDef.LLMProviderID, authDef.LLMSource)
			if found {
				return p, m, b, a, true
			}
		}
	}

	// 4. 系统全局默认模型
	var sysDef rag.RagSystemDefaultModel
	if err := global.LRAG_DB.WithContext(ctx).Where("model_type = ?", modelType).First(&sysDef).Error; err == nil {
		var m rag.RagLLMProvider
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND enabled = ?", sysDef.LLMProviderID, true).First(&m).Error; err == nil {
			return m.Name, m.ModelName, m.BaseURL, m.APIKey, true
		}
	}

	// 5. 全局搜索该类型的管理员模型
	var admins []rag.RagLLMProvider
	if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&admins).Error; err == nil {
		for _, a := range admins {
			if slices.Contains(a.ModelTypes, modelType) {
				return a.Name, a.ModelName, a.BaseURL, a.APIKey, true
			}
		}
	}

	// 6. 用户自定义模型
	if uid > 0 {
		var users []rag.RagUserLLM
		if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND enabled = ?", uid, true).Find(&users).Error; err == nil {
			for _, u := range users {
				if slices.Contains(u.ModelTypes, modelType) {
					return u.Provider, u.ModelName, u.BaseURL, u.APIKey, true
				}
			}
		}
	}

	return "", "", "", "", false
}

// resolveByIDAndSource 按 ID 和来源解析模型配置
func resolveByIDAndSource(ctx context.Context, uid uint, modelID uint, source string) (provider, modelName, baseURL, apiKey string, ok bool) {
	if modelID == 0 {
		return "", "", "", "", false
	}
	if source == "" {
		source = "admin"
	}
	if source == "user" {
		var m rag.RagUserLLM
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ? AND enabled = ?", modelID, uid, true).First(&m).Error; err == nil {
			return m.Provider, m.ModelName, m.BaseURL, m.APIKey, true
		}
	} else {
		var m rag.RagLLMProvider
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND enabled = ?", modelID, true).First(&m).Error; err == nil {
			return m.Name, m.ModelName, m.BaseURL, m.APIKey, true
		}
	}
	return "", "", "", "", false
}

// ResolveDefaultLLMWithSystem 完整回退链解析对话默认 LLM：用户默认 → 角色默认 → 系统全局默认
func ResolveDefaultLLMWithSystem(ctx context.Context, uid uint, authorityId uint, modelType string) (providerID uint, source string, ok bool) {
	if modelType == "" {
		modelType = interfaces.ModelTypeChat
	}
	// 1. 用户默认
	var userDef rag.RagUserDefaultLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND model_type = ?", uid, modelType).First(&userDef).Error; err == nil {
		return userDef.LLMProviderID, userDef.LLMSource, true
	}
	// 2. 角色默认
	if authorityId > 0 {
		var authDef rag.RagAuthorityDefaultLLM
		if err := global.LRAG_DB.WithContext(ctx).Where("authority_id = ? AND model_type = ?", authorityId, modelType).First(&authDef).Error; err == nil {
			return authDef.LLMProviderID, authDef.LLMSource, true
		}
	}
	// 3. 系统全局默认
	var sysDef rag.RagSystemDefaultModel
	if err := global.LRAG_DB.WithContext(ctx).Where("model_type = ?", modelType).First(&sysDef).Error; err == nil {
		return sysDef.LLMProviderID, "admin", true
	}
	return 0, "", false
}

// ================== 全局共享知识库 ==================

// ListGlobalKnowledgeBases 获取全局共享知识库列表（含知识库名称等信息）
func (s *SystemModelService) ListGlobalKnowledgeBases(ctx context.Context) ([]map[string]any, error) {
	var list []rag.RagGlobalKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(list))
	for _, g := range list {
		item := map[string]any{
			"id":              g.ID,
			"knowledgeBaseId": g.KnowledgeBaseID,
			"description":     g.Description,
		}
		var kb rag.RagKnowledgeBase
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", g.KnowledgeBaseID).First(&kb).Error; err == nil {
			item["knowledgeBaseName"] = kb.Name
			item["ownerName"] = ""
		}
		result = append(result, item)
	}
	return result, nil
}

// SetGlobalKnowledgeBase 添加一个知识库为全局共享
func (s *SystemModelService) SetGlobalKnowledgeBase(ctx context.Context, req request.GlobalKnowledgeBaseSet) error {
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", req.KnowledgeBaseID).First(&kb).Error; err != nil {
		return errors.New("知识库不存在")
	}
	gkb := &rag.RagGlobalKnowledgeBase{
		KnowledgeBaseID: req.KnowledgeBaseID,
		Description:     req.Description,
	}
	return global.LRAG_DB.WithContext(ctx).
		Where("knowledge_base_id = ?", req.KnowledgeBaseID).
		FirstOrCreate(gkb).Error
}

// RemoveGlobalKnowledgeBase 移除一个全局共享知识库
func (s *SystemModelService) RemoveGlobalKnowledgeBase(ctx context.Context, kbID uint) error {
	return global.LRAG_DB.WithContext(ctx).
		Where("knowledge_base_id = ?", kbID).
		Delete(&rag.RagGlobalKnowledgeBase{}).Error
}

// GetGlobalKnowledgeBaseIDs 获取全局知识库 ID 列表（供对话检索时使用）
func (s *SystemModelService) GetGlobalKnowledgeBaseIDs(ctx context.Context) ([]uint, error) {
	var list []rag.RagGlobalKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(list))
	for _, g := range list {
		ids = append(ids, g.KnowledgeBaseID)
	}
	return ids, nil
}

// ListAllKnowledgeBases 列出系统中所有知识库（供管理员选择设为全局共享）
func (s *SystemModelService) ListAllKnowledgeBases(ctx context.Context) ([]map[string]any, error) {
	var list []rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(list))
	for _, kb := range list {
		result = append(result, map[string]any{
			"id":   kb.ID,
			"name": kb.Name,
		})
	}
	return result, nil
}
