package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// RagAgent Agent 流程编排定义
type RagAgent struct {
	global.LRAG_MODEL
	OwnerID uint   `json:"ownerId" gorm:"index;comment:创建者用户ID"`
	Name    string `json:"name" gorm:"size:128;comment:Agent 名称"`
	Desc    string `json:"desc" gorm:"size:512;comment:描述"`
	DSL     string `json:"dsl" gorm:"type:longtext;comment:DSL JSON 工作流定义"`
}

func (RagAgent) TableName() string {
	return "rag_agents"
}
