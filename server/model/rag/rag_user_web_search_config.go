package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagUserWebSearchConfig 用户互联网搜索配置（默认搜索引擎及配置项）
// UseSystemDefault 为 true 时忽略 Provider/Config，直接使用系统默认配置
type RagUserWebSearchConfig struct {
	global.LRAG_MODEL
	UserID           uint           `json:"userId" gorm:"uniqueIndex:idx_user_web_search;comment:用户ID"`
	UseSystemDefault bool           `json:"useSystemDefault" gorm:"default:true;comment:是否使用系统默认互联网搜索配置"`
	Provider         string         `json:"provider" gorm:"size:32;comment:搜索引擎 duckduckgo|baidu"`
	Config           common.JSONMap `json:"config" gorm:"type:text;comment:引擎配置 JSON 如 {\"apiKey\":\"xxx\"}"`
}

func (RagUserWebSearchConfig) TableName() string {
	return "rag_user_web_search_configs"
}
