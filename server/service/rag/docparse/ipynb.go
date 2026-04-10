package docparse

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseIPynbText 从 Jupyter Notebook（.ipynb）抽取代码与 markdown 单元文本。
func ParseIPynbText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("ipynb 数据为空")
	}
	var nb struct {
		Cells []struct {
			CellType string          `json:"cell_type"`
			Source   json.RawMessage `json:"source"`
		} `json:"cells"`
	}
	if err := json.Unmarshal(data, &nb); err != nil {
		return "", fmt.Errorf("解析 ipynb 失败: %w", err)
	}
	if len(nb.Cells) == 0 {
		return "", fmt.Errorf("ipynb 无单元格")
	}
	var parts []string
	for _, c := range nb.Cells {
		if c.CellType != "markdown" && c.CellType != "code" && c.CellType != "raw" {
			continue
		}
		text := normalizeIPynbSource(c.Source)
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		parts = append(parts, "["+c.CellType+"]\n"+text)
	}
	out := strings.TrimSpace(strings.Join(parts, "\n\n"))
	if out == "" {
		return "", fmt.Errorf("ipynb 未提取到文本")
	}
	return out, nil
}

func normalizeIPynbSource(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var lines []string
	if json.Unmarshal(raw, &lines) == nil {
		return strings.Join(lines, "")
	}
	return string(raw)
}
