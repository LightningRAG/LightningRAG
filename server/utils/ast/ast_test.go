package ast

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestAst(t *testing.T) {
	filename := filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin", "lrag", "plugin.go")
	if _, err := os.Stat(filename); err != nil {
		t.Skip("skip: plugin fixture not present:", filename)
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		t.Error(err)
		return
	}
	err = ast.Print(fileSet, file)
	if err != nil {
		t.Error(err)
		return
	}
	err = printer.Fprint(os.Stdout, token.NewFileSet(), file)
	if err != nil {
		panic(err)
	}

}
