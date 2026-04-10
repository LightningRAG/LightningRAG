package docparse

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/service/rag/docparse/pypdfplain"
)

// PDFPypdfDocInfo 为 pypdf 文档信息字典类型别名（供 rag 等包稳定引用）。
type PDFPypdfDocInfo = pypdfplain.DocInfo

// ParsePDFFromReader 从 PDF 流提取纯文本（引擎见 pdf_engine.go）。
func ParsePDFFromReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 PDF 数据失败: %w", err)
	}
	return ParsePDFFromBytes(data)
}

// ParsePDFFromBytes 从内存解析 PDF 文本；默认优先 pypdf，失败或空结果回退 ledongthuc。
func ParsePDFFromBytes(data []byte) (string, error) {
	switch pdfEngineMode() {
	case "pypdf":
		return pypdfplain.ExtractPlainFull(data)
	case "ledongthuc":
		return parsePDFFromMemoryLedongthuc(data)
	case "auto":
		s, err := parsePDFFromMemoryLedongthuc(data)
		if err == nil && strings.TrimSpace(s) != "" {
			return s, nil
		}
		ps, perr := pypdfplain.ExtractPlainFull(data)
		if perr == nil {
			return ps, nil
		}
		if err != nil {
			return "", fmt.Errorf("ledongthuc: %v; pypdf: %w", err, perr)
		}
		return "", perr
	default:
		ps, perr := pypdfplain.ExtractPlainFull(data)
		if perr == nil && strings.TrimSpace(ps) != "" {
			return ps, nil
		}
		s, err := parsePDFFromMemoryLedongthuc(data)
		if err != nil {
			if perr != nil {
				return "", fmt.Errorf("pypdf: %w; ledongthuc: %v", perr, err)
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

// GetPDFPageCount 获取页数；default 模式下优先 pypdf，失败或 0 页再 ledongthuc。
func GetPDFPageCount(data []byte) (int, error) {
	switch pdfEngineMode() {
	case "pypdf":
		return pypdfplain.ExtractPageCount(data)
	case "ledongthuc":
		return getPDFPageCountLedongthuc(data)
	case "auto":
		n, err := getPDFPageCountLedongthuc(data)
		if err == nil && n > 0 {
			return n, nil
		}
		n2, err2 := pypdfplain.ExtractPageCount(data)
		if err2 == nil {
			return n2, nil
		}
		if err != nil {
			return 0, fmt.Errorf("ledongthuc: %v; pypdf: %w", err, err2)
		}
		return 0, err2
	default:
		n, err := pypdfplain.ExtractPageCount(data)
		if err == nil && n > 0 {
			return n, nil
		}
		n2, err2 := getPDFPageCountLedongthuc(data)
		if err2 != nil {
			if err != nil {
				return 0, fmt.Errorf("pypdf: %w; ledongthuc: %v", err, err2)
			}
			return 0, err2
		}
		return n2, nil
	}
}

// ParsePDFByPage 按页提取；default 优先 pypdf，失败或各页皆空再 ledongthuc。
func ParsePDFByPage(data []byte) ([]string, error) {
	switch pdfEngineMode() {
	case "pypdf":
		return pypdfplain.ExtractPages(data)
	case "ledongthuc":
		return parsePDFByPageLedongthuc(data)
	case "auto":
		pages, err := parsePDFByPageLedongthuc(data)
		if err == nil && pdfPagesAnyNonEmpty(pages) {
			return pages, nil
		}
		ps, perr := pypdfplain.ExtractPages(data)
		if perr == nil {
			return ps, nil
		}
		if err != nil {
			return nil, fmt.Errorf("ledongthuc: %v; pypdf: %w", err, perr)
		}
		return nil, perr
	default:
		ps, perr := pypdfplain.ExtractPages(data)
		if perr == nil && pdfPagesAnyNonEmpty(ps) {
			return ps, nil
		}
		pages, err := parsePDFByPageLedongthuc(data)
		if err != nil {
			if perr != nil {
				return nil, fmt.Errorf("pypdf: %w; ledongthuc: %v", perr, err)
			}
			return nil, err
		}
		return pages, nil
	}
}

// ExtractPDFPypdfDocInfo 使用 references/pypdf 读取文档信息字典。
func ExtractPDFPypdfDocInfo(data []byte) (*PDFPypdfDocInfo, error) {
	return pypdfplain.ExtractDocInfo(data)
}

// ExtractPDFPypdfURILinks 提取注释中的 /URI。
func ExtractPDFPypdfURILinks(data []byte) ([]string, error) {
	return pypdfplain.ExtractURILinks(data)
}

// ExtractPDFPypdfPageLabels 返回每页逻辑页码标签。
func ExtractPDFPypdfPageLabels(data []byte) ([]string, error) {
	return pypdfplain.ExtractPageLabels(data)
}

// ExtractPDFPypdfAttachmentNames 返回嵌入附件文件名。
func ExtractPDFPypdfAttachmentNames(data []byte) ([]string, error) {
	return pypdfplain.ExtractAttachmentNames(data)
}

// ExtractPDFPypdfXMPMetadata 返回 XMP 摘要字段。
func ExtractPDFPypdfXMPMetadata(data []byte) (map[string]any, error) {
	return pypdfplain.ExtractXMPMetadata(data)
}
