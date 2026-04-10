package component

import (
	"fmt"
	"strings"
	"sync"
)

func init() {
	Register("VariableAggregator", NewVariableAggregator)
}

// VariableAggregator 对齐上游编排：每组 variables 按顺序取第一个「非空」值，写入以 group_name 为名的输出
type VariableAggregator struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewVariableAggregator 创建 VariableAggregator
func NewVariableAggregator(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &VariableAggregator{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (v *VariableAggregator) ComponentName() string { return "VariableAggregator" }

func (v *VariableAggregator) Invoke(map[string]any) error {
	v.mu.Lock()
	v.err = ""
	v.mu.Unlock()

	groups := parseAggregatorGroups(v.params["groups"])
	if len(groups) == 0 {
		v.fail("groups 不能为空")
		return fmt.Errorf("groups 不能为空")
	}

	var summary []string
	for _, g := range groups {
		name := strings.TrimSpace(g.Name)
		if name == "" {
			v.fail("group_name 不能为空")
			return fmt.Errorf("group_name 不能为空")
		}
		var picked any
		var pickedFrom string
		for _, sel := range g.Selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			val, ok := v.canvas.GetVariableValue(sel)
			if !ok || val == nil {
				continue
			}
			if isEmptyValue(val) {
				continue
			}
			picked = val
			pickedFrom = sel
			break
		}
		v.mu.Lock()
		v.output[name] = picked
		v.mu.Unlock()
		if pickedFrom != "" {
			summary = append(summary, name+"<-"+pickedFrom)
		} else {
			summary = append(summary, name+":<empty>")
		}
	}

	v.mu.Lock()
	v.output["formalized_content"] = strings.Join(summary, "\n")
	v.mu.Unlock()
	return nil
}

type aggGroup struct {
	Name      string
	Selectors []string
}

func parseAggregatorGroups(v any) []aggGroup {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []aggGroup
	for _, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		g := aggGroup{Name: getStrParam(m, "group_name")}
		raw := m["variables"]
		sl, ok := raw.([]any)
		if !ok {
			continue
		}
		for _, item := range sl {
			switch t := item.(type) {
			case map[string]any:
				if ref := getStrParam(t, "value"); ref != "" {
					g.Selectors = append(g.Selectors, ref)
				}
			case string:
				if strings.TrimSpace(t) != "" {
					g.Selectors = append(g.Selectors, strings.TrimSpace(t))
				}
			}
		}
		out = append(out, g)
	}
	return out
}

func isEmptyValue(v any) bool {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case []any:
		return len(t) == 0
	case map[string]any:
		return len(t) == 0
	case float64:
		return t == 0
	case int:
		return t == 0
	case bool:
		return !t
	default:
		return false
	}
}

func (v *VariableAggregator) fail(msg string) {
	v.mu.Lock()
	v.err = msg
	v.mu.Unlock()
}

func (v *VariableAggregator) Output(key string) any {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.output[key]
}

func (v *VariableAggregator) OutputAll() map[string]any {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make(map[string]any)
	for k, val := range v.output {
		out[k] = val
	}
	return out
}

func (v *VariableAggregator) SetOutput(key string, value any) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.output[key] = value
}

func (v *VariableAggregator) Error() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.err
}

func (v *VariableAggregator) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.output = make(map[string]any)
	v.err = ""
}
