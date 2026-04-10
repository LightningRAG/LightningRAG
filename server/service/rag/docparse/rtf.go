package docparse

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseRTFPlainText 简易 RTF → 纯文本（丢弃控制字；跳过 {\*…} 目的地块）。
func ParseRTFPlainText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("RTF 数据为空")
	}
	s := stripRTFBytes(data)
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("RTF 未提取到文本")
	}
	return s, nil
}

func stripRTFBytes(b []byte) string {
	var out strings.Builder
	out.Grow(len(b) / 4)

	depth := 0
	skipDepth := -1 // >=0 时处于跳过块（如 {\*…}）

	for i := 0; i < len(b); {
		if skipDepth >= 0 && depth < skipDepth {
			skipDepth = -1
		}
		skipping := skipDepth >= 0 && depth >= skipDepth

		c := b[i]
		switch {
		case c == '{':
			depth++
			i++
			if !skipping && i+2 < len(b) && b[i] == '\\' && b[i+1] == '*' {
				skipDepth = depth
			}
		case c == '}':
			if depth > 0 {
				depth--
			}
			if skipDepth > depth {
				skipDepth = -1
			}
			i++
		case skipping:
			i++
		case c == '\\':
			i++
			if i >= len(b) {
				break
			}
			switch b[i] {
			case '\\', '{', '}':
				out.WriteByte(b[i])
				i++
			case '\'':
				if i+2 < len(b) {
					hx := string(b[i+1 : i+3])
					v, err := strconv.ParseUint(hx, 16, 8)
					if err == nil {
						out.WriteByte(byte(v))
					}
					i += 3
				} else {
					i++
				}
			case '\n', '\r':
				i++
			default:
				start := i
				for i < len(b) && isRTFAlpha(b[i]) {
					i++
				}
				word := string(b[start:i])
				for i < len(b) && b[i] >= '0' && b[i] <= '9' {
					i++
				}
				if i < len(b) && b[i] == ' ' {
					i++
				}
				switch word {
				case "par", "line", "row":
					out.WriteByte('\n')
				case "tab":
					out.WriteByte('\t')
				}
			}
		default:
			if c == '\r' || c == '\n' {
				i++
				continue
			}
			if c >= 32 && c < 0x80 {
				out.WriteByte(c)
			}
			i++
		}
	}

	lines := strings.Split(out.String(), "\n")
	var compact []string
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if t != "" {
			compact = append(compact, t)
		}
	}
	return strings.Join(compact, "\n")
}

func isRTFAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
