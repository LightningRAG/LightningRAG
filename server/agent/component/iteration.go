package component

import (
	"fmt"
	"strings"
	"sync"
)

func init() {
	Register("Iteration", NewIteration)
}

// iterationCurrentKey 迭代当前段的全局变量名，下游组件可引用 {iteration.current}
const iterationCurrentKey = "iteration.current"

// Iteration 迭代组件：按分隔符拆分输入，对每段执行下游组件，合并结果
type Iteration struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewIteration 创建 Iteration 组件
func NewIteration(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Iteration{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (i *Iteration) ComponentName() string {
	return "Iteration"
}

// Invoke 执行：拆分输入，逐段调用下游，合并输出
func (i *Iteration) Invoke(inputs map[string]any) error {
	i.mu.Lock()
	i.err = ""
	i.mu.Unlock()

	inputRef := getStrParam(i.params, "input")
	if inputRef == "" {
		inputRef = "sys.query"
	}
	inputText := ""
	if v, ok := i.canvas.GetVariableValue(inputRef); ok && v != nil {
		inputText = fmt.Sprint(v)
	} else {
		inputText = i.canvas.ResolveString(NormalizeSingleRefForResolve(inputRef))
	}

	delimiter := i.getDelimiter()
	segments := i.split(inputText, delimiter)
	if len(segments) == 0 {
		i.mu.Lock()
		i.output["formalized_content"] = ""
		i.mu.Unlock()
		return nil
	}

	downstreamID := getStrParam(i.params, "downstream")
	if downstreamID == "" {
		downs := i.canvas.GetComponentDownstream(i.id)
		if len(downs) > 0 {
			downstreamID = downs[0]
		}
	}
	if downstreamID == "" {
		i.mu.Lock()
		i.err = "Iteration 未配置 downstream"
		i.mu.Unlock()
		return fmt.Errorf("Iteration 未配置 downstream")
	}

	var results []string
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		i.canvas.SetVariableValue(iterationCurrentKey, seg)
		if err := i.canvas.InvokeComponent(downstreamID); err != nil {
			i.mu.Lock()
			i.err = err.Error()
			i.mu.Unlock()
			return err
		}
		comp, _ := i.canvas.GetComponent(downstreamID)
		if comp != nil {
			if c := comp.Output("content"); c != nil {
				results = append(results, fmt.Sprint(c))
			}
		}
	}

	merged := strings.Join(results, "\n\n")
	i.mu.Lock()
	i.output["formalized_content"] = merged
	// 跳过已执行的下游，next_id 指向下游的下游（Iteration 内部已执行 downstream）
	if nextDowns := i.canvas.GetComponentDownstream(downstreamID); len(nextDowns) > 0 {
		i.output["next_id"] = nextDowns[0]
	}
	i.mu.Unlock()
	return nil
}

func (i *Iteration) getDelimiter() string {
	if custom := getStrParam(i.params, "delimiter_text"); getStrParam(i.params, "delimiter") == "custom" && custom != "" {
		return custom
	}
	d := getStrParam(i.params, "delimiter")
	switch strings.ToLower(d) {
	case "newline", "line_break", "\n":
		return "\n"
	case "semicolon", ";":
		return ";"
	case "comma", ",":
		return ","
	case "tab", "\t":
		return "\t"
	case "dash", "-":
		return "-"
	case "slash", "/":
		return "/"
	case "underline", "_":
		return "_"
	default:
		if d != "" {
			return d
		}
		return ","
	}
}

func (i *Iteration) split(s, delim string) []string {
	if delim == "" {
		return []string{s}
	}
	return strings.Split(s, delim)
}

// Output 获取输出
func (i *Iteration) Output(key string) any {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.output[key]
}

// OutputAll 获取所有输出
func (i *Iteration) OutputAll() map[string]any {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range i.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (i *Iteration) SetOutput(key string, value any) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.output[key] = value
}

// Error 返回错误
func (i *Iteration) Error() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.err
}

// Reset 重置
func (i *Iteration) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.output = make(map[string]any)
	i.err = ""
}
