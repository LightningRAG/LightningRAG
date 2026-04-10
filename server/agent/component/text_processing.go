package component

import (
	"fmt"
	"strings"
	"sync"
)

func init() {
	Register("TextProcessing", NewTextProcessing)
}

// TextProcessing 文本处理组件：合并或拆分文本
type TextProcessing struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewTextProcessing 创建 TextProcessing 组件
func NewTextProcessing(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &TextProcessing{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (t *TextProcessing) ComponentName() string {
	return "TextProcessing"
}

// Invoke 执行
func (t *TextProcessing) Invoke(inputs map[string]any) error {
	t.mu.Lock()
	t.err = ""
	t.mu.Unlock()

	method := strings.ToLower(getStrParam(t.params, "method"))
	if method == "" {
		method = "split"
	}
	delimiter := t.getDelimiter()

	var result string
	if method == "merge" {
		script := getStrParam(t.params, "script")
		result = t.canvas.ResolveString(script)
	} else {
		splitRef := getStrParam(t.params, "split_ref")
		if splitRef == "" {
			splitRef = "sys.query"
		}
		inputText := ""
		if v, ok := t.canvas.GetVariableValue(splitRef); ok && v != nil {
			inputText = fmt.Sprint(v)
		} else {
			inputText = t.canvas.ResolveString(NormalizeSingleRefForResolve(splitRef))
		}
		segments := strings.Split(inputText, delimiter)
		result = strings.Join(segments, "\n")
	}

	t.mu.Lock()
	t.output["content"] = result
	t.output["result"] = result
	t.mu.Unlock()
	return nil
}

func (t *TextProcessing) getDelimiter() string {
	if custom := getStrParam(t.params, "delimiter_text"); getStrParam(t.params, "delimiter") == "custom" && custom != "" {
		return custom
	}
	d := getStrParam(t.params, "delimiter")
	switch strings.ToLower(d) {
	case "newline", "line_break", "\n":
		return "\n"
	case "semicolon", ";":
		return ";"
	case "comma", ",":
		return ","
	case "tab", "\t":
		return "\t"
	case "space", " ":
		return " "
	default:
		if d != "" {
			return d
		}
		return "\n"
	}
}

// Output 获取输出
func (t *TextProcessing) Output(key string) any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.output[key]
}

// OutputAll 获取所有输出
func (t *TextProcessing) OutputAll() map[string]any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range t.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (t *TextProcessing) SetOutput(key string, value any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output[key] = value
}

// Error 返回错误
func (t *TextProcessing) Error() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.err
}

// Reset 重置
func (t *TextProcessing) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output = make(map[string]any)
	t.err = ""
}
