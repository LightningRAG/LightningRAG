package component

import (
	"fmt"
	"strings"
	"sync"
)

func init() {
	Register("Message", NewMessage)
}

// Message 输出组件，将内容展示给用户
type Message struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	mu     sync.RWMutex
}

// NewMessage 创建 Message 组件
func NewMessage(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Message{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (m *Message) ComponentName() string {
	return "Message"
}

// Invoke 执行：解析 content 中的变量引用并输出
func (m *Message) Invoke(inputs map[string]any) error {
	m.mu.Lock()
	contentList := getContentList(m.params)
	var parts []string
	for _, c := range contentList {
		var s string
		switch v := c.(type) {
		case string:
			s = v
		default:
			s = fmt.Sprint(v)
		}
		resolved := m.canvas.ResolveString(s)
		parts = append(parts, resolved)
	}
	content := strings.Join(parts, "\n")
	m.output["content"] = content
	m.mu.Unlock()
	// 流式模式下通过 onChunk 推送内容，确保 RunStream 能返回 Message 组件的输出
	if content != "" {
		if rc := m.canvas.RunContext(); rc != nil && rc.GetStreamCallback() != nil {
			rc.GetStreamCallback()(content)
		}
	}
	return nil
}

// Output 获取输出
func (m *Message) Output(key string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.output[key]
}

// OutputAll 获取所有输出
func (m *Message) OutputAll() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range m.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (m *Message) SetOutput(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.output[key] = value
}

// Error 返回错误
func (m *Message) Error() string {
	return ""
}

// Reset 重置
func (m *Message) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.output = make(map[string]any)
}

// getContentList 从 params 中获取 content，支持 []any 和 []string 类型
func getContentList(params map[string]any) []any {
	v, ok := params["content"]
	if !ok || v == nil {
		return nil
	}
	if arr, ok := v.([]any); ok {
		return arr
	}
	if arr, ok := v.([]string); ok {
		out := make([]any, len(arr))
		for i, s := range arr {
			out[i] = s
		}
		return out
	}
	return nil
}
