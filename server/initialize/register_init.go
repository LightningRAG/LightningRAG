package initialize

import (
	_ "github.com/LightningRAG/LightningRAG/server/source/example"
	_ "github.com/LightningRAG/LightningRAG/server/source/system"
)

func init() {
	// do nothing,only import source package so that inits can be registered
}
