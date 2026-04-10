package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

type SysAutoCodePackage struct {
	global.LRAG_MODEL
	Desc        string `json:"desc" gorm:"comment:Description"`
	Label       string `json:"label" gorm:"comment:Display name"`
	Template    string `json:"template"  gorm:"comment:模版"`
	PackageName string `json:"packageName" gorm:"comment:包名"`
	Module      string `json:"-" example:"模块"`
}

func (s *SysAutoCodePackage) TableName() string {
	return "sys_auto_code_packages"
}
