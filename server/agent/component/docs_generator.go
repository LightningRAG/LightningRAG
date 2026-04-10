package component

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	docx "github.com/mmonterroca/docxgo/v2"
)

//go:embed fonts/DroidSansFallback.ttf
var embeddedDroidSansFallback []byte

// 可通过环境变量指定其它 TrueType 字体（须为 TTF，不支持 OTF/PostScript 轮廓）
const envPDFFontPath = "LIGHTNINGRAG_PDF_FONT"

func init() {
	Register("DocsGenerator", NewDocsGenerator)
}

// DocsGenerator 文档生成组件，支持 Markdown 内容转 PDF/DOCX/TXT
type DocsGenerator struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewDocsGenerator 创建 DocsGenerator 组件
func NewDocsGenerator(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &DocsGenerator{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (d *DocsGenerator) ComponentName() string {
	return "DocsGenerator"
}

// Invoke 执行文档生成
func (d *DocsGenerator) Invoke(inputs map[string]any) error {
	d.mu.Lock()
	d.err = ""
	d.mu.Unlock()

	content := d.canvas.ResolveString(getStrParam(d.params, "content"))
	if content == "" {
		content = getStrParam(d.params, "content")
	}
	title := d.canvas.ResolveString(getStrParam(d.params, "title"))
	if title == "" {
		title = getStrParam(d.params, "title")
	}
	outputFormat := strings.ToLower(getStrParam(d.params, "output_format"))
	if outputFormat == "" {
		outputFormat = "pdf"
	}
	outputDir := d.canvas.ResolveString(getStrParam(d.params, "output_dir"))
	if outputDir == "" {
		outputDir = getStrParam(d.params, "output_dir")
	}
	if outputDir == "" {
		outputDir = filepath.Join(os.TempDir(), "lightningrag_docs")
	}
	filename := d.canvas.ResolveString(getStrParam(d.params, "filename"))
	if filename == "" {
		filename = getStrParam(d.params, "filename")
	}
	if filename == "" {
		filename = fmt.Sprintf("doc_%d", time.Now().Unix())
	}
	// 确保扩展名正确
	ext := "." + outputFormat
	if outputFormat == "docx" {
		ext = ".docx"
	}
	if !strings.HasSuffix(strings.ToLower(filename), ext) {
		filename = filename + ext
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		d.mu.Lock()
		d.err = "创建输出目录失败: " + err.Error()
		d.mu.Unlock()
		return err
	}

	filePath := filepath.Join(outputDir, filename)
	var genErr error

	switch outputFormat {
	case "txt":
		genErr = d.generateTXT(filePath, content)
	case "pdf":
		genErr = d.generatePDF(filePath, content, title)
	case "docx":
		genErr = d.generateDOCX(filePath, content, title)
	default:
		d.mu.Lock()
		d.err = "不支持的输出格式: " + outputFormat + "，支持: pdf, docx, txt"
		d.mu.Unlock()
		return fmt.Errorf("不支持的输出格式: %s", outputFormat)
	}

	if genErr != nil {
		d.mu.Lock()
		d.err = genErr.Error()
		d.mu.Unlock()
		return genErr
	}

	// 读取文件生成 base64（用于 download 展示）
	var pdfBase64 string
	if outputFormat == "pdf" {
		if b, err := os.ReadFile(filePath); err == nil {
			pdfBase64 = base64.StdEncoding.EncodeToString(b)
		}
	}

	downloadJSON, _ := json.Marshal(map[string]any{
		"filename":  filename,
		"file_path": filePath,
		"format":    outputFormat,
		"base64":    pdfBase64,
		"success":   true,
	})

	d.mu.Lock()
	d.output["file_path"] = filePath
	d.output["filename"] = filename
	d.output["success"] = true
	d.output["download"] = string(downloadJSON)
	if pdfBase64 != "" {
		d.output["pdf_base64"] = pdfBase64
	}
	d.mu.Unlock()
	return nil
}

func (d *DocsGenerator) generateTXT(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func (d *DocsGenerator) loadPDFFontBytes() ([]byte, error) {
	if p := strings.TrimSpace(os.Getenv(envPDFFontPath)); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("读取 %s 指向的字体失败: %w", envPDFFontPath, err)
		}
		if len(b) == 0 {
			return nil, fmt.Errorf("%s 字体文件为空", envPDFFontPath)
		}
		return b, nil
	}
	if len(embeddedDroidSansFallback) == 0 {
		return nil, fmt.Errorf("内置中文字体未嵌入，请设置 %s 指向 .ttf 文件", envPDFFontPath)
	}
	return embeddedDroidSansFallback, nil
}

func (d *DocsGenerator) generatePDF(path, content, title string) error {
	fontBytes, err := d.loadPDFFontBytes()
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	// Helvetica 仅覆盖拉丁文；中文等需 UTF-8 TrueType（gofpdf 不支持 OTTO/OTF）
	pdf.AddUTF8FontFromBytes("cjk", "", fontBytes)
	pdf.AddUTF8FontFromBytes("cjk", "B", fontBytes)

	pdf.AddPage()

	const family = "cjk"
	// 标题：无单独粗体文件时用同字库 + 样式 B（部分字形粗体效果有限，主要靠字号区分）
	if title != "" {
		pdf.SetFont(family, "B", 16)
		pdf.MultiCell(0, 8, title, "", "L", false)
		pdf.Ln(4)
	}

	pdf.SetFont(family, "", 11)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			pdf.Ln(5)
			continue
		}
		// Markdown 粗体 **x** 简单弱化：去掉星号，避免乱码符号
		line = simplifyMarkdownLineForPDF(line)
		pdf.MultiCell(0, 6, line, "", "L", false)
	}

	return pdf.OutputFileAndClose(path)
}

// 去掉常见 Markdown 标记，避免在纯文本 PDF 里显示 **、# 等干扰阅读（不做完整 MD 解析）
func simplifyMarkdownLineForPDF(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "### ") {
		s = strings.TrimPrefix(s, "### ")
	} else if strings.HasPrefix(s, "## ") {
		s = strings.TrimPrefix(s, "## ")
	} else if strings.HasPrefix(s, "# ") {
		s = strings.TrimPrefix(s, "# ")
	}
	for strings.Contains(s, "**") {
		s = strings.Replace(s, "**", "", 2)
	}
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		s = strings.Trim(s, "`")
	}
	return strings.TrimSpace(s)
}

func (d *DocsGenerator) generateDOCX(path, content, title string) error {
	doc := docx.NewDocument()

	if title != "" {
		para, err := doc.AddParagraph()
		if err != nil {
			return err
		}
		run, err := para.AddRun()
		if err != nil {
			return err
		}
		run.SetText(title)
		run.SetBold(true)
		run.SetSize(32) // 16pt (half-points)
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		para, err := doc.AddParagraph()
		if err != nil {
			return err
		}
		run, err := para.AddRun()
		if err != nil {
			return err
		}
		run.SetText(line)
	}

	return doc.SaveAs(path)
}

// Output 获取输出
func (d *DocsGenerator) Output(key string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.output[key]
}

// OutputAll 获取所有输出
func (d *DocsGenerator) OutputAll() map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range d.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (d *DocsGenerator) SetOutput(key string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.output[key] = value
}

// Error 返回错误
func (d *DocsGenerator) Error() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.err
}

// Reset 重置
func (d *DocsGenerator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.output = make(map[string]any)
	d.err = ""
}
