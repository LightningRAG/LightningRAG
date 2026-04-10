package system

import (
	"context"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/LightningRAG/LightningRAG/server/utils/autocode"
	"os"
	"path/filepath"
	"text/template"
)

func (s *autoCodeTemplate) CreateMcp(ctx context.Context, info request.AutoMcpTool) (toolFilePath string, err error) {
	mcpTemplatePath := filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "resource", "mcp", "tools.tpl")
	mcpToolPath := filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "mcp")

	var files *template.Template

	templateName := filepath.Base(mcpTemplatePath)

	files, err = template.New(templateName).Funcs(autocode.GetTemplateFuncMap()).ParseFiles(mcpTemplatePath)
	if err != nil {
		return
	}

	fileName := utils.HumpToUnderscore(info.Name)

	toolFilePath = filepath.Join(mcpToolPath, fileName+".go")

	f, err := os.Create(toolFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	// 执行模板，将内容写入文件
	err = files.Execute(f, info)
	if err != nil {
		return
	}

	return

}
