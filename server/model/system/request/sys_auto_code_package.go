package request

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	model "github.com/LightningRAG/LightningRAG/server/model/system"
)

type SysAutoCodePackageCreate struct {
	Desc        string `json:"desc" example:"Description"`
	Label       string `json:"label" example:"Display name"`
	Template    string `json:"template"  example:"模版"`
	PackageName string `json:"packageName" example:"包名"`
	Module      string `json:"-" example:"模块"`
}

func (r *SysAutoCodePackageCreate) AutoCode() AutoCode {
	return AutoCode{
		Package: r.PackageName,
		Module:  global.LRAG_CONFIG.AutoCode.Module,
	}
}

func (r *SysAutoCodePackageCreate) Create() model.SysAutoCodePackage {
	return model.SysAutoCodePackage{
		Desc:        r.Desc,
		Label:       r.Label,
		Template:    r.Template,
		PackageName: r.PackageName,
		Module:      global.LRAG_CONFIG.AutoCode.Module,
	}
}
