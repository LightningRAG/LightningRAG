package docparse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ParseJSONText 对齐 references/ragflow/deepdoc/parser/json_parser.py 的用途：
// 将 JSON / JSONL 转为可检索的纯文本（本实现为整文档规范化文本，切片仍由上游 TextSplitter/ChunkDocument 完成）。
func ParseJSONText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("JSON 数据为空")
	}
	if !utf8.Valid(data) {
		data = bytes.ToValidUTF8(data, []byte(" "))
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return "", fmt.Errorf("JSON 内容为空")
	}
	if isJSONLFormat(s) {
		return parseJSONLToText(data)
	}
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return "", fmt.Errorf("解析 JSON 失败: %w", err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	out := strings.TrimSpace(string(b))
	if out == "" {
		return "", fmt.Errorf("JSON 未产生文本")
	}
	return out, nil
}

// isJSONLFormat 对齐 RAGFlowJsonParser.is_jsonl_format：整段能 parse 成单 JSON 则否；否则抽样行前若干行是否多为独立 JSON。
func isJSONLFormat(txt string) bool {
	lines := strings.Split(strings.TrimSpace(txt), "\n")
	var nonEmpty []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t != "" {
			nonEmpty = append(nonEmpty, t)
		}
	}
	if len(nonEmpty) == 0 {
		return false
	}
	var single any
	if json.Unmarshal([]byte(strings.TrimSpace(txt)), &single) == nil {
		return false
	}
	const sampleLimit = 10
	threshold := 0.8
	n := len(nonEmpty)
	if n > sampleLimit {
		n = sampleLimit
	}
	valid := 0
	for i := 0; i < n; i++ {
		var v any
		if json.Unmarshal([]byte(nonEmpty[i]), &v) == nil {
			valid++
		}
	}
	if valid == 0 {
		return false
	}
	return float64(valid)/float64(n) >= threshold
}

func parseJSONLToText(data []byte) (string, error) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	// 允许较长行
	const max = 32 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, max)

	var lines []string
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var v any
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			return "", fmt.Errorf("JSONL 第 %d 行解析失败: %w", lineNo, err)
		}
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("JSONL 第 %d 行序列化失败: %w", lineNo, err)
		}
		lines = append(lines, string(b))
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("读取 JSONL 失败: %w", err)
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("JSONL 无有效行")
	}
	return strings.Join(lines, "\n\n"), nil
}
