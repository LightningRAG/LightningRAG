package component

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

func compareNumeric(a, b string) int {
	fa, ea := strconv.ParseFloat(strings.TrimSpace(a), 64)
	fb, eb := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if ea != nil || eb != nil {
		return strings.Compare(a, b)
	}
	if fa < fb {
		return -1
	}
	if fa > fb {
		return 1
	}
	return 0
}

func init() {
	Register("Switch", NewSwitch)
}

// SwitchCase 单个分支条件
type SwitchCase struct {
	Conditions []SwitchCondition `json:"conditions"`
	Logic      string            `json:"logic"` // AND / OR
	Downstream string            `json:"downstream"`
}

// SwitchCondition 条件项
type SwitchCondition struct {
	Ref   string `json:"ref"`   // 变量引用，如 retrieval_0@formalized_content
	Op    string `json:"op"`    // equals, not_equal, contains, not_contains, is_empty, not_empty, starts_with, ends_with, less_than, greater_than
	Value any    `json:"value"` // 比较值
}

// Switch 条件分支组件，根据规则选择下游
type Switch struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewSwitch 创建 Switch 组件
func NewSwitch(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Switch{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (s *Switch) ComponentName() string {
	return "Switch"
}

// Invoke 执行：评估 cases，选择第一个匹配的 downstream 写入 next_id
func (s *Switch) Invoke(inputs map[string]any) error {
	s.mu.Lock()
	s.err = ""
	s.mu.Unlock()

	cases := s.getCases()
	if len(cases) == 0 {
		s.mu.Lock()
		s.err = "Switch 至少需要一个 case"
		s.mu.Unlock()
		return fmt.Errorf("Switch 至少需要一个 case")
	}

	for _, c := range cases {
		if s.evalCase(c) {
			s.mu.Lock()
			s.output["next_id"] = c.Downstream
			s.output["selected_case"] = c.Downstream
			s.mu.Unlock()
			return nil
		}
	}

	// 无匹配：取第一个 case 的 downstream 作为默认（或可配置 default）
	s.mu.Lock()
	s.output["next_id"] = cases[0].Downstream
	s.output["selected_case"] = cases[0].Downstream
	s.mu.Unlock()
	return nil
}

func (s *Switch) getCases() []SwitchCase {
	v, ok := s.params["cases"]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []SwitchCase
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		var c SwitchCase
		c.Downstream = getStrParam(m, "downstream")
		c.Logic = getStrParam(m, "logic")
		if c.Logic == "" {
			c.Logic = "AND"
		}
		if conds, ok := m["conditions"].([]any); ok {
			for _, cc := range conds {
				cm, ok := cc.(map[string]any)
				if !ok {
					continue
				}
				c.Conditions = append(c.Conditions, SwitchCondition{
					Ref:   getStrParam(cm, "ref"),
					Op:    getStrParam(cm, "op"),
					Value: cm["value"],
				})
			}
		}
		out = append(out, c)
	}
	return out
}

func (s *Switch) evalCase(c SwitchCase) bool {
	if len(c.Conditions) == 0 {
		return false
	}
	logicAnd := strings.ToUpper(c.Logic) == "AND"
	for _, cond := range c.Conditions {
		ok := s.evalCondition(cond)
		if logicAnd && !ok {
			return false
		}
		if !logicAnd && ok {
			return true
		}
	}
	return logicAnd
}

func (s *Switch) evalCondition(cond SwitchCondition) bool {
	var val string
	if v, ok := s.canvas.GetVariableValue(cond.Ref); ok && v != nil {
		val = fmt.Sprint(v)
	} else {
		val = s.canvas.ResolveString(NormalizeSingleRefForResolve(cond.Ref))
	}
	op := strings.ToLower(cond.Op)
	if op == "" {
		op = "equals"
	}
	cmpVal := cond.Value
	var cmpStr string
	if cmpVal != nil {
		cmpStr = fmt.Sprint(cmpVal)
	}

	switch op {
	case "is_empty", "empty":
		return strings.TrimSpace(val) == ""
	case "not_empty":
		return strings.TrimSpace(val) != ""
	case "equals", "eq":
		return val == cmpStr
	case "not_equal", "ne":
		return val != cmpStr
	case "contains":
		return strings.Contains(val, cmpStr)
	case "not_contains":
		return !strings.Contains(val, cmpStr)
	case "starts_with":
		return strings.HasPrefix(val, cmpStr)
	case "ends_with":
		return strings.HasSuffix(val, cmpStr)
	case "less_than", "lt":
		return compareNumeric(val, cmpStr) < 0
	case "less_equal", "le":
		return compareNumeric(val, cmpStr) <= 0
	case "greater_than", "gt":
		return compareNumeric(val, cmpStr) > 0
	case "greater_equal", "ge":
		return compareNumeric(val, cmpStr) >= 0
	default:
		return false
	}
}

// Output 获取输出
func (s *Switch) Output(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.output[key]
}

// OutputAll 获取所有输出
func (s *Switch) OutputAll() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range s.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (s *Switch) SetOutput(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output[key] = value
}

// Error 返回错误
func (s *Switch) Error() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

// Reset 重置
func (s *Switch) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = make(map[string]any)
	s.err = ""
}
