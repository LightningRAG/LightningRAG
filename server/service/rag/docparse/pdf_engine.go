package docparse

import (
	"os"
	"strings"
)

// pdfEngineMode 解析 LIGHTNINGRAG_PDF_ENGINE：
//   - 未设置或其它值 — default：优先 pypdf，失败或正文为空再 ledongthuc
//   - pypdf — 仅 pypdf
//   - ledongthuc / thuc — 仅 ledongthuc
//   - auto — 优先 ledongthuc，失败或空再 pypdf（兼容旧行为）
func pdfEngineMode() string {
	e := strings.ToLower(strings.TrimSpace(os.Getenv("LIGHTNINGRAG_PDF_ENGINE")))
	switch e {
	case "pypdf":
		return "pypdf"
	case "ledongthuc", "thuc":
		return "ledongthuc"
	case "auto":
		return "auto"
	default:
		return "default"
	}
}
