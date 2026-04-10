package rag

import (
	"io"

	"github.com/LightningRAG/LightningRAG/server/service/rag/docparse"
)

// PDF 文本抽取由 docparse 统一实现（对齐 Ragflow）：
//   - 默认：优先 pypdf（docparse/pypdfplain + references/pypdf），失败或正文为空再 ledongthuc/pdf。
//   - LIGHTNINGRAG_PDF_ENGINE：pypdf | ledongthuc | thuc | auto（先 ledongthuc 再 pypdf）。
// 环境变量：LIGHTNINGRAG_PYPDF_SRC、LIGHTNINGRAG_REPO_ROOT、LIGHTNINGRAG_PYTHON、
// LIGHTNINGRAG_PDF_PASSWORD、LIGHTNINGRAG_PYPDF_EXTRACTION_MODE、页范围与 full 附加块、
// LIGHTNINGRAG_PYPDF_STRICT、LIGHTNINGRAG_PYPDF_ROOT_RECOVERY_LIMIT 等见仓库文档与 pypdfplain 包注释。

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

// ExtractPDFPypdfDocInfo 使用 references/pypdf 读取文档信息字典（非 XMP）。
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

// ExtractPDFPypdfXMPMetadata 返回 pypdf xmp_metadata 摘要字段。
func ExtractPDFPypdfXMPMetadata(data []byte) (map[string]any, error) {
	return docparse.ExtractPDFPypdfXMPMetadata(data)
}
