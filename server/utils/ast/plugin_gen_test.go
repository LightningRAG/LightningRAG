package ast

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"path/filepath"
	"testing"
)

func TestPluginGenModel_Injection(t *testing.T) {
	skipIfNoLRagPlugin(t)
	type fields struct {
		Type        Type
		Path        string
		ImportPath  string
		PackageName string
		StructName  string
		IsNew       bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 LragUser 结构体注入",
			fields: fields{
				Type:        TypePluginGen,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "gen", "main.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				PackageName: "model",
				StructName:  "User",
				IsNew:       false,
			},
		},
		{
			name: "测试 LragUser 结构体注入",
			fields: fields{
				Type:        TypePluginGen,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "gen", "main.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				PackageName: "model",
				StructName:  "User",
				IsNew:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PluginGen{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				ImportPath:  tt.fields.ImportPath,
				PackageName: tt.fields.PackageName,
				StructName:  tt.fields.StructName,
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

func TestPluginGenModel_Rollback(t *testing.T) {
	skipIfNoLRagPlugin(t)
	type fields struct {
		Type        Type
		Path        string
		ImportPath  string
		PackageName string
		StructName  string
		IsNew       bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 LragUser 回滚",
			fields: fields{
				Type:        TypePluginGen,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "gen", "main.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				PackageName: "model",
				StructName:  "User",
				IsNew:       false,
			},
		},
		{
			name: "测试 LragUser 回滚",
			fields: fields{
				Type:        TypePluginGen,
				Path:        filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "gen", "main.go"),
				ImportPath:  `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/model"`,
				PackageName: "model",
				StructName:  "User",
				IsNew:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PluginGen{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				ImportPath:  tt.fields.ImportPath,
				PackageName: tt.fields.PackageName,
				StructName:  tt.fields.StructName,
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
