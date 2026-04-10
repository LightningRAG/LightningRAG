package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagFileStorageConfig 知识库文件存储配置
// 支持 local, qiniu, tencent-cos, aliyun-oss, huawei-obs, aws-s3, cloudflare-r2, minio
type RagFileStorageConfig struct {
	global.LRAG_MODEL
	Name                  string         `json:"name" gorm:"size:64;comment:配置名称"`
	Provider              string         `json:"provider" gorm:"size:32;comment:存储类型 local|qiniu|tencent-cos|aliyun-oss|minio等"`
	Config                common.JSONMap `json:"config" gorm:"type:text;comment:连接配置 JSON"`
	Enabled               bool           `json:"enabled" gorm:"default:true"`
	AllowAll              bool           `json:"allowAll" gorm:"default:true;comment:是否所有角色可在创建知识库时选用"`
	AllowedAuthorityIDs   []uint         `json:"allowedAuthorityIds" gorm:"serializer:json;type:text;comment:allow_all为false时允许的角色ID"`
}

func (RagFileStorageConfig) TableName() string {
	return "rag_file_storage_configs"
}
