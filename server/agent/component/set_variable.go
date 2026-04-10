package component

import (
	"fmt"
	"strings"
	"sync"
)

func init() {
	Register("SetVariable", NewSetVariable)
}

// SetVariable 将解析后的值写入全局变量（如 sys.xxx），便于下游节点引用
type SetVariable struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewSetVariable 创建 SetVariable 组件
func NewSetVariable(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &SetVariable{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (s *SetVariable) ComponentName() string { return "SetVariable" }

func (s *SetVariable) Invoke(map[string]any) error {
	s.mu.Lock()
	s.err = ""
	s.mu.Unlock()

	assignments := parseAssignments(s.params["assignments"])
	if len(assignments) == 0 {
		s.mu.Lock()
		s.err = "assignments 不能为空"
		s.mu.Unlock()
		return fmt.Errorf("assignments 不能为空")
	}

	summary := make([]string, 0, len(assignments))
	for _, a := range assignments {
		key := strings.TrimSpace(a.Key)
		if key == "" {
			continue
		}
		if strings.Contains(key, "@") {
			s.mu.Lock()
			s.err = "变量名不能包含 @，请使用 sys.xxx 形式的全局键"
			s.mu.Unlock()
			return fmt.Errorf("非法变量键: %s", key)
		}
		val := s.canvas.ResolveString(a.Value)
		s.canvas.SetVariableValue(key, val)
		summary = append(summary, key+"="+val)
	}

	out := strings.Join(summary, "\n")
	s.mu.Lock()
	s.output["formalized_content"] = out
	s.output["summary"] = out
	s.mu.Unlock()
	return nil
}

func (s *SetVariable) Output(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.output[key]
}

func (s *SetVariable) OutputAll() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range s.output {
		out[k] = v
	}
	return out
}

func (s *SetVariable) SetOutput(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output[key] = value
}

func (s *SetVariable) Error() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

func (s *SetVariable) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = make(map[string]any)
	s.err = ""
}

type assignmentKV struct {
	Key   string
	Value string
}

func parseAssignments(v any) []assignmentKV {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []assignmentKV
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := getStrParam(m, "key")
		val := getStrParam(m, "value")
		if key == "" {
			continue
		}
		out = append(out, assignmentKV{Key: key, Value: val})
	}
	return out
}
