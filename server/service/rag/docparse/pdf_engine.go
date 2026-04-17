package docparse

import (
	"os"
	"strings"
)

// pdfEngineMode 解析 LIGHTNINGRAG_PDF_ENGINE：
//   - 未设置或其它值 — default：优先 pdf-go（ExtractTextAdvanced），失败或正文为空再 ledongthuc
//   - pdfgo / pypdf — 仅 pdf-go（pypdf 为兼容旧环境变量）
//   - ledongthuc / thuc — 仅 ledongthuc
//   - auto — 优先 ledongthuc，失败或空再 pdf-go
func pdfEngineMode() string {
	e := strings.ToLower(strings.TrimSpace(os.Getenv("LIGHTNINGRAG_PDF_ENGINE")))
	switch e {
	case "pypdf", "pdfgo":
		return "pdfgo"
	case "ledongthuc", "thuc":
		return "ledongthuc"
	case "auto":
		return "auto"
	default:
		return "default"
	}
}
