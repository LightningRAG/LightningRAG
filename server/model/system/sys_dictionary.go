// 自动生成模板SysDictionary
package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// 如果含有time.Time 请自行import time包
type SysDictionary struct {
	global.LRAG_MODEL
	Name                 string                `json:"name" form:"name" gorm:"column:name;comment:Display name"`        // Display name
	Type                 string                `json:"type" form:"type" gorm:"column:type;comment:字典名（英）"`              // 字典名（英）
	Status               *bool                 `json:"status" form:"status" gorm:"column:status;comment:状态"`            // 状态
	Desc                 string                `json:"desc" form:"desc" gorm:"column:desc;comment:Description"`         // Description
	ParentID             *uint                 `json:"parentID" form:"parentID" gorm:"column:parent_id;comment:父级字典ID"` // 父级字典ID
	Children             []SysDictionary       `json:"children" gorm:"foreignKey:ParentID"`                             // 子字典
	SysDictionaryDetails []SysDictionaryDetail `json:"sysDictionaryDetails" form:"sysDictionaryDetails"`
}

func (SysDictionary) TableName() string {
	return "sys_dictionaries"
}
