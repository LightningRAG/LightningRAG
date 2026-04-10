package system

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
	model "github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
)

func init() {
	root, err := filepath.Abs("../../..")
	if err != nil {
		return
	}
	global.LRAG_CONFIG.AutoCode.Root = root
	global.LRAG_CONFIG.AutoCode.Server = "server"
}

func Test_autoCodePackage_Create(t *testing.T) {
	if global.LRAG_DB == nil {
		t.Skip("skip: global.LRAG_DB not initialized")
	}
	type args struct {
		ctx  context.Context
		info *request.SysAutoCodePackageCreate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "测试 package",
			args: args{
				ctx: context.Background(),
				info: &request.SysAutoCodePackageCreate{
					Template:    "package",
					PackageName: "lrag",
				},
			},
			wantErr: false,
		},
		{
			name: "测试 plugin",
			args: args{
				ctx: context.Background(),
				info: &request.SysAutoCodePackageCreate{
					Template:    "plugin",
					PackageName: "lrag",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &autoCodePackage{}
			if err := a.Create(tt.args.ctx, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_autoCodePackage_templates(t *testing.T) {
	type args struct {
		ctx       context.Context
		entity    model.SysAutoCodePackage
		info      request.AutoCode
		isPackage bool
	}
	tests := []struct {
		name      string
		args      args
		wantCode  map[string]string
		wantEnter map[string]map[string]string
		wantErr   bool
	}{
		{
			name: "测试1",
			args: args{
				ctx: context.Background(),
				entity: model.SysAutoCodePackage{
					Desc:        "Description",
					Label:       "Display name",
					Template:    "plugin",
					PackageName: "preview",
				},
				info: request.AutoCode{
					Abbreviation:    "user",
					HumpPackageName: "user",
				},
				isPackage: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &autoCodePackage{}
			gotCode, gotEnter, gotCreates, err := s.templates(tt.args.ctx, tt.args.entity, tt.args.info, tt.args.isPackage)
			if (err != nil) != tt.wantErr {
				t.Errorf("templates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for key, value := range gotCode {
				t.Log("\n")
				t.Log(key)
				t.Log(value)
				t.Log("\n")
			}
			t.Log(gotCreates)
			// nil 与空 map 视为等价（templates 可能返回其一）
			if !(len(gotEnter) == 0 && len(tt.wantEnter) == 0) && !reflect.DeepEqual(gotEnter, tt.wantEnter) {
				t.Errorf("templates() gotEnter = %v, want %v", gotEnter, tt.wantEnter)
			}
		})
	}
}
