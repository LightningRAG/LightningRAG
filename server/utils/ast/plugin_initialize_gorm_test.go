package ast

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"path/filepath"
	"testing"
)

func TestPluginInitializeGorm_Injection(t *testing.T) {
	skipIfNoLRagPlugin(t)
	type fields struct {
		Type        Type
		Path        string
		ImportPath  string
		StructName  string
		PackageName string
		IsNew       bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 &model.User{} 注入",
			fields: fields{
				Type:        TypePluginInitializeGorm,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "initialize", "gorm.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				StructName:  "User",
				PackageName: "model",
				IsNew:       false,
			},
		},
		{
			name: "测试 new(model.ExaCustomer) 注入",
			fields: fields{
				Type:        TypePluginInitializeGorm,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "initialize", "gorm.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				StructName:  "User",
				PackageName: "model",
				IsNew:       true,
			},
		},
		{
			name: "测试 new(model.SysUsers) 注入",
			fields: fields{
				Type:        TypePluginInitializeGorm,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "initialize", "gorm.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				StructName:  "SysUser",
				PackageName: "model",
				IsNew:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipUnlessFile(t, tt.fields.Path)
			a := &PluginInitializeGorm{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				ImportPath:  tt.fields.ImportPath,
				StructName:  tt.fields.StructName,
				PackageName: tt.fields.PackageName,
				IsNew:       tt.fields.IsNew,
			}
			file, err := a.Parse(a.Path, nil)
			if err != nil {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if file == nil {
				t.Fatal("Parse() returned nil file")
			}
			a.Injection(file)
			err = a.Format(a.Path, nil, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Injection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPluginInitializeGorm_Rollback(t *testing.T) {
	skipIfNoLRagPlugin(t)
	type fields struct {
		Type        Type
		Path        string
		ImportPath  string
		StructName  string
		PackageName string
		IsNew       bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 &model.User{} 回滚",
			fields: fields{
				Type:        TypePluginInitializeGorm,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "initialize", "gorm.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				StructName:  "User",
				PackageName: "model",
				IsNew:       false,
			},
		},
		{
			name: "测试 new(model.ExaCustomer) 回滚",
			fields: fields{
				Type:        TypePluginInitializeGorm,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "initialize", "gorm.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				StructName:  "User",
				PackageName: "model",
				IsNew:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipUnlessFile(t, tt.fields.Path)
			a := &PluginInitializeGorm{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				ImportPath:  tt.fields.ImportPath,
				StructName:  tt.fields.StructName,
				PackageName: tt.fields.PackageName,
				IsNew:       tt.fields.IsNew,
			}
			file, err := a.Parse(a.Path, nil)
			if err != nil {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if file == nil {
				t.Fatal("Parse() returned nil file")
			}
			a.Rollback(file)
			err = a.Format(a.Path, nil, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
