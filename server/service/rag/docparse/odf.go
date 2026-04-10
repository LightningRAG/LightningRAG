package docparse

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ParseODTText 从 OpenDocument Text（.odt，ZIP + content.xml）抽取段落/标题文本。
// 对齐 Ragflow/naive 对开放文档生态的常见支持范围（LibreOffice / ODF）。
func ParseODTText(data []byte) (string, error) {
	return parseODFContentXML(data, "odt")
}

// ParseODSText 从 OpenDocument Spreadsheet（.ods）抽取表格单元格文本（按行拼接）。
func ParseODSText(data []byte) (string, error) {
	return parseODFContentXML(data, "ods")
}

// ParseODPText 从 OpenDocument Presentation（.odp）抽取文本（与 ODT 相同遍历 text:p / text:h）。
func ParseODPText(data []byte) (string, error) {
	return parseODFContentXML(data, "odp")
}

// ParseODGText 从 OpenDocument Drawing（.odg，绘图/图表包）抽取 content.xml 中的文本节点（与 ODP 同类遍历）。
func ParseODGText(data []byte) (string, error) {
	return parseODFContentXML(data, "odg")
}

func parseODFContentXML(data []byte, kind string) (string, error) {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return "", fmt.Errorf("不是有效的 ODF zip 包")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 ODF zip 失败: %w", err)
	}
	var rc io.ReadCloser
	for _, f := range zr.File {
		if f.Name == "content.xml" {
			rc, err = f.Open()
			break
		}
	}
	if rc == nil {
		return "", fmt.Errorf("ODF 中未找到 content.xml")
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	if kind == "ods" {
		return extractODSTableText(bytes.NewReader(raw))
	}
	// odt / odp / odg：段落与标题文本
	return extractODTText(bytes.NewReader(raw))
}

func extractODTText(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	var sb strings.Builder
	var stack []string
	flushPara := func() {
		if len(stack) == 0 {
			return
		}
		line := strings.TrimSpace(strings.Join(stack, ""))
		stack = stack[:0]
		if line == "" {
			return
		}
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(line)
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return sb.String(), nil
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p", "h":
				stack = stack[:0]
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p", "h":
				flushPara()
			}
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" {
				stack = append(stack, s)
			}
		}
	}
	flushPara()
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("ODF 文本文档未提取到内容")
	}
	return out, nil
}

func extractODSTableText(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	var sb strings.Builder
	var row []string
	var inCell bool
	var cellBuf strings.Builder

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return sb.String(), nil
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "table-row":
				row = row[:0]
			case "table-cell":
				inCell = true
				cellBuf.Reset()
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "table-cell":
				inCell = false
				cell := strings.TrimSpace(cellBuf.String())
				row = append(row, cell)
			case "table-row":
				nonEmpty := false
				for _, c := range row {
					if strings.TrimSpace(c) != "" {
						nonEmpty = true
						break
					}
				}
				if nonEmpty {
					if sb.Len() > 0 {
						sb.WriteString("\n")
					}
					sb.WriteString(strings.Join(row, "\t"))
				}
				row = row[:0]
			}
		case xml.CharData:
			if inCell {
				cellBuf.WriteString(string(t))
			}
		}
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("ODS 未提取到文本")
	}
	return out, nil
}
