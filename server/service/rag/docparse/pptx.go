package docparse

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// 对齐 references/ragflow/deepdoc/parser/ppt_parser.py：从 OOXML pptx 抽取幻灯片文本（按 slide 顺序拼接）。
var rePptxAText = regexp.MustCompile(`<a:t[^>]*>([\s\S]*?)</a:t>`)

// ParsePPTXText 从 .pptx（ZIP + OOXML）提取可见文本；非 zip 或非法 pptx 返回错误。
func ParsePPTXText(data []byte) (string, error) {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return "", fmt.Errorf("不是有效的 pptx（需为 OOXML zip 包；旧版 .ppt 请转换为 .pptx）")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 pptx zip 失败: %w", err)
	}

	type slide struct {
		n    int
		name string
	}
	var slides []slide
	for _, f := range zr.File {
		base := path.Base(f.Name)
		dir := path.Dir(f.Name)
		if dir != "ppt/slides" || !strings.HasPrefix(base, "slide") || !strings.HasSuffix(base, ".xml") {
			continue
		}
		mid := strings.TrimSuffix(strings.TrimPrefix(base, "slide"), ".xml")
		num, err := strconv.Atoi(mid)
		if err != nil {
			continue
		}
		slides = append(slides, slide{n: num, name: f.Name})
	}
	if len(slides) == 0 {
		return "", fmt.Errorf("pptx 中未找到幻灯片 XML")
	}
	sort.Slice(slides, func(i, j int) bool { return slides[i].n < slides[j].n })

	notesBySlide := loadPPTXNotesBySlideNum(zr)

	var sb strings.Builder
	for _, s := range slides {
		rc, err := zr.Open(s.name)
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			continue
		}
		text := extractPptxSlideText(string(raw))
		text = strings.TrimSpace(text)
		if note := strings.TrimSpace(notesBySlide[s.n]); note != "" {
			if text != "" {
				text = text + "\n[Speaker notes]\n" + note
			} else {
				text = "[Speaker notes]\n" + note
			}
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(text)
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("pptx 未提取到文本")
	}
	return out, nil
}

func loadPPTXNotesBySlideNum(zr *zip.Reader) map[int]string {
	out := make(map[int]string)
	for _, f := range zr.File {
		base := path.Base(f.Name)
		dir := path.Dir(f.Name)
		if dir != "ppt/notesSlides" || !strings.HasPrefix(base, "notesSlide") || !strings.HasSuffix(base, ".xml") {
			continue
		}
		mid := strings.TrimSuffix(strings.TrimPrefix(base, "notesSlide"), ".xml")
		num, err := strconv.Atoi(mid)
		if err != nil {
			continue
		}
		rc, err := zr.Open(f.Name)
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			continue
		}
		t := strings.TrimSpace(extractPptxSlideText(string(raw)))
		if t != "" {
			out[num] = t
		}
	}
	return out
}

func extractPptxSlideText(xmlStr string) string {
	matches := rePptxAText.FindAllStringSubmatch(xmlStr, -1)
	var parts []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		t := unescapeXMLText(m[1])
		t = strings.TrimSpace(t)
		if t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, "\n")
}

func unescapeXMLText(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&apos;", "'")
	return s
}
