package rag

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	reHTMLTag    = regexp.MustCompile(`<[^>]+>`)
	reWhitespace = regexp.MustCompile(`\s{3,}`)
	reHTMLEntity = regexp.MustCompile(`&[a-zA-Z]+;|&#[0-9]+;`)
)

// ParseHTMLContent 从 HTML 文件提取纯文本（去除标签）
func ParseHTMLContent(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 HTML 数据失败: %w", err)
	}

	if !utf8.Valid(data) {
		data = bytes.ToValidUTF8(data, []byte(" "))
	}

	text := string(data)

	// 移除 script 和 style 标签及其内容
	text = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(text, " ")

	// 块级标签转换为换行
	for _, tag := range []string{"</p>", "</div>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>", "<br>", "<br/>", "<br />", "</li>", "</tr>"} {
		text = strings.ReplaceAll(text, tag, "\n")
		text = strings.ReplaceAll(text, strings.ToUpper(tag), "\n")
	}

	text = reHTMLTag.ReplaceAllString(text, " ")

	// 解码常见 HTML 实体
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = reHTMLEntity.ReplaceAllString(text, " ")

	text = reWhitespace.ReplaceAllString(text, "\n")
	text = strings.TrimSpace(text)

	if text == "" {
		return "", fmt.Errorf("HTML 未提取到文本")
	}
	return text, nil
}
