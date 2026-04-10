package component

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

func init() {
	Register("StringTransform", NewStringTransform)
}

// StringTransform 对齐上游编排：split（多分隔符）/ merge（模板）
type StringTransform struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewStringTransform 创建 StringTransform
func NewStringTransform(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &StringTransform{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (s *StringTransform) ComponentName() string { return "StringTransform" }

func (s *StringTransform) Invoke(map[string]any) error {
	s.mu.Lock()
	s.err = ""
	s.mu.Unlock()

	mode := strings.ToLower(strings.TrimSpace(getStrParam(s.params, "mode")))
	if mode == "" {
		mode = "split"
	}

	var result any
	var err error
	switch mode {
	case "split":
		result, err = s.doSplit()
	case "merge":
		result, err = s.doMerge()
	default:
		err = fmt.Errorf("不支持的 mode: %s", mode)
	}
	if err != nil {
		s.fail(err.Error())
		return err
	}

	outStr := ""
	switch t := result.(type) {
	case []any:
		b, _ := json.Marshal(t)
		outStr = string(b)
	case string:
		outStr = t
	default:
		outStr = fmt.Sprint(t)
	}

	s.mu.Lock()
	s.output["result"] = result
	s.output["formalized_content"] = outStr
	s.mu.Unlock()
	return nil
}

func (s *StringTransform) resolveStringParam(key string) string {
	ref := getStrParam(s.params, key)
	if ref == "" {
		return ""
	}
	if v, ok := s.canvas.GetVariableValue(ref); ok {
		return strings.TrimSpace(fmt.Sprint(v))
	}
	return s.canvas.ResolveString(ref)
}

func (s *StringTransform) doSplit() (any, error) {
	text := s.resolveStringParam("input")
	if text == "" {
		text = getStrParam(s.params, "input_literal")
	}
	dels := parseStringListParam(s.params["delimiters"])
	if len(dels) == 0 {
		dels = []string{","}
	}
	var parts []string
	if len(dels) == 1 {
		parts = strings.Split(text, dels[0])
	} else {
		escaped := make([]string, len(dels))
		for i, d := range dels {
			escaped[i] = regexp.QuoteMeta(d)
		}
		pat := strings.Join(escaped, "|")
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		parts = re.Split(text, -1)
	}
	var out []any
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *StringTransform) doMerge() (string, error) {
	tpl := getStrParam(s.params, "template")
	if tpl == "" {
		tpl = "{0}"
	}
	vars := parseMergeVariables(s.params["merge_variables"])
	out := tpl
	for k, ref := range vars {
		val := ""
		if ref != "" {
			if v, ok := s.canvas.GetVariableValue(ref); ok {
				val = fmt.Sprint(v)
			} else {
				val = s.canvas.ResolveString(ref)
			}
		}
		out = strings.ReplaceAll(out, "{"+k+"}", val)
	}
	return out, nil
}

func parseStringListParam(v any) []string {
	if v == nil {
		return nil
	}
	if arr, ok := v.([]any); ok {
		var s []string
		for _, x := range arr {
			s = append(s, strings.TrimSpace(fmt.Sprint(x)))
		}
		return s
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		var arr []string
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			return arr
		}
		return []string{s}
	}
	return nil
}

func parseMergeVariables(v any) map[string]string {
	out := make(map[string]string)
	if v == nil {
		return out
	}
	if m, ok := v.(map[string]any); ok {
		for k, val := range m {
			out[k] = strings.TrimSpace(fmt.Sprint(val))
		}
		return out
	}
	arr, ok := v.([]any)
	if !ok {
		return out
	}
	for _, x := range arr {
		item, ok := x.(map[string]any)
		if !ok {
			continue
		}
		k := getStrParam(item, "key")
		if k == "" {
			continue
		}
		if ref := getStrParam(item, "ref"); ref != "" {
			out[k] = ref
		} else if val := getStrParam(item, "value"); val != "" {
			out[k] = val
		}
	}
	return out
}

func (s *StringTransform) fail(msg string) {
	s.mu.Lock()
	s.err = msg
	s.mu.Unlock()
}

func (s *StringTransform) Output(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.output[key]
}

func (s *StringTransform) OutputAll() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]any)
	for k, v := range s.output {
		cp[k] = v
	}
	return cp
}

func (s *StringTransform) SetOutput(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output[key] = value
}

func (s *StringTransform) Error() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

func (s *StringTransform) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = make(map[string]any)
	s.err = ""
}
