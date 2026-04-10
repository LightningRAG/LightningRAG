package initialize

import (
	"context"
	model "github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/LightningRAG/LightningRAG/server/plugin/plugin-tool/utils"
)

func Dictionary(ctx context.Context) {
	entities := []model.SysDictionary{}
	utils.RegisterDictionaries(entities...)
}
