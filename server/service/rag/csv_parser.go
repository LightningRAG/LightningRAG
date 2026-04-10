package rag

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// guessCSVDelimiter 按首行逗号/分号数量粗判（欧洲 CSV 常用 ';'）。
func guessCSVDelimiter(data []byte) rune {
	line := data
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		line = data[:i]
	}
	commas := bytes.Count(line, []byte{','})
	semi := bytes.Count(line, []byte{';'})
	if semi > commas {
		return ';'
	}
	return ','
}

// ParseCSVContent 从 CSV 文件提取文本
func ParseCSVContent(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 CSV 数据失败: %w", err)
	}

	if !utf8.Valid(data) {
		data = bytes.ToValidUTF8(data, []byte(" "))
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = guessCSVDelimiter(data)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	var sb strings.Builder
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		line := strings.Join(record, "\t")
		if strings.TrimSpace(line) != "" {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("CSV 未提取到文本")
	}
	return result, nil
}

// ParseTSVContent 从 TSV（制表符分隔）提取文本，与 CSV 相同展平策略。
func ParseTSVContent(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("读取 TSV 数据失败: %w", err)
	}

	if !utf8.Valid(data) {
		data = bytes.ToValidUTF8(data, []byte(" "))
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = '\t'
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	var sb strings.Builder
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		line := strings.Join(record, "\t")
		if strings.TrimSpace(line) != "" {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("TSV 未提取到文本")
	}
	return result, nil
}
