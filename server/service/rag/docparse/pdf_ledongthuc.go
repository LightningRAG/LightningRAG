package docparse

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// formatPDFOutlineLines 将 PDF 书签树展平为带缩进的标题行（深度优先）；无书签时返回 nil。
func formatPDFOutlineLines(root pdf.Outline) []string {
	var lines []string
	var walk func(o pdf.Outline, depth int)
	walk = func(o pdf.Outline, depth int) {
		t := strings.TrimSpace(o.Title)
		if t != "" {
			lines = append(lines, strings.Repeat("  ", depth)+t)
		}
		for i := range o.Child {
			walk(o.Child[i], depth+1)
		}
	}
	for i := range root.Child {
		walk(root.Child[i], 0)
	}
	if len(lines) == 0 {
		return nil
	}
	return lines
}

func parsePDFFromMemoryLedongthuc(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("PDF 数据为空")
	}

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 PDF 失败: %w", err)
	}

	n := reader.NumPage()
	fonts := make(map[string]*pdf.Font)
	var sb strings.Builder

	for i := 1; i <= n; i++ {
		p := reader.Page(i)
		for _, name := range p.Fonts() {
			if _, ok := fonts[name]; !ok {
				f := p.Font(name)
				fonts[name] = &f
			}
		}
		text, err := p.GetPlainText(fonts)
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(text)
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("PDF 未提取到文本（可能是扫描件或图片型 PDF，需 OCR 处理）")
	}
	if ol := formatPDFOutlineLines(reader.Outline()); len(ol) > 0 {
		var ob strings.Builder
		ob.WriteString(result)
		ob.WriteString("\n\n--- PDF 大纲 / 书签 ---\n")
		ob.WriteString(strings.Join(ol, "\n"))
		result = ob.String()
	}
	return result, nil
}

func getPDFPageCountLedongthuc(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("PDF 数据为空")
	}
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return 0, err
	}
	return reader.NumPage(), nil
}

func pdfPagesAnyNonEmpty(pages []string) bool {
	for _, s := range pages {
		if strings.TrimSpace(s) != "" {
			return true
		}
	}
	return false
}

func parsePDFByPageLedongthuc(data []byte) ([]string, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("PDF 数据为空")
	}
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("解析 PDF 失败: %w", err)
	}

	n := reader.NumPage()
	fonts := make(map[string]*pdf.Font)
	pages := make([]string, 0, n)

	for i := 1; i <= n; i++ {
		p := reader.Page(i)
		for _, name := range p.Fonts() {
			if _, ok := fonts[name]; !ok {
				f := p.Font(name)
				fonts[name] = &f
			}
		}
		text, err := p.GetPlainText(fonts)
		if err != nil {
			pages = append(pages, "")
			continue
		}
		pages = append(pages, strings.TrimSpace(text))
	}
	return pages, nil
}
