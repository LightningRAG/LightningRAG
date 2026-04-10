package component

import (
	"fmt"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/agent/dsl"
)

// Factory 组件工厂函数
type Factory func(canvas Canvas, id string, params map[string]any) (Component, error)

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register 注册组件
func Register(name string, f Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = f
}

// Create 根据 component_name 创建组件实例
func Create(canvas Canvas, id string, def dsl.ComponentDef) (Component, error) {
	name := def.Obj.ComponentName
	if name == "" {
		return nil, fmt.Errorf("component_name is empty")
	}
	registryMu.RLock()
	f, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown component: %s", name)
	}
	params := def.Obj.Params
	if params == nil {
		params = make(map[string]any)
	}
	return f(canvas, id, params)
}
