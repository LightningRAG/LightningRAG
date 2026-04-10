package ast

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"path/filepath"
)

func init() {
	global.LRAG_CONFIG.AutoCode.Root, _ = filepath.Abs("../../../")
	global.LRAG_CONFIG.AutoCode.Server = "server"
}
