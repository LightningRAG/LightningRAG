package llm

import (
	"bufio"
	"io"
)

// newStreamLineScanner 解析 SSE / NDJSON 流式行；默认 Scanner 单条上限 64KB，推理模型长 data 行易触发 bufio.ErrTooLong。
func newStreamLineScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	return s
}
