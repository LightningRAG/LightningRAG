package rag

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// extractIWorkPreviewPDF 从 Apple iWork 单文件包（.pages / .numbers / .key，实为 ZIP）中读取 QuickLook/Preview.pdf。
// 对齐 Ragflow file_service 将 .pages 与演示类一并考虑；无法解析 iwa 正文时至少抽取系统生成的预览 PDF 供检索。
func extractIWorkPreviewPDF(data []byte) ([]byte, error) {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return nil, fmt.Errorf("Apple iWork 文稿应为 ZIP 单文件包")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("打开 iWork 包失败: %w", err)
	}
	for _, f := range zr.File {
		ln := strings.ToLower(filepath.ToSlash(f.Name))
		if ln == "quicklook/preview.pdf" {
			return readZipFileEntry(f)
		}
	}
	for _, f := range zr.File {
		ln := strings.ToLower(filepath.ToSlash(f.Name))
		if strings.HasSuffix(ln, "/preview.pdf") {
			return readZipFileEntry(f)
		}
	}
	return nil, fmt.Errorf("包内未找到 QuickLook/Preview.pdf，请在 Pages/Numbers/Keynote 中导出为 PDF 或 Office 格式后上传")
}

func readZipFileEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
