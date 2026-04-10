package docparse

import (
	"archive/zip"
	"bytes"
	"io"
	"path"
	"strings"
)

// RefineFileTypeByContent 在扩展名不可靠（如 .txt / 无后缀）或内容与扩展名矛盾时，用魔数 + ZIP 内布局纠偏 fileType。
func RefineFileTypeByContent(filename string, hinted string, data []byte) string {
	h := strings.ToLower(strings.TrimSpace(hinted))
	if len(data) == 0 {
		return h
	}
	lowName := strings.ToLower(filename)
	if IsBzip2Magic(data) && (strings.HasSuffix(lowName, ".bz2") || strings.HasSuffix(lowName, ".tbz2") || h == "txt" || h == "" || h == "bz2") {
		return "bz2"
	}
	if isGzipMagic(data) && (strings.HasSuffix(lowName, ".gz") || strings.HasSuffix(lowName, ".svgz") || h == "txt" || h == "" || h == "gz") {
		return "gz"
	}
	if isPDFMagic(data) {
		if h == "txt" || h == "" || h == "pdf" {
			return "pdf"
		}
	}
	if !isZIPMagic(data) {
		return h
	}
	k := zipDocumentKind(data)
	if k == "" {
		return h
	}
	// 金山 WPS 等使用专用扩展名但内为 OOXML ZIP 时，与 Ragflow 将此类归入「文档」一致
	if h == "wps" && k == "docx" {
		return "docx"
	}
	if h == "et" && k == "xlsx" {
		return "xlsx"
	}
	if h == "dps" && k == "pptx" {
		return "pptx"
	}
	if h == "txt" || h == "" {
		return k
	}
	if contradictsExtension(h, k) {
		return k
	}
	return h
}

func contradictsExtension(have, fromZip string) bool {
	// 明显错扩展：标成表格实为 Word 等
	pairs := []struct{ a, b string }{
		{"xlsx", "docx"}, {"xlsx", "pptx"}, {"xlsx", "epub"}, {"xlsx", "odt"}, {"xlsx", "ods"}, {"xlsx", "odp"}, {"xlsx", "odg"},
		{"docx", "xlsx"}, {"docx", "pptx"}, {"docx", "ods"}, {"docx", "odt"}, {"docx", "odp"}, {"docx", "odg"}, {"docx", "epub"},
		{"pptx", "docx"}, {"pptx", "xlsx"}, {"pptx", "odp"}, {"pptx", "odg"},
		{"txt", "docx"}, {"txt", "xlsx"}, {"txt", "pptx"}, {"txt", "pdf"},
		{"txt", "epub"}, {"txt", "odt"}, {"txt", "ods"}, {"txt", "odp"}, {"txt", "odg"},
		{"pdf", "docx"}, {"pdf", "xlsx"}, {"pdf", "pptx"}, {"pdf", "epub"}, {"pdf", "odt"}, {"pdf", "ods"}, {"pdf", "odp"}, {"pdf", "odg"},
		{"odg", "docx"}, {"odg", "xlsx"}, {"odg", "pptx"}, {"odg", "epub"},
		{"odt", "odg"}, {"odp", "odg"}, {"ods", "odg"},
		{"odg", "odt"}, {"odg", "odp"}, {"odg", "ods"},
	}
	for _, p := range pairs {
		if have == p.a && fromZip == p.b {
			return true
		}
	}
	return false
}

// IsBzip2Magic 检测 bzip2 流魔数（BZh + 块大小数字 1–9）。
func IsBzip2Magic(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	if b[0] != 'B' || b[1] != 'Z' || b[2] != 'h' {
		return false
	}
	return b[3] >= '1' && b[3] <= '9'
}

func isGzipMagic(b []byte) bool {
	return len(b) >= 2 && b[0] == 0x1f && b[1] == 0x8b
}

func isPDFMagic(b []byte) bool {
	i := 0
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		i = 3
	}
	for i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
		i++
	}
	return len(b) >= i+4 && string(b[i:i+4]) == "%PDF"
}

func isZIPMagic(b []byte) bool {
	return len(b) >= 4 && b[0] == 'P' && b[1] == 'K' && (b[2] == 3 || b[2] == 5 || b[2] == 7) && (b[3] == 4 || b[3] == 6 || b[3] == 8)
}

func zipDocumentKind(data []byte) string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(zr.File) == 0 {
		return ""
	}
	lowerNames := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		lowerNames = append(lowerNames, strings.ToLower(f.Name))
	}

	hasPath := func(suffix string) bool {
		suffix = strings.ToLower(suffix)
		for _, n := range lowerNames {
			if n == suffix || strings.HasSuffix(n, "/"+suffix) {
				return true
			}
		}
		return false
	}

	if hasPath("word/document.xml") {
		return "docx"
	}
	if hasPath("xl/workbook.xml") {
		return "xlsx"
	}
	if hasPath("ppt/presentation.xml") {
		return "pptx"
	}
	if hasPath("meta-inf/container.xml") {
		for _, n := range lowerNames {
			ext := strings.ToLower(path.Ext(n))
			if ext == ".xhtml" || ext == ".html" || ext == ".htm" {
				return "epub"
			}
		}
	}

	// ODF / EPUB mimetype 文件（ODF 规范为未压缩首项）
	for _, f := range zr.File {
		if strings.ToLower(path.Base(f.Name)) != "mimetype" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		b, _ := io.ReadAll(rc)
		_ = rc.Close()
		mt := strings.TrimSpace(string(b))
		switch {
		case strings.HasPrefix(mt, "application/epub+zip"):
			return "epub"
		case strings.Contains(mt, "opendocument.text"):
			return "odt"
		case strings.Contains(mt, "opendocument.spreadsheet"):
			return "ods"
		case strings.Contains(mt, "opendocument.presentation"):
			return "odp"
		case strings.Contains(mt, "opendocument.graphics"):
			return "odg"
		}
		break
	}

	if hasPath("content.xml") && hasPath("meta-inf/manifest.xml") {
		// 无 mimetype 的残缺 ODF：按 manifest 路径粗判
		for _, f := range zr.File {
			if !strings.EqualFold(f.Name, "meta-inf/manifest.xml") {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				break
			}
			raw, _ := io.ReadAll(rc)
			_ = rc.Close()
			s := strings.ToLower(string(raw))
			if strings.Contains(s, "application/vnd.oasis.opendocument.presentation") {
				return "odp"
			}
			if strings.Contains(s, "application/vnd.oasis.opendocument.spreadsheet") {
				return "ods"
			}
			if strings.Contains(s, "application/vnd.oasis.opendocument.text") {
				return "odt"
			}
			if strings.Contains(s, "application/vnd.oasis.opendocument.graphics") {
				return "odg"
			}
			break
		}
	}

	return ""
}
