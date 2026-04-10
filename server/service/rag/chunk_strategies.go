package rag

import (
	"regexp"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/model/rag"
)

// ChunkConfig 切片配置
type ChunkConfig struct {
	Method    string // general|qa|book|paper|laws|presentation|table|one
	ChunkSize int
	Overlap   int
	Delimiter string // 自定义分段标识符
}

// ChunkConfigFromKB 从知识库配置构建切片配置
func ChunkConfigFromKB(kb *rag.RagKnowledgeBase) ChunkConfig {
	cfg := ChunkConfig{
		Method:    kb.ChunkMethod,
		ChunkSize: kb.ChunkSize,
		Overlap:   kb.ChunkOverlap,
		Delimiter: kb.Delimiter,
	}
	if cfg.Method == "" {
		cfg.Method = rag.ChunkMethodGeneral
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 500
	}
	if cfg.Overlap < 0 || cfg.Overlap >= cfg.ChunkSize {
		cfg.Overlap = 50
	}
	if cfg.Delimiter == "" {
		cfg.Delimiter = `\n!?。；！？`
	}
	return cfg
}

// ChunkDocument 根据配置对文档内容进行切片
func ChunkDocument(content string, cfg ChunkConfig) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	switch strings.ToLower(cfg.Method) {
	case rag.ChunkMethodOne:
		return chunkOne(content)
	case rag.ChunkMethodQA:
		return chunkQA(content, cfg.ChunkSize)
	case rag.ChunkMethodBook:
		return chunkBook(content, cfg.ChunkSize, cfg.Overlap)
	case rag.ChunkMethodPaper:
		return chunkPaper(content, cfg.ChunkSize, cfg.Overlap)
	case rag.ChunkMethodLaws:
		return chunkLaws(content, cfg.ChunkSize, cfg.Overlap)
	case rag.ChunkMethodPresentation:
		return chunkPresentation(content, cfg.ChunkSize)
	case rag.ChunkMethodTable:
		return chunkTable(content, cfg.ChunkSize)
	default:
		return chunkGeneral(content, cfg.ChunkSize, cfg.Overlap, cfg.Delimiter)
	}
}

// ---- chunkOne: 整篇文档作为一个切片 ----

func chunkOne(content string) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	return []string{content}
}

// ---- chunkGeneral: 通用切片 - 按自定义分隔符分段再合并 ----

func chunkGeneral(content string, chunkSize, overlap int, delimiter string) []string {
	segments := splitByDelimiters(content, delimiter)
	return mergeSegments(segments, chunkSize, overlap)
}

// splitByDelimiters 按分隔符字符列表拆分文本
func splitByDelimiters(text, delimiter string) []string {
	delims := parseDelimiterString(delimiter)
	if len(delims) == 0 {
		return splitParagraphs(text)
	}

	var segments []string
	var current strings.Builder

	runes := []rune(text)
	for _, r := range runes {
		current.WriteRune(r)
		if containsRune(delims, r) {
			s := strings.TrimSpace(current.String())
			if s != "" {
				segments = append(segments, s)
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

// parseDelimiterString 解析分隔符字符串（支持 \n \t 等转义）
func parseDelimiterString(s string) []rune {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\r`, "\r")
	return []rune(s)
}

func containsRune(runes []rune, target rune) bool {
	for _, r := range runes {
		if r == target {
			return true
		}
	}
	return false
}

// mergeSegments 将小段合并到 chunkSize 内，支持 overlap
func mergeSegments(segments []string, chunkSize, overlap int) []string {
	if len(segments) == 0 {
		return nil
	}
	var chunks []string
	var current strings.Builder

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		segRunes := []rune(seg)

		if current.Len() == 0 {
			if len(segRunes) > chunkSize {
				chunks = append(chunks, splitByChars(seg, chunkSize, overlap)...)
			} else {
				current.WriteString(seg)
			}
			continue
		}

		if len([]rune(current.String()))+1+len(segRunes) <= chunkSize {
			current.WriteString("\n")
			current.WriteString(seg)
		} else {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			olap := extractOverlap(current.String(), overlap)
			current.Reset()
			if olap != "" {
				current.WriteString(olap)
				current.WriteString("\n")
			}
			if len(segRunes) > chunkSize {
				chunks = append(chunks, splitByChars(seg, chunkSize, overlap)...)
			} else {
				current.WriteString(seg)
			}
		}
	}
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	return filterEmpty(chunks)
}

// ---- chunkQA: Q&A 对提取 ----

var reQASimple = regexp.MustCompile(`(?mi)^(Q|问|问题|Question)\s*[:：]\s*`)

func chunkQA(content string, chunkSize int) []string {
	if !reQASimple.MatchString(content) {
		return chunkGeneral(content, chunkSize, 0, `\n!?。；！？`)
	}

	var chunks []string
	lines := strings.Split(content, "\n")
	var qBuf, aBuf strings.Builder
	inQ := false
	inA := false

	flush := func() {
		q := strings.TrimSpace(qBuf.String())
		a := strings.TrimSpace(aBuf.String())
		if q != "" {
			chunk := "Q: " + q
			if a != "" {
				chunk += "\nA: " + a
			}
			chunks = append(chunks, chunk)
		}
		qBuf.Reset()
		aBuf.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if matched := regexp.MustCompile(`(?i)^(Q|问|问题|Question)\s*[:：]\s*`).FindStringIndex(trimmed); matched != nil {
			flush()
			qBuf.WriteString(trimmed[matched[1]:])
			inQ = true
			inA = false
		} else if matched := regexp.MustCompile(`(?i)^(A|答|回答|Answer)\s*[:：]\s*`).FindStringIndex(trimmed); matched != nil {
			aBuf.WriteString(trimmed[matched[1]:])
			inQ = false
			inA = true
		} else {
			if inA {
				aBuf.WriteString("\n")
				aBuf.WriteString(trimmed)
			} else if inQ {
				qBuf.WriteString(" ")
				qBuf.WriteString(trimmed)
			}
		}
	}
	flush()

	return filterEmpty(chunks)
}

// ---- chunkBook: 按标题层级分块 ----

var reHeading = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
var reNumberedHeading = regexp.MustCompile(`(?m)^(\d+\.(?:\d+\.)*)\s+(.+)$`)

func chunkBook(content string, chunkSize, overlap int) []string {
	sections := splitByHeadings(content)
	if len(sections) <= 1 {
		return chunkGeneral(content, chunkSize, overlap, `\n!?。；！？`)
	}
	return mergeSectionsToChunks(sections, chunkSize, overlap)
}

func splitByHeadings(content string) []string {
	lines := strings.Split(content, "\n")
	var sections []string
	var current strings.Builder

	for _, line := range lines {
		if reHeading.MatchString(line) || reNumberedHeading.MatchString(line) {
			if current.Len() > 0 {
				sections = append(sections, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		sections = append(sections, strings.TrimSpace(current.String()))
	}
	return sections
}

func mergeSectionsToChunks(sections []string, chunkSize, overlap int) []string {
	var chunks []string
	for _, sec := range sections {
		sec = strings.TrimSpace(sec)
		if sec == "" {
			continue
		}
		if len([]rune(sec)) <= chunkSize {
			chunks = append(chunks, sec)
		} else {
			chunks = append(chunks, splitByChars(sec, chunkSize, overlap)...)
		}
	}
	return filterEmpty(chunks)
}

// ---- chunkPaper: 论文按章节分块 ----

var rePaperSection = regexp.MustCompile(`(?mi)^(?:(?:\d+\.?\s+)|(?:#{1,3}\s+))?(Abstract|Introduction|Related Work|Methodology|Method|Methods|Experiments?|Results?|Discussion|Conclusion|References|Acknowledgements?|Appendix|摘要|引言|方法|实验|结果|讨论|结论|参考文献|致谢)\s*$`)

func chunkPaper(content string, chunkSize, overlap int) []string {
	lines := strings.Split(content, "\n")
	var sections []string
	var current strings.Builder

	for _, line := range lines {
		if rePaperSection.MatchString(strings.TrimSpace(line)) || reHeading.MatchString(line) {
			if current.Len() > 0 {
				sections = append(sections, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		sections = append(sections, strings.TrimSpace(current.String()))
	}

	if len(sections) <= 1 {
		return chunkGeneral(content, chunkSize, overlap, `\n!?。；！？`)
	}
	return mergeSectionsToChunks(sections, chunkSize, overlap)
}

// ---- chunkLaws: 法律文档按条款分块 ----

var reLawArticle = regexp.MustCompile(`(?m)^(?:第[一二三四五六七八九十百千零\d]+[条章节款项]|Article\s+\d+|Section\s+\d+|§\s*\d+)`)

func chunkLaws(content string, chunkSize, overlap int) []string {
	lines := strings.Split(content, "\n")
	var articles []string
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if reLawArticle.MatchString(trimmed) {
			if current.Len() > 0 {
				articles = append(articles, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		articles = append(articles, strings.TrimSpace(current.String()))
	}

	if len(articles) <= 1 {
		return chunkGeneral(content, chunkSize, overlap, `\n!?。；！？`)
	}
	return mergeSectionsToChunks(articles, chunkSize, overlap)
}

// ---- chunkPresentation: 按页/幻灯片分块 ----

var rePageBreak = regexp.MustCompile(`(?m)^-{3,}$|^={3,}$|^---\s*$|^\f`)

func chunkPresentation(content string, chunkSize int) []string {
	pages := rePageBreak.Split(content, -1)

	if len(pages) <= 1 {
		pages = strings.Split(content, "\n\n\n")
	}

	var chunks []string
	for i, page := range pages {
		page = strings.TrimSpace(page)
		if page == "" {
			continue
		}
		if len([]rune(page)) > chunkSize {
			sub := splitByChars(page, chunkSize, 0)
			for j, s := range sub {
				chunks = append(chunks, addPagePrefix(s, i+1, j > 0))
			}
		} else {
			chunks = append(chunks, addPagePrefix(page, i+1, false))
		}
	}
	return filterEmpty(chunks)
}

func addPagePrefix(text string, pageNum int, isContinuation bool) string {
	return strings.TrimSpace(text)
}

// ---- chunkTable: 按行分块 ----

func chunkTable(content string, chunkSize int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	header := ""
	startIdx := 0
	if len(lines) > 1 {
		header = strings.TrimSpace(lines[0])
		startIdx = 1
		if startIdx < len(lines) && isSeparatorLine(lines[startIdx]) {
			startIdx++
		}
	}

	var chunks []string
	var current strings.Builder
	if header != "" {
		current.WriteString(header)
		current.WriteString("\n")
	}

	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || isSeparatorLine(line) {
			continue
		}

		testLen := len([]rune(current.String())) + 1 + len([]rune(line))
		if current.Len() > 0 && testLen > chunkSize {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
			if header != "" {
				current.WriteString(header)
				current.WriteString("\n")
			}
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	return filterEmpty(chunks)
}

func isSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		if r != '-' && r != '=' && r != '|' && r != '+' && r != ' ' {
			return false
		}
	}
	return true
}

// ---- 工具函数 ----

func filterEmpty(chunks []string) []string {
	var result []string
	for _, c := range chunks {
		c = strings.TrimSpace(c)
		if c != "" {
			result = append(result, c)
		}
	}
	return result
}
