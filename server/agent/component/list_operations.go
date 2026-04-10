package component

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func init() {
	Register("ListOperations", NewListOperations)
}

// ListOperations 对齐上游编排：对列表做 topN / head / tail / filter / sort / drop_duplicates
type ListOperations struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewListOperations 创建 ListOperations
func NewListOperations(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &ListOperations{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (l *ListOperations) ComponentName() string { return "ListOperations" }

func (l *ListOperations) Invoke(map[string]any) error {
	l.mu.Lock()
	l.err = ""
	l.mu.Unlock()

	raw := l.resolveInput()
	list, err := coerceToAnySlice(raw)
	if err != nil {
		l.fail(err.Error())
		return err
	}

	op := strings.TrimSpace(strings.ToLower(getStrParam(l.params, "operation")))
	if op == "" {
		op = "topn"
	}

	var result []any
	switch op {
	case "topn", "top_n":
		n := getIntParam(l.params, "n", 10)
		if n < 0 {
			n = 0
		}
		if n > len(list) {
			n = len(list)
		}
		result = append([]any(nil), list[:n]...)
	case "head":
		n := getIntParam(l.params, "n", 1)
		if n < 0 {
			n = 0
		}
		if n > len(list) {
			n = len(list)
		}
		result = append([]any(nil), list[:n]...)
	case "tail":
		n := getIntParam(l.params, "n", 1)
		if n < 0 {
			n = 0
		}
		if n > len(list) {
			n = len(list)
		}
		result = append([]any(nil), list[len(list)-n:]...)
	case "filter":
		result, err = l.applyFilter(list)
	case "sort":
		result, err = l.applySort(list)
	case "drop_duplicates":
		result = l.dropDuplicates(list)
	default:
		err = fmt.Errorf("不支持的 operation: %s", op)
	}
	if err != nil {
		l.fail(err.Error())
		return err
	}

	first, last := any(nil), any(nil)
	if len(result) > 0 {
		first = result[0]
		last = result[len(result)-1]
	}

	b, _ := json.Marshal(result)
	summary := string(b)
	if len(summary) > 2000 {
		summary = summary[:2000] + "..."
	}

	l.mu.Lock()
	l.output["result"] = result
	l.output["first"] = first
	l.output["last"] = last
	l.output["formalized_content"] = summary
	l.mu.Unlock()
	return nil
}

func (l *ListOperations) resolveInput() any {
	ref := getStrParam(l.params, "input")
	if ref == "" {
		ref = getStrParam(l.params, "query")
	}
	if ref != "" {
		if v, ok := l.canvas.GetVariableValue(ref); ok {
			return v
		}
	}
	if lit := getStrParam(l.params, "input_literal"); lit != "" {
		return lit
	}
	return nil
}

func coerceToAnySlice(v any) ([]any, error) {
	if v == nil {
		return []any{}, nil
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return []any{}, nil
		}
		var arr []any
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			return arr, nil
		}
		lines := strings.Split(s, "\n")
		var out []any
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				out = append(out, line)
			}
		}
		return out, nil
	}
	if arr, ok := v.([]any); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("ListOperations 输入需为 JSON 数组或可解析的字符串")
}

func (l *ListOperations) applyFilter(list []any) ([]any, error) {
	field := getStrParam(l.params, "field")
	want := getStrParam(l.params, "value")
	op := strings.TrimSpace(getStrParam(l.params, "filter_operator"))
	if op == "" {
		op = "="
	}
	var out []any
	for _, item := range list {
		got := fieldValue(item, field)
		if matchFilter(got, want, op) {
			out = append(out, item)
		}
	}
	return out, nil
}

func fieldValue(item any, field string) string {
	if field == "" || field == "." {
		return strings.TrimSpace(fmt.Sprint(item))
	}
	m, ok := item.(map[string]any)
	if !ok {
		return ""
	}
	if v, ok := m[field]; ok {
		return strings.TrimSpace(fmt.Sprint(v))
	}
	return ""
}

func matchFilter(got, want, op string) bool {
	switch op {
	case "=", "eq":
		return got == want
	case "≠", "!=", "ne":
		return got != want
	case "contains":
		return strings.Contains(got, want)
	case "start with", "startswith", "starts with":
		return strings.HasPrefix(got, want)
	case "end with", "endswith", "ends with":
		return strings.HasSuffix(got, want)
	default:
		return got == want
	}
}

func (l *ListOperations) applySort(list []any) ([]any, error) {
	by := strings.ToLower(strings.TrimSpace(getStrParam(l.params, "sort_by")))
	if by == "" {
		by = "letter"
	}
	order := strings.ToLower(strings.TrimSpace(getStrParam(l.params, "sort_order")))
	if order == "" {
		order = "asc"
	}
	desc := order == "desc" || order == "descending"

	out := append([]any(nil), list...)
	sort.SliceStable(out, func(i, j int) bool {
		ai := fmt.Sprint(out[i])
		aj := fmt.Sprint(out[j])
		if by == "numeric" || by == "number" {
			ni, ei := strconv.ParseFloat(ai, 64)
			nj, ej := strconv.ParseFloat(aj, 64)
			if ei == nil && ej == nil {
				if desc {
					return ni > nj
				}
				return ni < nj
			}
		}
		if desc {
			return ai > aj
		}
		return ai < aj
	})
	return out, nil
}

func (l *ListOperations) dropDuplicates(list []any) []any {
	key := getStrParam(l.params, "dedupe_key")
	seen := make(map[string]struct{})
	var out []any
	for _, item := range list {
		k := fieldValue(item, key)
		if k == "" {
			b, _ := json.Marshal(item)
			k = string(b)
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, item)
	}
	return out
}

func (l *ListOperations) fail(msg string) {
	l.mu.Lock()
	l.err = msg
	l.mu.Unlock()
}

func (l *ListOperations) Output(key string) any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.output[key]
}

func (l *ListOperations) OutputAll() map[string]any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range l.output {
		out[k] = v
	}
	return out
}

func (l *ListOperations) SetOutput(key string, value any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output[key] = value
}

func (l *ListOperations) Error() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.err
}

func (l *ListOperations) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = make(map[string]any)
	l.err = ""
}
