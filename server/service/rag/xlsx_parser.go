package rag

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ParseXLSXContent 从 Excel 文件提取文本
// 遍历所有工作表和行，将数据拼接为文本
func ParseXLSXContent(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 Excel 数据失败: %w", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("解析 Excel 失败: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	sheets := f.GetSheetList()

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

		if len(sheets) > 1 {
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString("## ")
			sb.WriteString(sheetName)
			sb.WriteString("\n")
		}

		for _, row := range rows {
			nonEmpty := false
			for _, cell := range row {
				if strings.TrimSpace(cell) != "" {
					nonEmpty = true
					break
				}
			}
			if !nonEmpty {
				continue
			}
			line := strings.Join(row, "\t")
			sb.WriteString(strings.TrimRight(line, "\t "))
			sb.WriteString("\n")
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("Excel 未提取到文本")
	}
	return result, nil
}
