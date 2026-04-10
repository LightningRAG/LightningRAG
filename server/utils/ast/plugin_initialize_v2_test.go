package ast

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"path/filepath"
	"testing"
)

func TestPluginInitialize_Injection(t *testing.T) {
	skipIfNoLRagPlugin(t)
	type fields struct {
		Type       Type
		Path       string
		PluginPath string
		ImportPath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 Lrag插件 注册注入",
			fields: fields{
				Type:       TypePluginInitializeV2,
				Path:       filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "plugin.go"),
				PluginPath: filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "register.go"),
				ImportPath: `"github.com/LightningRAG/LightningRAG/server/plugin/lrag"`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipUnlessFile(t, tt.fields.Path)
			skipUnlessFile(t, tt.fields.PluginPath)
			a := PluginInitializeV2{
				Type:       tt.fields.Type,
				Path:       tt.fields.Path,
				PluginPath: tt.fields.PluginPath,
				ImportPath: tt.fields.ImportPath,
			}
			file, err := a.Parse("", nil)
			if err != nil {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if file == nil {
				t.Fatal("Parse() returned nil file")
			}
			a.Injection(file)
			err = a.Format("", nil, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Injection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPluginInitialize_Rollback(t *testing.T) {
	skipIfNoLRagPlugin(t)
	type fields struct {
		Type        Type
		Path        string
		PluginPath  string
		ImportPath  string
		PluginName  string
		StructName  string
		PackageName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 Lrag插件 回滚",
			fields: fields{
				Type:       TypePluginInitializeV2,
				Path:       filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "plugin.go"),
				PluginPath: filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "register.go"),
				ImportPath: `"github.com/LightningRAG/LightningRAG/server/plugin/lrag"`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipUnlessFile(t, tt.fields.Path)
			skipUnlessFile(t, tt.fields.PluginPath)
			a := PluginInitializeV2{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				PluginPath:  tt.fields.PluginPath,
				ImportPath:  tt.fields.ImportPath,
				StructName:  "Plugin",
				PackageName: "lrag",
			}
			file, err := a.Parse("", nil)
			if err != nil {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if file == nil {
				t.Fatal("Parse() returned nil file")
			}
			a.Rollback(file)
			err = a.Format("", nil, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
