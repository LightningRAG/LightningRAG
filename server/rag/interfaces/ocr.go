// Package interfaces 定义 OCR 光学字符识别接口，参考 references 目录内 ocr_model
package interfaces

import "context"

// OCRResult  OCR 解析结果
type OCRResult struct {
	Text     string   // 提取的文本
	Sections []string // 可选的段落/区块（如 PDF 解析）
	Tables   []string // 可选的表格内容
}

// OCR 光学字符识别接口，参考 references 目录内 ocr_model.Base
// 支持从图片或 PDF 中提取文字
type OCR interface {
	// ExtractText 从图片/PDF 字节中提取文字
	ExtractText(ctx context.Context, data []byte, filename string) (*OCRResult, error)

	// ProviderName 返回提供商名称
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string
}
