package rag

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ParseDOCXContent 从 DOCX 文件提取纯文本
// DOCX 是 ZIP 格式，包含 word/document.xml
func ParseDOCXContent(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 DOCX 数据失败: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 DOCX 失败（非有效 ZIP）: %w", err)
	}

	var sb strings.Builder

	byName := make(map[string]*zip.File, len(zipReader.File))
	for _, f := range zipReader.File {
		byName[f.Name] = f
	}

	var docPart *zip.File
	var headers, footers []string
	for _, f := range zipReader.File {
		n := f.Name
		switch {
		case n == "word/document.xml":
			docPart = f
		case strings.HasPrefix(n, "word/header") && strings.HasSuffix(n, ".xml"):
			headers = append(headers, n)
		case strings.HasPrefix(n, "word/footer") && strings.HasSuffix(n, ".xml"):
			footers = append(footers, n)
		}
	}
	sort.Strings(headers)
	sort.Strings(footers)

	var extra []string
	extra = append(extra, headers...)
	extra = append(extra, footers...)
	for _, fixed := range []string{"word/footnotes.xml", "word/endnotes.xml", "word/comments.xml"} {
		if _, ok := byName[fixed]; ok {
			extra = append(extra, fixed)
		}
	}

	parts := make([]*zip.File, 0, 1+len(extra))
	if docPart != nil {
		parts = append(parts, docPart)
	}
	for _, name := range extra {
		if f := byName[name]; f != nil {
			parts = append(parts, f)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("DOCX 中未找到可解析的 word 文档部件")
	}
	for i, f := range parts {
		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("打开 %s 失败: %w", f.Name, err)
		}
		text, err := extractTextFromDocXML(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if sb.Len() > 0 {
			if i > 0 && strings.Contains(f.Name, "header") {
				sb.WriteString("\n\n## Header\n")
			} else if i > 0 && strings.Contains(f.Name, "footer") {
				sb.WriteString("\n\n## Footer\n")
			} else if i > 0 && strings.Contains(f.Name, "footnote") {
				sb.WriteString("\n\n## Footnotes\n")
			} else if i > 0 && strings.Contains(f.Name, "endnote") {
				sb.WriteString("\n\n## Endnotes\n")
			} else if i > 0 && strings.Contains(f.Name, "comment") {
				sb.WriteString("\n\n## Comments\n")
			} else if i > 0 {
				sb.WriteString("\n\n")
			}
		}
		sb.WriteString(text)
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("DOCX 未提取到文本")
	}
	return result, nil
}

// extractTextFromDocXML 从 word/document.xml 的 XML 流中提取文本
func extractTextFromDocXML(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var sb strings.Builder
	var inParagraph bool
	var paragraphText strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return sb.String(), nil
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				inParagraph = true
				paragraphText.Reset()
			case "tab":
				if inParagraph {
					paragraphText.WriteString("\t")
				}
			case "br":
				if inParagraph {
					paragraphText.WriteString("\n")
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				if inParagraph {
					text := strings.TrimSpace(paragraphText.String())
					if text != "" {
						if sb.Len() > 0 {
							sb.WriteString("\n")
						}
						sb.WriteString(text)
					}
					inParagraph = false
				}
			}
		case xml.CharData:
			if inParagraph {
				paragraphText.Write(t)
			}
		}
	}

	return sb.String(), nil
}
