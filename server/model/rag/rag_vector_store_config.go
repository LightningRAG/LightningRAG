package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagVectorStoreConfig 向量存储配置
type RagVectorStoreConfig struct {
	global.LRAG_MODEL
	Name                  string         `json:"name" gorm:"size:64;comment:配置名称"`
	Provider              string         `json:"provider" gorm:"size:32;comment:存储类型 postgresql|elasticsearch"`
	Config                common.JSONMap `json:"config" gorm:"type:text;comment:连接配置 JSON"`
	Enabled               bool           `json:"enabled" gorm:"default:true"`
	AllowAll              bool           `json:"allowAll" gorm:"default:true;comment:是否所有角色可在创建知识库时选用"`
	AllowedAuthorityIDs   []uint         `json:"allowedAuthorityIds" gorm:"serializer:json;type:text;comment:allow_all为false时允许的角色ID"`
}

func (RagVectorStoreConfig) TableName() string {
	return "rag_vector_store_configs"
}
