package rag

import (
	"io"

	"github.com/LightningRAG/LightningRAG/server/service/rag/docparse"
)

// PDF 文本抽取由 docparse 统一实现：
//   - 默认：优先 github.com/lightningrag/pdf-go（每页 ExtractTextAdvanced），失败或正文为空再 ledongthuc/pdf。
//   - LIGHTNINGRAG_PDF_ENGINE：pdfgo | pypdf（同 pdf-go）| ledongthuc | thuc | auto（先 ledongthuc 再 pdf-go）。

// ParsePDFContent 从 PDF 文件中提取纯文本。
func ParsePDFContent(r io.Reader) (string, error) {
	return docparse.ParsePDFFromReader(r)
}

// ParsePDFFromMemory 从内存字节解析 PDF 文本。
func ParsePDFFromMemory(data []byte) (string, error) {
	return docparse.ParsePDFFromBytes(data)
}

// ParsePDFFromFile 从文件路径解析 PDF。
func ParsePDFFromFile(path string) (string, error) {
	return docparse.ParsePDFFromFile(path)
}

// GetPDFPageCount 获取 PDF 页数。
func GetPDFPageCount(data []byte) (int, error) {
	return docparse.GetPDFPageCount(data)
}

// ParsePDFByPage 按页提取 PDF 文本。
func ParsePDFByPage(data []byte) ([]string, error) {
	return docparse.ParsePDFByPage(data)
}

// PDFPypdfDocInfo 为 pypdf 文档信息字段（与 docparse 一致）。
type PDFPypdfDocInfo = docparse.PDFPypdfDocInfo

// ExtractPDFPypdfDocInfo 读取文档信息字典（非 XMP）；由 pdf-go 实现。
func ExtractPDFPypdfDocInfo(data []byte) (*PDFPypdfDocInfo, error) {
	return docparse.ExtractPDFPypdfDocInfo(data)
}

// ExtractPDFPypdfURILinks 提取页面注释中的 /URI。
func ExtractPDFPypdfURILinks(data []byte) ([]string, error) {
	return docparse.ExtractPDFPypdfURILinks(data)
}

// ExtractPDFPypdfPageLabels 返回每页逻辑页码标签。
func ExtractPDFPypdfPageLabels(data []byte) ([]string, error) {
	return docparse.ExtractPDFPypdfPageLabels(data)
}

// ExtractPDFPypdfAttachmentNames 返回嵌入附件文件名。
func ExtractPDFPypdfAttachmentNames(data []byte) ([]string, error) {
	return docparse.ExtractPDFPypdfAttachmentNames(data)
}

// ExtractPDFPypdfXMPMetadata 返回 XMP 摘要字段（pdf-go 解析）。
func ExtractPDFPypdfXMPMetadata(data []byte) (map[string]any, error) {
	return docparse.ExtractPDFPypdfXMPMetadata(data)
}
