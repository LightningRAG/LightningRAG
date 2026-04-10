package docparse

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLToPlainText 将 HTML 转为可读纯文本（供 EML/MSG/EPUB 等复用）。
// 对齐 Ragflow deepdoc HtmlParser 思路：不遍历 script/style/noscript 等标签子树，避免噪声与内联脚本文本。
func HTMLToPlainText(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return strings.TrimSpace(s)
	}
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.CommentNode:
			return
		case html.TextNode:
			sb.WriteString(n.Data)
		case html.ElementNode:
			switch strings.ToLower(n.Data) {
			case "script", "style", "noscript", "iframe", "template", "object", "embed":
				return
			case "br", "p", "div", "tr", "h1", "h2", "h3", "h4", "h5", "h6", "li":
				sb.WriteString("\n")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	lines := strings.Split(sb.String(), "\n")
	var compact []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t != "" {
			compact = append(compact, t)
		}
	}
	return strings.Join(compact, "\n")
}
