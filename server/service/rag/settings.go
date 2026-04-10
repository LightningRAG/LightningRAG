package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
)

// superAdminAuthorityID 超级管理员角色，可选用全部向量/文件存储配置
const superAdminAuthorityID = uint(888)

func ragSettingUsableByAuthority(allowAll bool, allowed []uint, authorityID uint) bool {
	if authorityID == superAdminAuthorityID {
		return true
	}
	if allowAll {
		return true
	}
	return slices.Contains(allowed, authorityID)
}

func vectorStoreUsableByAuthority(cfg *rag.RagVectorStoreConfig, authorityID uint) bool {
	return ragSettingUsableByAuthority(cfg.AllowAll, cfg.AllowedAuthorityIDs, authorityID)
}

// ========== 向量存储配置 CRUD ==========

func (s *KnowledgeBaseService) CreateVectorStoreConfig(ctx context.Context, req request.VectorStoreConfigCreate) (*rag.RagVectorStoreConfig, error) {
	if !isSupportedVectorStoreProvider(req.Provider) {
		return nil, fmt.Errorf("不支持的向量存储类型: %s，支持: postgresql, elasticsearch", req.Provider)
	}
	allowAll := true
	if req.AllowAll != nil {
		allowAll = *req.AllowAll
	}
	ids := req.AllowedAuthorityIDs
	if ids == nil {
		ids = []uint{}
	}
	cfg := &rag.RagVectorStoreConfig{
		Name:                  req.Name,
		Provider:              req.Provider,
		Config:                common.JSONMap(req.Config),
		Enabled:               req.Enabled,
		AllowAll:              allowAll,
		AllowedAuthorityIDs:   ids,
	}
	if cfg.Config == nil {
		cfg.Config = common.JSONMap{}
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *KnowledgeBaseService) UpdateVectorStoreConfig(ctx context.Context, req request.VectorStoreConfigUpdate) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		if !isSupportedVectorStoreProvider(req.Provider) {
			return fmt.Errorf("不支持的向量存储类型: %s", req.Provider)
		}
		updates["provider"] = req.Provider
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.AllowAll != nil {
		updates["allow_all"] = *req.AllowAll
	}
	if req.AllowedAuthorityIDs != nil {
		b, err := json.Marshal(*req.AllowedAuthorityIDs)
		if err != nil {
			return fmt.Errorf("allowedAuthorityIds: %w", err)
		}
		updates["allowed_authority_ids"] = b
	}
	if len(updates) == 0 {
		return nil
	}
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagVectorStoreConfig{}).
		Where("id = ?", req.ID).Updates(updates).Error
}

// EnsureVectorStoreUsableByAuthority 校验当前角色是否可选用该向量存储（创建知识库等）
func (s *KnowledgeBaseService) EnsureVectorStoreUsableByAuthority(ctx context.Context, authorityID, vectorStoreID uint) error {
	if vectorStoreID == 0 {
		return fmt.Errorf("请选择有效的向量存储")
	}
	var cfg rag.RagVectorStoreConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND enabled = ?", vectorStoreID, true).First(&cfg).Error; err != nil {
		return fmt.Errorf("向量存储不可用或不存在")
	}
	if !vectorStoreUsableByAuthority(&cfg, authorityID) {
		return fmt.Errorf("当前角色无权使用该向量存储")
	}
	return nil
}

func (s *KnowledgeBaseService) DeleteVectorStoreConfig(ctx context.Context, id uint) error {
	var count int64
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagKnowledgeBase{}).
		Where("vector_store_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该向量存储已被 %d 个知识库使用，无法删除", count)
	}
	return global.LRAG_DB.WithContext(ctx).Delete(&rag.RagVectorStoreConfig{}, id).Error
}

func (s *KnowledgeBaseService) GetVectorStoreConfig(ctx context.Context, id uint) (*rag.RagVectorStoreConfig, error) {
	var cfg rag.RagVectorStoreConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *KnowledgeBaseService) ListVectorStoreConfigsFull(ctx context.Context, req request.VectorStoreConfigList) ([]rag.RagVectorStoreConfig, int64, error) {
	var list []rag.RagVectorStoreConfig
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagVectorStoreConfig{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Scopes(req.Paginate()).Find(&list).Error
	return list, total, err
}

func isSupportedVectorStoreProvider(p string) bool {
	for _, name := range registry.ListVectorStoreProviders() {
		if name == p {
			return true
		}
	}
	return false
}

// ========== 文件存储配置 CRUD ==========

var supportedFileStorageProviders = []string{"local", "qiniu", "tencent-cos", "aliyun-oss", "huawei-obs", "aws-s3", "cloudflare-r2", "minio"}

func fileStorageUsableByAuthority(cfg *rag.RagFileStorageConfig, authorityID uint) bool {
	return ragSettingUsableByAuthority(cfg.AllowAll, cfg.AllowedAuthorityIDs, authorityID)
}

// EnsureFileStorageUsableByAuthority 校验当前角色是否可选用该文件存储（创建知识库等）
func (s *KnowledgeBaseService) EnsureFileStorageUsableByAuthority(ctx context.Context, authorityID, fileStorageID uint) error {
	if fileStorageID == 0 {
		return fmt.Errorf("请选择有效的文件存储")
	}
	var cfg rag.RagFileStorageConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND enabled = ?", fileStorageID, true).First(&cfg).Error; err != nil {
		return fmt.Errorf("文件存储不可用或不存在")
	}
	if !fileStorageUsableByAuthority(&cfg, authorityID) {
		return fmt.Errorf("当前角色无权使用该文件存储")
	}
	return nil
}

func (s *KnowledgeBaseService) CreateFileStorageConfig(ctx context.Context, req request.FileStorageConfigCreate) (*rag.RagFileStorageConfig, error) {
	if !isSupportedFileStorageProvider(req.Provider) {
		return nil, fmt.Errorf("不支持的文件存储类型: %s", req.Provider)
	}
	allowAll := true
	if req.AllowAll != nil {
		allowAll = *req.AllowAll
	}
	ids := req.AllowedAuthorityIDs
	if ids == nil {
		ids = []uint{}
	}
	cfg := &rag.RagFileStorageConfig{
		Name:                  req.Name,
		Provider:              req.Provider,
		Config:                common.JSONMap(req.Config),
		Enabled:               req.Enabled,
		AllowAll:              allowAll,
		AllowedAuthorityIDs:   ids,
	}
	if cfg.Config == nil {
		cfg.Config = common.JSONMap{}
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *KnowledgeBaseService) UpdateFileStorageConfig(ctx context.Context, req request.FileStorageConfigUpdate) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		if !isSupportedFileStorageProvider(req.Provider) {
			return fmt.Errorf("不支持的文件存储类型: %s", req.Provider)
		}
		updates["provider"] = req.Provider
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.AllowAll != nil {
		updates["allow_all"] = *req.AllowAll
	}
	if req.AllowedAuthorityIDs != nil {
		b, err := json.Marshal(*req.AllowedAuthorityIDs)
		if err != nil {
			return fmt.Errorf("allowedAuthorityIds: %w", err)
		}
		updates["allowed_authority_ids"] = b
	}
	if len(updates) == 0 {
		return nil
	}
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagFileStorageConfig{}).
		Where("id = ?", req.ID).Updates(updates).Error
}

func (s *KnowledgeBaseService) DeleteFileStorageConfig(ctx context.Context, id uint) error {
	var count int64
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagKnowledgeBase{}).
		Where("file_storage_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该文件存储已被 %d 个知识库使用，无法删除", count)
	}
	return global.LRAG_DB.WithContext(ctx).Delete(&rag.RagFileStorageConfig{}, id).Error
}

func (s *KnowledgeBaseService) GetFileStorageConfig(ctx context.Context, id uint) (*rag.RagFileStorageConfig, error) {
	var cfg rag.RagFileStorageConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *KnowledgeBaseService) ListFileStorageConfigsFull(ctx context.Context, req request.FileStorageConfigList) ([]rag.RagFileStorageConfig, int64, error) {
	var list []rag.RagFileStorageConfig
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagFileStorageConfig{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Scopes(req.Paginate()).Find(&list).Error
	return list, total, err
}

func isSupportedFileStorageProvider(p string) bool {
	for _, name := range supportedFileStorageProviders {
		if name == p {
			return true
		}
	}
	return false
}

// ListFileStorageConfigs 列出当前角色可选用的已启用文件存储（供创建知识库时选择）
func (s *KnowledgeBaseService) ListFileStorageConfigs(ctx context.Context, authorityID uint) ([]map[string]any, error) {
	var list []rag.RagFileStorageConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		// 插入默认配置（使用系统默认）
		defaultCfg := &rag.RagFileStorageConfig{
			Name:                "默认 (系统配置)",
			Provider:            "local",
			Enabled:             true,
			AllowAll:            true,
			AllowedAuthorityIDs: []uint{},
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(defaultCfg).Error; err != nil {
			return nil, err
		}
		list = []rag.RagFileStorageConfig{*defaultCfg}
	}
	result := make([]map[string]any, 0, len(list))
	for _, c := range list {
		if !fileStorageUsableByAuthority(&c, authorityID) {
			continue
		}
		label := c.Name
		if c.Provider != "" {
			label += " (" + c.Provider + ")"
		}
		result = append(result, map[string]any{
			"id":       c.ID,
			"label":    label,
			"name":     c.Name,
			"provider": c.Provider,
		})
	}
	return result, nil
}
