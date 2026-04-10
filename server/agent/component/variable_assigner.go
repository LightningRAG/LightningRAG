package component

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

func init() {
	Register("VariableAssigner", NewVariableAssigner)
}

// VariableAssigner 按 VariableAssigner 语义（references 目录内有同名参考）：对全局变量做覆盖、清空、列表追加、数值运算等
type VariableAssigner struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewVariableAssigner 创建 VariableAssigner
func NewVariableAssigner(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &VariableAssigner{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (v *VariableAssigner) ComponentName() string { return "VariableAssigner" }

func (v *VariableAssigner) Invoke(map[string]any) error {
	v.mu.Lock()
	v.err = ""
	v.mu.Unlock()

	items := parseAssignerItems(v.params["variables"])
	if len(items) == 0 {
		v.mu.Lock()
		v.output["formalized_content"] = "ok"
		v.mu.Unlock()
		return nil
	}

	for _, it := range items {
		if strings.TrimSpace(it.Variable) == "" || strings.TrimSpace(it.Operator) == "" {
			v.fail("variable / operator 不能为空")
			return fmt.Errorf("variable / operator 不能为空")
		}
		if strings.Contains(it.Variable, "@") {
			v.fail("目标 variable 应为全局键，如 sys.xxx")
			return fmt.Errorf("非法目标键")
		}

		cur := v.readGlobal(it.Variable)
		op := strings.TrimSpace(strings.ToLower(it.Operator))
		var next any
		var err error
		switch op {
		case "overwrite":
			next, err = v.resolveRefOrLiteral(it.Parameter)
		case "clear":
			next = clearValue(cur)
		case "set":
			next, err = v.resolveRefOrLiteral(it.Parameter)
		case "append":
			next, err = v.opAppend(cur, it.Parameter)
		case "extend":
			next, err = v.opExtend(cur, it.Parameter)
		case "remove_first":
			next = opRemoveFirst(cur)
		case "remove_last":
			next = opRemoveLast(cur)
		case "+=":
			next, err = v.opAdd(cur, it.Parameter)
		case "-=":
			next, err = v.opSub(cur, it.Parameter)
		case "*=":
			next, err = v.opMul(cur, it.Parameter)
		case "/=":
			next, err = v.opDiv(cur, it.Parameter)
		default:
			v.fail("不支持的操作符: " + it.Operator)
			return fmt.Errorf("不支持的操作符")
		}
		if err != nil {
			v.fail(err.Error())
			return err
		}
		v.canvas.SetVariableValue(it.Variable, next)
	}

	v.mu.Lock()
	v.output["formalized_content"] = "ok"
	v.output["success"] = true
	v.mu.Unlock()
	return nil
}

type assignerItem struct {
	Variable  string
	Operator  string
	Parameter string
}

func parseAssignerItems(v any) []assignerItem {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []assignerItem
	for _, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, assignerItem{
			Variable:  getStrParam(m, "variable"),
			Operator:  getStrParam(m, "operator"),
			Parameter: getStrParam(m, "parameter"),
		})
	}
	return out
}

func (v *VariableAssigner) readGlobal(key string) any {
	if g := v.canvas.GetGlobals(); g != nil {
		if val, ok := g[key]; ok {
			return val
		}
	}
	return nil
}

func (v *VariableAssigner) resolveRefOrLiteral(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	if strings.Contains(s, "{") {
		return v.canvas.ResolveString(s), nil
	}
	if v, ok := v.canvas.GetVariableValue(s); ok {
		return v, nil
	}
	return s, nil
}

func (v *VariableAssigner) resolveParamValue(s string) (any, error) {
	return v.resolveRefOrLiteral(s)
}

func (v *VariableAssigner) opAppend(cur any, param string) (any, error) {
	pv, err := v.resolveParamValue(param)
	if err != nil {
		return nil, err
	}
	list, ok := asAnySlice(cur)
	if !ok {
		return nil, fmt.Errorf("append 要求目标为数组")
	}
	return append(list, pv), nil
}

func (v *VariableAssigner) opExtend(cur any, param string) (any, error) {
	pv, err := v.resolveParamValue(param)
	if err != nil {
		return nil, err
	}
	a, ok1 := asAnySlice(cur)
	b, ok2 := asAnySlice(pv)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("extend 要求目标与参数均为数组")
	}
	out := make([]any, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out, nil
}

func opRemoveFirst(cur any) any {
	list, ok := asAnySlice(cur)
	if !ok || len(list) == 0 {
		return cur
	}
	return list[1:]
}

func opRemoveLast(cur any) any {
	list, ok := asAnySlice(cur)
	if !ok || len(list) == 0 {
		return cur
	}
	return list[:len(list)-1]
}

func (v *VariableAssigner) opAdd(cur any, param string) (any, error) {
	pv, err := v.resolveParamValue(param)
	if err != nil {
		return nil, err
	}
	a, ok1 := toNumber(cur)
	b, ok2 := toNumber(pv)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("+= 要求数值类型")
	}
	return a + b, nil
}

func (v *VariableAssigner) opSub(cur any, param string) (any, error) {
	pv, err := v.resolveParamValue(param)
	if err != nil {
		return nil, err
	}
	a, ok1 := toNumber(cur)
	b, ok2 := toNumber(pv)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("-= 要求数值类型")
	}
	return a - b, nil
}

func (v *VariableAssigner) opMul(cur any, param string) (any, error) {
	pv, err := v.resolveParamValue(param)
	if err != nil {
		return nil, err
	}
	a, ok1 := toNumber(cur)
	b, ok2 := toNumber(pv)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("*= 要求数值类型")
	}
	return a * b, nil
}

func (v *VariableAssigner) opDiv(cur any, param string) (any, error) {
	pv, err := v.resolveParamValue(param)
	if err != nil {
		return nil, err
	}
	a, ok1 := toNumber(cur)
	b, ok2 := toNumber(pv)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("/= 要求数值类型")
	}
	if b == 0 {
		return nil, fmt.Errorf("除以零")
	}
	return a / b, nil
}

func clearValue(cur any) any {
	switch cur.(type) {
	case []any:
		return []any{}
	case string:
		return ""
	case map[string]any:
		return map[string]any{}
	case bool:
		return false
	case float64, int, int64, float32:
		return float64(0)
	case nil:
		return ""
	default:
		return ""
	}
}

func asAnySlice(v any) ([]any, bool) {
	if v == nil {
		return []any{}, true
	}
	if s, ok := v.([]any); ok {
		return s, true
	}
	return nil, false
}

func toNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func (v *VariableAssigner) fail(msg string) {
	v.mu.Lock()
	v.err = msg
	v.mu.Unlock()
}

func (v *VariableAssigner) Output(key string) any {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.output[key]
}

func (v *VariableAssigner) OutputAll() map[string]any {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make(map[string]any)
	for k, val := range v.output {
		out[k] = val
	}
	return out
}

func (v *VariableAssigner) SetOutput(key string, value any) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.output[key] = value
}

func (v *VariableAssigner) Error() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.err
}

func (v *VariableAssigner) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.output = make(map[string]any)
	v.err = ""
}
