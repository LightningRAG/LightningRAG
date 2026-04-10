package docparse

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/extrame/xls"
)

// ParseXLSText 从二进制 Excel 97–2003（.xls / OLE）抽取文本，对齐 Ragflow spreadsheet 对 xls 的支持。
func ParseXLSText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("xls 数据为空")
	}
	wb, err := xls.OpenReader(bytes.NewReader(data), "utf-8")
	if err != nil || wb == nil {
		wb, err = xls.OpenReader(bytes.NewReader(data), "gbk")
	}
	if err != nil {
		return "", fmt.Errorf("解析 xls 失败: %w", err)
	}
	if wb == nil {
		return "", fmt.Errorf("解析 xls 失败")
	}
	if wb.NumSheets() == 0 {
		return "", fmt.Errorf("xls 中无工作表")
	}

	var sb strings.Builder
	for si := 0; si < wb.NumSheets(); si++ {
		sh := wb.GetSheet(si)
		if sh == nil {
			continue
		}
		name := strings.TrimSpace(sh.Name)
		if wb.NumSheets() > 1 {
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString("## ")
			sb.WriteString(name)
			sb.WriteString("\n")
		}
		maxR := int(sh.MaxRow)
		for ri := 0; ri <= maxR; ri++ {
			row := sh.Row(ri)
			if row == nil {
				continue
			}
			fc, lc := row.FirstCol(), row.LastCol()
			if lc < fc {
				continue
			}
			var cells []string
			nonEmpty := false
			for ci := fc; ci <= lc; ci++ {
				c := strings.TrimSpace(row.Col(ci))
				if c != "" {
					nonEmpty = true
				}
				cells = append(cells, c)
			}
			if !nonEmpty {
				continue
			}
			sb.WriteString(strings.Join(cells, "\t"))
			sb.WriteString("\n")
		}
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("xls 未提取到文本")
	}
	return out, nil
}
