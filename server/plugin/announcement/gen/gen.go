package main

import (
	"path/filepath"

	"github.com/LightningRAG/LightningRAG/server/plugin/announcement/model"
	"gorm.io/gen"
)

//go:generate go mod tidy
//go:generate go mod download
//go:generate go run gen.go

func main() {
	g := gen.NewGenerator(gen.Config{OutPath: filepath.Join("..", "..", "..", "announcement", "blender", "model", "dao"), Mode: gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface})
	g.ApplyBasic(
		new(model.Info),
	)
	g.Execute()
}
