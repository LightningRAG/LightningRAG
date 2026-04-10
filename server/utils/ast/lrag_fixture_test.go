package ast

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
)

func lragPluginRoot() string {
	return filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag")
}

func skipIfNoLRagPlugin(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(lragPluginRoot()); err != nil {
		t.Skipf("skip: lrag plugin tree missing (%s)", lragPluginRoot())
	}
}

func skipUnlessFile(t *testing.T, p string) {
	t.Helper()
	if p == "" {
		return
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("skip: fixture file missing: %s", p)
	}
}
