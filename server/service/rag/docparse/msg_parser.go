package docparse

import (
	"fmt"
	"os"
	"strings"

	msgparser "github.com/willthrom/outlook-msg-parser"
)

// ParseMSGText 解析 Outlook .msg（OLE），对齐 references/ragflow 邮件 setups 中的 msg 后缀。
// 依赖 github.com/willthrom/outlook-msg-parser；需写入临时文件（库 API 限制）。
func ParseMSGText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("msg 数据为空")
	}
	f, err := os.CreateTemp("", "lr-msg-*.msg")
	if err != nil {
		return "", fmt.Errorf("创建临时 msg 文件失败: %w", err)
	}
	path := f.Name()
	defer func() { _ = os.Remove(path) }()

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("写入 msg 失败: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	m, err := msgparser.ParseMsgFile(path)
	if err != nil {
		return "", fmt.Errorf("解析 msg 失败: %w", err)
	}

	var sb strings.Builder
	add := func(k, v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
	}
	add("Subject", m.Subject)
	add("From", strings.TrimSpace(m.FromName+" "+m.FromEmail))
	add("To", strings.TrimSpace(m.ToDisplay))
	if strings.TrimSpace(m.ToDisplay) == "" {
		add("To", strings.TrimSpace(m.To))
	}
	add("Cc", strings.TrimSpace(m.CCDisplay))
	if strings.TrimSpace(m.CCDisplay) == "" {
		add("Cc", strings.TrimSpace(m.CC))
	}
	add("Bcc", strings.TrimSpace(m.BCCDisplay))
	if !m.Date.IsZero() {
		add("Date", m.Date.String())
	}

	body := strings.TrimSpace(m.BodyPlainText)
	if body == "" || body == "No content available" {
		body = strings.TrimSpace(HTMLToPlainText(m.BodyHTML))
	}
	if body == "" || body == "No content available" {
		body = strings.TrimSpace(m.TransportMessageHeaders)
	}
	if body != "" && body != "No content available" {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(body)
	}

	for _, a := range m.Attachments {
		n := strings.TrimSpace(a.Name)
		if n == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("[attachment] ")
		sb.WriteString(n)
	}

	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("msg 未提取到文本")
	}
	return out, nil
}

// IsLikelyMSGOLE 判断字节流是否为常见 OLE 复合文档头（.msg / .doc / .xls 等均为 D0 CF 11 E0）。
func IsLikelyMSGOLE(data []byte) bool {
	return len(data) >= 4 && data[0] == 0xD0 && data[1] == 0xCF && data[2] == 0x11 && data[3] == 0xE0
}

// IsZIP 判断是否为 zip 头（pptx/docx/xlsx）。
func IsZIP(data []byte) bool {
	return len(data) >= 4 && data[0] == 'P' && data[1] == 'K'
}

// SniffExcelKind 区分 xlsx（zip）与 xls（OLE），用于扩展名不可靠时的纠偏。
func SniffExcelKind(data []byte) string {
	switch {
	case IsZIP(data):
		return "xlsx"
	case IsLikelyMSGOLE(data):
		return "xls"
	default:
		return ""
	}
}

// NormalizeExcelFileType 若扩展名与魔数不一致，以魔数为准（例如误命名 .xlsx 的 xls）。
func NormalizeExcelFileType(fileType string, data []byte) string {
	ft := strings.ToLower(strings.TrimSpace(fileType))
	if ft != "xlsx" && ft != "xls" {
		return ft
	}
	k := SniffExcelKind(data)
	if k == "" {
		return ft
	}
	if (ft == "xlsx" && k == "xls") || (ft == "xls" && k == "xlsx") {
		return k
	}
	return ft
}
