package component

import (
	"sync"
)

func init() {
	Register("Begin", NewBegin)
}

// Begin 入口组件，接收用户输入
type Begin struct {
	id     string
	canvas Canvas
	output map[string]any
	mu     sync.RWMutex
}

// NewBegin 创建 Begin 组件
func NewBegin(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Begin{
		id:     id,
		canvas: canvas,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (b *Begin) ComponentName() string {
	return "Begin"
}

// Invoke 执行
func (b *Begin) Invoke(inputs map[string]any) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for k, v := range inputs {
		if val, ok := v.(map[string]any); ok && val != nil {
			if vv, ok := val["value"]; ok {
				b.output[k] = vv
			} else {
				b.output[k] = v
			}
		} else {
			b.output[k] = v
		}
	}
	// 同步到 globals
	if q, ok := inputs["query"]; ok {
		if m, ok := q.(map[string]any); ok && m["value"] != nil {
			b.canvas.SetVariableValue("sys.query", m["value"])
		}
	}
	return nil
}

// Output 获取输出
func (b *Begin) Output(key string) any {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.output[key]
}

// OutputAll 获取所有输出
func (b *Begin) OutputAll() map[string]any {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range b.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (b *Begin) SetOutput(key string, value any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output[key] = value
}

// Error 返回错误
func (b *Begin) Error() string {
	return ""
}

// Reset 重置
func (b *Begin) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output = make(map[string]any)
}
