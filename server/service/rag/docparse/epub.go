package docparse

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

// ParseEPUBText 从 EPUB（ZIP + HTML/XHTML）抽取正文，对齐常见电子书入库场景。
func ParseEPUBText(data []byte) (string, error) {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return "", fmt.Errorf("不是有效的 EPUB（需为 zip）")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 EPUB zip 失败: %w", err)
	}

	type item struct {
		name string
	}
	var items []item
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		n := strings.ToLower(f.Name)
		if strings.HasPrefix(n, "meta-inf/") {
			continue
		}
		ext := path.Ext(n)
		switch ext {
		case ".xhtml", ".html", ".htm":
			items = append(items, item{name: f.Name})
		}
	}
	if len(items) == 0 {
		return "", fmt.Errorf("EPUB 中未找到 html/xhtml 正文")
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })

	var sb strings.Builder
	for _, it := range items {
		f, err := zr.Open(it.name)
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			continue
		}
		t := HTMLToPlainText(string(raw))
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(t)
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("EPUB 未提取到文本")
	}
	return out, nil
}
