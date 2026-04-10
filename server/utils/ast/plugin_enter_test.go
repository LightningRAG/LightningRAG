package ast

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginEnter_Injection(t *testing.T) {
	type fields struct {
		Type            Type
		Path            string
		ImportPath      string
		StructName      string
		StructCamelName string
		ModuleName      string
		GroupName       string
		PackageName     string
		ServiceName     string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 Lrag插件UserApi 注入",
			fields: fields{
				Type:            TypePluginApiEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "api", "enter.go"),
				ImportPath:      `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/service"`,
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "serviceUser",
				GroupName:       "Service",
				PackageName:     "service",
				ServiceName:     "User",
			},
			wantErr: false,
		},
		{
			name: "测试 Lrag插件UserRouter 注入",
			fields: fields{
				Type:            TypePluginRouterEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "router", "enter.go"),
				ImportPath:      `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/api"`,
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "userApi",
				GroupName:       "Api",
				PackageName:     "api",
				ServiceName:     "User",
			},
			wantErr: false,
		},
		{
			name: "测试 Lrag插件UserService 注入",
			fields: fields{
				Type:            TypePluginServiceEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "service", "enter.go"),
				ImportPath:      "",
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "",
				GroupName:       "",
				PackageName:     "",
				ServiceName:     "",
			},
			wantErr: false,
		},
		{
			name: "测试 lrag的User 注入",
			fields: fields{
				Type:            TypePluginServiceEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "service", "enter.go"),
				ImportPath:      "",
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "",
				GroupName:       "",
				PackageName:     "",
				ServiceName:     "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.fields.Path); err != nil {
				t.Skip("skip: plugin fixture not present:", tt.fields.Path)
			}
			a := &PluginEnter{
				Type:            tt.fields.Type,
				Path:            tt.fields.Path,
				ImportPath:      tt.fields.ImportPath,
				StructName:      tt.fields.StructName,
				StructCamelName: tt.fields.StructCamelName,
				ModuleName:      tt.fields.ModuleName,
				GroupName:       tt.fields.GroupName,
				PackageName:     tt.fields.PackageName,
				ServiceName:     tt.fields.ServiceName,
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

func TestPluginEnter_Rollback(t *testing.T) {
	type fields struct {
		Type            Type
		Path            string
		ImportPath      string
		StructName      string
		StructCamelName string
		ModuleName      string
		GroupName       string
		PackageName     string
		ServiceName     string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "测试 Lrag插件UserRouter 回滚",
			fields: fields{
				Type:            TypePluginRouterEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "router", "enter.go"),
				ImportPath:      `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/api"`,
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "userApi",
				GroupName:       "Api",
				PackageName:     "api",
				ServiceName:     "User",
			},
			wantErr: false,
		},
		{
			name: "测试 Lrag插件UserApi 回滚",
			fields: fields{
				Type:            TypePluginApiEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "api", "enter.go"),
				ImportPath:      `"github.com/LightningRAG/LightningRAG/server/plugin/lrag/service"`,
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "serviceUser",
				GroupName:       "Service",
				PackageName:     "service",
				ServiceName:     "User",
			},
			wantErr: false,
		},
		{
			name: "测试 Lrag插件UserService 回滚",
			fields: fields{
				Type:            TypePluginServiceEnter,
				Path:            filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "service", "enter.go"),
				ImportPath:      "",
				StructName:      "User",
				StructCamelName: "user",
				ModuleName:      "",
				GroupName:       "",
				PackageName:     "",
				ServiceName:     "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.fields.Path); err != nil {
				t.Skip("skip: plugin fixture not present:", tt.fields.Path)
			}
			a := &PluginEnter{
				Type:            tt.fields.Type,
				Path:            tt.fields.Path,
				ImportPath:      tt.fields.ImportPath,
				StructName:      tt.fields.StructName,
				StructCamelName: tt.fields.StructCamelName,
				ModuleName:      tt.fields.ModuleName,
				GroupName:       tt.fields.GroupName,
				PackageName:     tt.fields.PackageName,
				ServiceName:     tt.fields.ServiceName,
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
