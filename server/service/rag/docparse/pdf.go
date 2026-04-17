package docparse

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PDFPypdfDocInfo 为 /Info 字典与页数等字段（历史导出名；现由 pdf-go 填充）。
type PDFPypdfDocInfo struct {
	PageCount           int    `json:"page_count"`
	PDFHeader           string `json:"pdf_header"`
	Title               string `json:"title,omitempty"`
	Author              string `json:"author,omitempty"`
	Subject             string `json:"subject,omitempty"`
	Creator             string `json:"creator,omitempty"`
	Producer            string `json:"producer,omitempty"`
	Keywords            string `json:"keywords,omitempty"`
	CreationDate        string `json:"creation_date,omitempty"`
	ModificationDate    string `json:"modification_date,omitempty"`
	CreationDateRaw     string `json:"creation_date_raw,omitempty"`
	ModificationDateRaw string `json:"modification_date_raw,omitempty"`
}

// ParsePDFFromReader 从 PDF 流提取纯文本（引擎见 pdf_engine.go）。
func ParsePDFFromReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 PDF 数据失败: %w", err)
	}
	return ParsePDFFromBytes(data)
}

// ParsePDFFromBytes 从内存解析 PDF 文本；默认优先 pdf-go，失败或空结果回退 ledongthuc。
func ParsePDFFromBytes(data []byte) (string, error) {
	switch pdfEngineMode() {
	case "pdfgo":
		return extractPlainFullPdfGo(data)
	case "ledongthuc":
		return parsePDFFromMemoryLedongthuc(data)
	case "auto":
		s, err := parsePDFFromMemoryLedongthuc(data)
		if err == nil && strings.TrimSpace(s) != "" {
			return s, nil
		}
		ps, perr := extractPlainFullPdfGo(data)
		if perr == nil {
			return ps, nil
		}
		if err != nil {
			return "", fmt.Errorf("ledongthuc: %v; pdf-go: %w", err, perr)
		}
		return "", perr
	default:
		ps, perr := extractPlainFullPdfGo(data)
		if perr == nil && strings.TrimSpace(ps) != "" {
			return ps, nil
		}
		s, err := parsePDFFromMemoryLedongthuc(data)
		if err != nil {
			if perr != nil {
				return "", fmt.Errorf("pdf-go: %w; ledongthuc: %v", perr, err)
			}
			return "", err
		}
		return s, nil
	}
}

// ParsePDFFromFile 从文件路径解析 PDF。
func ParsePDFFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ParsePDFFromReader(f)
}

// GetPDFPageCount 获取页数；default 模式下优先 pdf-go，失败或 0 页再 ledongthuc。
func GetPDFPageCount(data []byte) (int, error) {
	switch pdfEngineMode() {
	case "pdfgo":
		return extractPageCountPdfGo(data)
	case "ledongthuc":
		return getPDFPageCountLedongthuc(data)
	case "auto":
		n, err := getPDFPageCountLedongthuc(data)
		if err == nil && n > 0 {
			return n, nil
		}
		n2, err2 := extractPageCountPdfGo(data)
		if err2 == nil {
			return n2, nil
		}
		if err != nil {
			return 0, fmt.Errorf("ledongthuc: %v; pdf-go: %w", err, err2)
		}
		return 0, err2
	default:
		n, err := extractPageCountPdfGo(data)
		if err == nil && n > 0 {
			return n, nil
		}
		n2, err2 := getPDFPageCountLedongthuc(data)
		if err2 != nil {
			if err != nil {
				return 0, fmt.Errorf("pdf-go: %w; ledongthuc: %v", err, err2)
			}
			return 0, err2
		}
		return n2, nil
	}
}

// ParsePDFByPage 按页提取；default 优先 pdf-go，失败或各页皆空再 ledongthuc。
func ParsePDFByPage(data []byte) ([]string, error) {
	switch pdfEngineMode() {
	case "pdfgo":
		return extractPagesPdfGo(data)
	case "ledongthuc":
		return parsePDFByPageLedongthuc(data)
	case "auto":
		pages, err := parsePDFByPageLedongthuc(data)
		if err == nil && pdfPagesAnyNonEmpty(pages) {
			return pages, nil
		}
		ps, perr := extractPagesPdfGo(data)
		if perr == nil {
			return ps, nil
		}
		if err != nil {
			return nil, fmt.Errorf("ledongthuc: %v; pdf-go: %w", err, perr)
		}
		return nil, perr
	default:
		ps, perr := extractPagesPdfGo(data)
		if perr == nil && pdfPagesAnyNonEmpty(ps) {
			return ps, nil
		}
		pages, err := parsePDFByPageLedongthuc(data)
		if err != nil {
			if perr != nil {
				return nil, fmt.Errorf("pdf-go: %w; ledongthuc: %v", perr, err)
			}
			return nil, err
		}
		return pages, nil
	}
}

// ExtractPDFPypdfDocInfo 读取文档信息字典（非 XMP）；由 pdf-go 实现。
func ExtractPDFPypdfDocInfo(data []byte) (*PDFPypdfDocInfo, error) {
	return extractDocInfoPdfGo(data)
}

// ExtractPDFPypdfURILinks 提取注释中的 /URI。
func ExtractPDFPypdfURILinks(data []byte) ([]string, error) {
	return extractURILinksPdfGo(data)
}

// ExtractPDFPypdfPageLabels 返回每页逻辑页码标签。
func ExtractPDFPypdfPageLabels(data []byte) ([]string, error) {
	return extractPageLabelsPdfGo(data)
}

// ExtractPDFPypdfAttachmentNames 返回嵌入附件文件名。
func ExtractPDFPypdfAttachmentNames(data []byte) ([]string, error) {
	return extractAttachmentNamesPdfGo(data)
}

// ExtractPDFPypdfXMPMetadata 返回 XMP 摘要字段。
func ExtractPDFPypdfXMPMetadata(data []byte) (map[string]any, error) {
	return extractXMPMetadataPdfGo(data)
}
