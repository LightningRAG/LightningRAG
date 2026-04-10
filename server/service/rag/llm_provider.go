package rag

import (
	"context"
	"errors"
	"slices"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"gorm.io/gorm"
)

// matchScenario 检查模型的 modelTypes 是否包含指定场景，空类型视为匹配 chat
func matchScenario(types []string, scenario string) bool {
	if scenario == "" {
		return true
	}
	if len(types) == 0 {
		return scenario == "chat" // 兼容旧数据，默认视为 chat
	}
	return slices.Contains(types, scenario)
}

// ListProviders 列出用户可用的 LLM（用户自定义的优先），可按 scenarioType 过滤，并标记默认模型
// scenarioType 为空时返回全部类型的模型
func (s *LLMProviderService) ListProviders(ctx context.Context, uid uint, authorityId uint, scenarioType string) ([]any, error) {
	var providers []rag.RagLLMProvider
	if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&providers).Error; err != nil {
		return nil, err
	}
	var userLLMs []rag.RagUserLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND enabled = ?", uid, true).Find(&userLLMs).Error; err != nil {
		return nil, err
	}
	// 获取默认模型（用户默认 > 角色默认）
	defID, defSrc, _, hasDef := s.ResolveDefaultLLM(ctx, uid, authorityId, scenarioType)
	result := make([]any, 0, len(providers)+len(userLLMs))
	for _, u := range userLLMs {
		if !matchScenario(u.ModelTypes, scenarioType) {
			continue
		}
		isDefault := hasDef && defSrc == "user" && defID == u.ID
		result = append(result, map[string]any{
			"id":                   u.ID,
			"name":                 u.Provider,
			"modelName":            u.ModelName,
			"modelTypes":           u.ModelTypes,
			"source":               "user",
			"isDefault":            isDefault,
			"supportsDeepThinking": u.SupportsDeepThinking,
			"supportsToolCall":     u.SupportsToolCall,
		})
	}
	for _, p := range providers {
		if !matchScenario(p.ModelTypes, scenarioType) {
			continue
		}
		isDefault := hasDef && defSrc == "admin" && defID == p.ID
		result = append(result, map[string]any{
			"id":                   p.ID,
			"name":                 p.Name,
			"modelName":            p.ModelName,
			"modelTypes":           p.ModelTypes,
			"source":               "admin",
			"isDefault":            isDefault,
			"supportsDeepThinking": p.SupportsDeepThinking,
			"supportsToolCall":     p.SupportsToolCall,
		})
	}
	return result, nil
}

// ListUserModels 列出用户自定义模型
func (s *LLMProviderService) ListUserModels(ctx context.Context, uid uint) ([]rag.RagUserLLM, error) {
	var list []rag.RagUserLLM
	err := global.LRAG_DB.WithContext(ctx).Where("user_id = ?", uid).Find(&list).Error
	return list, err
}

// AddUserModel 添加用户自定义模型
func (s *LLMProviderService) AddUserModel(ctx context.Context, uid uint, req request.LLMProviderAddUser) error {
	modelTypes := req.ModelTypes
	if len(modelTypes) == 0 {
		modelTypes = []string{"chat"} // 默认对话场景
	}
	userLLM := &rag.RagUserLLM{
		UserID:               uid,
		Provider:             req.Provider,
		ModelName:            req.ModelName,
		ModelTypes:           modelTypes,
		BaseURL:              req.BaseURL,
		APIKey:               req.APIKey,
		MaxContextTokens:     req.MaxContextTokens,
		SupportsDeepThinking: req.SupportsDeepThinking,
		SupportsToolCall:     req.SupportsToolCall,
		Enabled:              true,
	}
	return global.LRAG_DB.WithContext(ctx).Create(userLLM).Error
}

// UpdateUserModel 更新用户自定义模型
func (s *LLMProviderService) UpdateUserModel(ctx context.Context, uid uint, req request.LLMProviderUpdateUser) error {
	modelTypes := req.ModelTypes
	if len(modelTypes) == 0 {
		modelTypes = []string{"chat"}
	}
	q := global.LRAG_DB.WithContext(ctx).Model(&rag.RagUserLLM{}).Where("id = ? AND user_id = ?", req.ID, uid)
	if err := q.
		Select("Provider", "ModelName", "ModelTypes", "BaseURL", "MaxContextTokens", "SupportsDeepThinking", "SupportsToolCall").
		Updates(rag.RagUserLLM{
			Provider:             req.Provider,
			ModelName:            req.ModelName,
			ModelTypes:           modelTypes,
			BaseURL:              req.BaseURL,
			MaxContextTokens:     req.MaxContextTokens,
			SupportsDeepThinking: req.SupportsDeepThinking,
			SupportsToolCall:     req.SupportsToolCall,
		}).Error; err != nil {
		return err
	}
	if req.APIKey != "" {
		return global.LRAG_DB.WithContext(ctx).Model(&rag.RagUserLLM{}).
			Where("id = ? AND user_id = ?", req.ID, uid).
			Update("api_key", req.APIKey).Error
	}
	return nil
}

// DeleteUserModel 删除用户自定义模型
func (s *LLMProviderService) DeleteUserModel(ctx context.Context, uid uint, id uint) error {
	return global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, uid).Delete(&rag.RagUserLLM{}).Error
}

// SetAuthorityDefaultLLM 设置角色默认模型（管理员为角色设置）
func (s *LLMProviderService) SetAuthorityDefaultLLM(ctx context.Context, req request.SetAuthorityDefaultLLMReq) error {
	source := req.LLMSource
	if source == "" {
		source = "admin"
	}
	var existing rag.RagAuthorityDefaultLLM
	err := global.LRAG_DB.WithContext(ctx).Where("authority_id = ? AND model_type = ?", req.AuthorityId, req.ModelType).First(&existing).Error
	if err == nil {
		return global.LRAG_DB.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"llm_provider_id": req.LLMProviderID,
			"llm_source":      source,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Create(&rag.RagAuthorityDefaultLLM{
		AuthorityId:   req.AuthorityId,
		ModelType:     req.ModelType,
		LLMProviderID: req.LLMProviderID,
		LLMSource:     source,
	}).Error
}

// GetAuthorityDefaultLLMs 获取角色各类型默认模型
func (s *LLMProviderService) GetAuthorityDefaultLLMs(ctx context.Context, authorityId uint) ([]map[string]any, error) {
	var list []rag.RagAuthorityDefaultLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("authority_id = ?", authorityId).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(list))
	for _, d := range list {
		result = append(result, map[string]any{
			"modelType":     d.ModelType,
			"llmProviderId": d.LLMProviderID,
			"llmSource":     d.LLMSource,
		})
	}
	return result, nil
}

// SetUserDefaultLLM 设置用户默认模型（用户为自己设置）
func (s *LLMProviderService) SetUserDefaultLLM(ctx context.Context, uid uint, req request.SetUserDefaultLLMReq) error {
	source := req.LLMSource
	if source == "" {
		source = "user"
	}
	var existing rag.RagUserDefaultLLM
	err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND model_type = ?", uid, req.ModelType).First(&existing).Error
	if err == nil {
		return global.LRAG_DB.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"llm_provider_id": req.LLMProviderID,
			"llm_source":      source,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Create(&rag.RagUserDefaultLLM{
		UserID:        uid,
		ModelType:     req.ModelType,
		LLMProviderID: req.LLMProviderID,
		LLMSource:     source,
	}).Error
}

// GetUserDefaultLLMs 获取用户各类型默认模型
func (s *LLMProviderService) GetUserDefaultLLMs(ctx context.Context, uid uint) ([]map[string]any, error) {
	var list []rag.RagUserDefaultLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ?", uid).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(list))
	for _, d := range list {
		result = append(result, map[string]any{
			"modelType":     d.ModelType,
			"llmProviderId": d.LLMProviderID,
			"llmSource":     d.LLMSource,
		})
	}
	return result, nil
}

// ClearAuthorityDefaultLLM 清除角色某类型默认模型
func (s *LLMProviderService) ClearAuthorityDefaultLLM(ctx context.Context, authorityId uint, modelType string) error {
	return global.LRAG_DB.WithContext(ctx).Where("authority_id = ? AND model_type = ?", authorityId, modelType).Delete(&rag.RagAuthorityDefaultLLM{}).Error
}

// ClearUserDefaultLLM 清除用户某类型默认模型
func (s *LLMProviderService) ClearUserDefaultLLM(ctx context.Context, uid uint, modelType string) error {
	return global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND model_type = ?", uid, modelType).Delete(&rag.RagUserDefaultLLM{}).Error
}

// GetWebSearchConfig 获取用户互联网搜索配置
// 回退链：用户自定义（UseSystemDefault=false） → 系统全局默认 → DuckDuckGo
func (s *LLMProviderService) GetWebSearchConfig(ctx context.Context, uid uint) (provider string, config map[string]string, useSystemDefault bool, err error) {
	var cfg rag.RagUserWebSearchConfig
	err = global.LRAG_DB.WithContext(ctx).Where("user_id = ?", uid).First(&cfg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, false, err
	}

	// 未配置或 UseSystemDefault=true 时，回退到系统默认
	if errors.Is(err, gorm.ErrRecordNotFound) || cfg.UseSystemDefault {
		sysProv, sysCfg, sysErr := systemModelService.GetSystemWebSearchConfig(ctx)
		if sysErr != nil {
			return "", nil, true, sysErr
		}
		if sysProv != "" {
			return sysProv, sysCfg, true, nil
		}
		return "duckduckgo", nil, true, nil
	}

	if cfg.Provider == "" {
		return "duckduckgo", nil, false, nil
	}
	configMap := make(map[string]string)
	if cfg.Config != nil {
		for k, v := range cfg.Config {
			if str, ok := v.(string); ok {
				configMap[k] = str
			}
		}
	}
	return cfg.Provider, configMap, false, nil
}

// SetWebSearchConfig 设置用户互联网搜索配置
func (s *LLMProviderService) SetWebSearchConfig(ctx context.Context, uid uint, req request.WebSearchConfigSetReq) error {
	useSystemDefault := true
	if req.UseSystemDefault != nil {
		useSystemDefault = *req.UseSystemDefault
	}

	config := make(map[string]any)
	for k, v := range req.Config {
		config[k] = v
	}

	var existing rag.RagUserWebSearchConfig
	err := global.LRAG_DB.WithContext(ctx).Where("user_id = ?", uid).First(&existing).Error
	if err == nil {
		return global.LRAG_DB.WithContext(ctx).Model(&existing).
			Select("UseSystemDefault", "Provider", "Config").
			Updates(rag.RagUserWebSearchConfig{
				UseSystemDefault: useSystemDefault,
				Provider:         req.Provider,
				Config:           config,
			}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Create(&rag.RagUserWebSearchConfig{
		UserID:           uid,
		UseSystemDefault: useSystemDefault,
		Provider:         req.Provider,
		Config:           config,
	}).Error
}

// ResolveDefaultLLM 解析默认模型：用户默认 → 角色默认 → 系统全局默认
func (s *LLMProviderService) ResolveDefaultLLM(ctx context.Context, uid uint, authorityId uint, modelType string) (providerID uint, source string, isUserDefault bool, ok bool) {
	if modelType == "" {
		modelType = "chat"
	}
	// 1. 用户默认
	var userDef rag.RagUserDefaultLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND model_type = ?", uid, modelType).First(&userDef).Error; err == nil {
		return userDef.LLMProviderID, userDef.LLMSource, true, true
	}
	// 2. 角色默认
	var authDef rag.RagAuthorityDefaultLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("authority_id = ? AND model_type = ?", authorityId, modelType).First(&authDef).Error; err == nil {
		return authDef.LLMProviderID, authDef.LLMSource, false, true
	}
	// 3. 系统全局默认
	var sysDef rag.RagSystemDefaultModel
	if err := global.LRAG_DB.WithContext(ctx).Where("model_type = ?", modelType).First(&sysDef).Error; err == nil {
		return sysDef.LLMProviderID, "admin", false, true
	}
	return 0, "", false, false
}
