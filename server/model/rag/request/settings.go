package request

import (
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
)

// VectorStoreConfigCreate 创建向量存储配置
type VectorStoreConfigCreate struct {
	Name                  string         `json:"name" binding:"required"`
	Provider              string         `json:"provider" binding:"required"` // postgresql | elasticsearch
	Config                map[string]any `json:"config"`
	Enabled               bool           `json:"enabled"`
	AllowAll              *bool          `json:"allowAll"`
	AllowedAuthorityIDs   []uint         `json:"allowedAuthorityIds"`
}

// VectorStoreConfigUpdate 更新向量存储配置
type VectorStoreConfigUpdate struct {
	ID                    uint           `json:"id" binding:"required"`
	Name                  string         `json:"name"`
	Provider              string         `json:"provider"`
	Config                map[string]any `json:"config"`
	Enabled               *bool          `json:"enabled"`
	AllowAll              *bool          `json:"allowAll"`
	AllowedAuthorityIDs   *[]uint        `json:"allowedAuthorityIds"`
}

// VectorStoreConfigList 向量存储配置列表（管理用，含分页）
type VectorStoreConfigList struct {
	request.PageInfo
}

// FileStorageConfigCreate 创建文件存储配置
type FileStorageConfigCreate struct {
	Name                  string         `json:"name" binding:"required"`
	Provider              string         `json:"provider" binding:"required"` // local | qiniu | tencent-cos | aliyun-oss | minio 等
	Config                map[string]any `json:"config"`
	Enabled               bool           `json:"enabled"`
	AllowAll              *bool          `json:"allowAll"`              // 默认 true（省略时）
	AllowedAuthorityIDs   []uint         `json:"allowedAuthorityIds"`   // allow_all 为 false 时生效
}

// FileStorageConfigUpdate 更新文件存储配置
type FileStorageConfigUpdate struct {
	ID                    uint           `json:"id" binding:"required"`
	Name                  string         `json:"name"`
	Provider              string         `json:"provider"`
	Config                map[string]any `json:"config"`
	Enabled               *bool          `json:"enabled"`
	AllowAll              *bool          `json:"allowAll"`
	AllowedAuthorityIDs   *[]uint        `json:"allowedAuthorityIds"`
}

// FileStorageConfigList 文件存储配置列表（管理用，含分页）
type FileStorageConfigList struct {
	request.PageInfo
}

// GlobalKnowledgeBaseSet 设置全局知识库
type GlobalKnowledgeBaseSet struct {
	KnowledgeBaseID uint   `json:"knowledgeBaseId" binding:"required"`
	Description     string `json:"description"`
}

// GlobalKnowledgeBaseRemove 移除全局知识库
type GlobalKnowledgeBaseRemove struct {
	KnowledgeBaseID uint `json:"knowledgeBaseId" binding:"required"`
}
