package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
)

// RagSystemDefaultWebSearchConfig 系统全局默认互联网搜索配置（管理员设置）
// 回退链：用户自定义配置 → 系统全局默认 → DuckDuckGo
type RagSystemDefaultWebSearchConfig struct {
	global.LRAG_MODEL
	Provider string         `json:"provider" gorm:"size:32;comment:搜索引擎 duckduckgo|baidu"`
	Config   common.JSONMap `json:"config" gorm:"type:text;comment:引擎配置 JSON 如 {\"apiKey\":\"xxx\"}"`
}

func (RagSystemDefaultWebSearchConfig) TableName() string {
	return "rag_system_default_web_search_configs"
}
