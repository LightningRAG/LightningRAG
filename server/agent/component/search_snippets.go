package component

import (
	"fmt"
	"strings"
)

func splitCommaFields(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// webSnippet 网页/文献检索单条结果
type webSnippet struct {
	Title string
	URL   string
	Body  string
}

func formatWebSnippets(snips []webSnippet) string {
	if len(snips) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, s := range snips {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = fmt.Sprintf("Result %d", i+1)
		}
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, title))
		if b := strings.TrimSpace(s.Body); b != "" {
			sb.WriteString(b)
			sb.WriteString("\n")
		}
		if u := strings.TrimSpace(s.URL); u != "" {
			sb.WriteString("URL: ")
			sb.WriteString(u)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}
