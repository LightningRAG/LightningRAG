package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagEmbeddingProvider 嵌入模型提供商配置
type RagEmbeddingProvider struct {
	global.LRAG_MODEL
	Name       string         `json:"name" gorm:"size:64;comment:提供商名称"`
	ModelName  string         `json:"modelName" gorm:"size:128;comment:模型名称"`
	BaseURL    string         `json:"baseUrl" gorm:"size:256;comment:API Base URL"`
	APIKey     string         `json:"-" gorm:"size:512;comment:API Key"`
	Config     common.JSONMap `json:"config" gorm:"type:text;comment:额外配置"`
	Dimensions int            `json:"dimensions" gorm:"comment:向量维度"`
	Enabled    bool           `json:"enabled" gorm:"default:true"`
}

func (RagEmbeddingProvider) TableName() string {
	return "rag_embedding_providers"
}
