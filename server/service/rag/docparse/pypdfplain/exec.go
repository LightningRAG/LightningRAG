// Package pypdfplain 通过子进程调用内嵌的 pypdf_plain_extract.py，使用仓库 references/pypdf 解析 PDF（对齐 Ragflow PlainParser）。
package pypdfplain

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 环境变量（与 references/pypdf + Ragflow PlainParser 对齐）见仓库文档；
// 常用：LIGHTNINGRAG_PYTHON、LIGHTNINGRAG_PYPDF_SRC、LIGHTNINGRAG_REPO_ROOT、
// LIGHTNINGRAG_PDF_PASSWORD、LIGHTNINGRAG_PYPDF_EXTRACTION_MODE、页范围与 full 模式附加块等。

const extractTimeout = 3 * time.Minute

//go:embed pypdf_plain_extract.py
var plainExtractScript string

var (
	scriptPath     string
	scriptPathErr  error
	scriptPathOnce sync.Once
)

// PythonExecutable 返回用于运行脚本的解释器（LIGHTNINGRAG_PYTHON 或默认 python3 / Windows py）。
func PythonExecutable() string {
	if py := strings.TrimSpace(os.Getenv("LIGHTNINGRAG_PYTHON")); py != "" {
		return py
	}
	if runtime.GOOS == "windows" {
		return "py"
	}
	return "python3"
}

func scriptOnDisk() (string, error) {
	scriptPathOnce.Do(func() {
		dir := filepath.Join(os.TempDir(), "lightningrag-pypdf")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			scriptPathErr = err
			return
		}
		path := filepath.Join(dir, "plain_extract.py")
		if err := os.WriteFile(path, []byte(plainExtractScript), 0o644); err != nil {
			scriptPathErr = err
			return
		}
		scriptPath = path
	})
	return scriptPath, scriptPathErr
}

// FindPypdfSourceRoot 定位 references/pypdf（内含 pypdf/__init__.py）。
func FindPypdfSourceRoot() (string, error) {
	try := func(dir string) (string, bool) {
		if dir == "" {
			return "", false
		}
		candidate := filepath.Join(dir, "pypdf", "__init__.py")
		fi, err := os.Stat(candidate)
		if err != nil || fi.IsDir() {
			return "", false
		}
		return filepath.Clean(dir), true
	}
	if d := strings.TrimSpace(os.Getenv("LIGHTNINGRAG_PYPDF_SRC")); d != "" {
		if root, ok := try(d); ok {
			return root, nil
		}
		return "", fmt.Errorf("LIGHTNINGRAG_PYPDF_SRC 无效：缺少 pypdf/__init__.py: %s", d)
	}
	if repo := strings.TrimSpace(os.Getenv("LIGHTNINGRAG_REPO_ROOT")); repo != "" {
		if root, ok := try(filepath.Join(repo, "references", "pypdf")); ok {
			return root, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for i := 0; i < 12; i++ {
			if root, ok := try(filepath.Join(dir, "references", "pypdf")); ok {
				return root, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 16; i++ {
		if root, ok := try(filepath.Join(dir, "references", "pypdf")); ok {
			return root, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("未找到 references/pypdf，请设置 LIGHTNINGRAG_PYPDF_SRC 或 LIGHTNINGRAG_REPO_ROOT")
}

func pypdfCmd(script, root string, subcommandAndRange ...string) (string, []string) {
	py := PythonExecutable()
	if runtime.GOOS == "windows" && py == "py" {
		args := append([]string{"-3", script, root}, subcommandAndRange...)
		return py, args
	}
	args := append([]string{script, root}, subcommandAndRange...)
	return py, args
}

func pageRangeArgsFromEnv() []string {
	fromE := strings.TrimSpace(os.Getenv("LIGHTNINGRAG_PDF_FROM_PAGE"))
	toE := strings.TrimSpace(os.Getenv("LIGHTNINGRAG_PDF_TO_PAGE"))
	if fromE == "" && toE == "" {
		return nil
	}
	fromN := 0
	if fromE != "" {
		if v, err := strconv.Atoi(fromE); err == nil {
			fromN = v
		}
	}
	if fromN < 0 {
		fromN = 0
	}
	toN := int(^uint(0) >> 1)
	if toE != "" {
		if v, err := strconv.Atoi(toE); err == nil {
			toN = v
		}
	}
	if toN < 0 {
		toN = 0
	}
	return []string{strconv.Itoa(fromN), strconv.Itoa(toN)}
}

func scriptArgs(sub string) []string {
	a := []string{sub}
	a = append(a, pageRangeArgsFromEnv()...)
	return a
}

func runScript(data []byte, subcommandAndRange ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("PDF 数据为空")
	}
	root, err := FindPypdfSourceRoot()
	if err != nil {
		return nil, err
	}
	script, err := scriptOnDisk()
	if err != nil {
		return nil, err
	}
	py, args := pypdfCmd(script, root, subcommandAndRange...)
	ctx, cancel := context.WithTimeout(context.Background(), extractTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, py, args...)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Env = os.Environ()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("pypdf 子进程失败: %w (%s)", err, msg)
		}
		return nil, fmt.Errorf("pypdf 子进程失败: %w", err)
	}
	return out, nil
}

// ExtractPlainFull 全文 + 书签（PlainParser 等价）。
func ExtractPlainFull(data []byte) (string, error) {
	out, err := runScript(data, scriptArgs("full")...)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "", fmt.Errorf("pypdf 未提取到文本（可能是扫描件或图片型 PDF，需 OCR）")
	}
	return s, nil
}

type metaJSON struct {
	PageCount int `json:"page_count"`
}

// ExtractPageCount 使用 pypdf 的页数（与 len(reader.pages) 一致）。
func ExtractPageCount(data []byte) (int, error) {
	out, err := runScript(data, "meta")
	if err != nil {
		return 0, err
	}
	var m metaJSON
	if err := json.Unmarshal(bytes.TrimSpace(out), &m); err != nil {
		return 0, fmt.Errorf("pypdf meta JSON: %w", err)
	}
	if m.PageCount < 0 {
		return 0, fmt.Errorf("pypdf 返回非法页数")
	}
	return m.PageCount, nil
}

type pagesJSON struct {
	Pages []string `json:"pages"`
}

// ExtractPages 按页 extract_text，与 Ragflow 逐页 split 思路一致（页间无额外 \n\n）。
func ExtractPages(data []byte) ([]string, error) {
	out, err := runScript(data, scriptArgs("pages")...)
	if err != nil {
		return nil, err
	}
	var w pagesJSON
	if err := json.Unmarshal(bytes.TrimSpace(out), &w); err != nil {
		return nil, fmt.Errorf("pypdf pages JSON: %w", err)
	}
	return w.Pages, nil
}

// DocInfo 为 pypdf PdfReader 的 pdf_header 与 DocumentInformation 字段（JSON 与 docinfo 子命令一致）。
type DocInfo struct {
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

// ExtractDocInfo 读取文档信息字典（非 XMP）；无 Python/pypdf 时返回错误。
func ExtractDocInfo(data []byte) (*DocInfo, error) {
	out, err := runScript(data, "docinfo")
	if err != nil {
		return nil, err
	}
	var info DocInfo
	if err := json.Unmarshal(bytes.TrimSpace(out), &info); err != nil {
		return nil, fmt.Errorf("pypdf docinfo JSON: %w", err)
	}
	return &info, nil
}

type linksJSON struct {
	Links []string `json:"links"`
}

// ExtractURILinks 提取页面注释中的 /URI（对齐 Ragflow extract_links_from_pdf）。
func ExtractURILinks(data []byte) ([]string, error) {
	out, err := runScript(data, "links")
	if err != nil {
		return nil, err
	}
	var w linksJSON
	if err := json.Unmarshal(bytes.TrimSpace(out), &w); err != nil {
		return nil, fmt.Errorf("pypdf links JSON: %w", err)
	}
	return w.Links, nil
}

type pageLabelsJSON struct {
	PageLabels []string `json:"page_labels"`
}

// ExtractPageLabels 返回每页逻辑页码标签（对齐 pypdf PdfReader.page_labels）。
func ExtractPageLabels(data []byte) ([]string, error) {
	out, err := runScript(data, "pagelabels")
	if err != nil {
		return nil, err
	}
	var w pageLabelsJSON
	if err := json.Unmarshal(bytes.TrimSpace(out), &w); err != nil {
		return nil, fmt.Errorf("pypdf pagelabels JSON: %w", err)
	}
	return w.PageLabels, nil
}

type attachmentNamesJSON struct {
	AttachmentNames []string `json:"attachment_names"`
}

// ExtractAttachmentNames 返回嵌入附件文件名（去重排序；不含附件内容）。
func ExtractAttachmentNames(data []byte) ([]string, error) {
	out, err := runScript(data, "attachmentnames")
	if err != nil {
		return nil, err
	}
	var w attachmentNamesJSON
	if err := json.Unmarshal(bytes.TrimSpace(out), &w); err != nil {
		return nil, fmt.Errorf("pypdf attachmentnames JSON: %w", err)
	}
	return w.AttachmentNames, nil
}

// ExtractXMPMetadata 返回 pypdf xmp_metadata 摘要字段；JSON 中 xmp 为 null 时返回 (nil, nil)。
func ExtractXMPMetadata(data []byte) (map[string]any, error) {
	out, err := runScript(data, "xmp")
	if err != nil {
		return nil, err
	}
	var w struct {
		XMP *map[string]any `json:"xmp"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), &w); err != nil {
		return nil, fmt.Errorf("pypdf xmp JSON: %w", err)
	}
	if w.XMP == nil {
		return nil, nil
	}
	return *w.XMP, nil
}
