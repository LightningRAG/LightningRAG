package channel

import (
	"fmt"
	"sort"
	"sync"
)

var (
	adaptersMu sync.RWMutex
	adapters   = make(map[string]Adapter)
)

// Register 注册渠道适配器（进程启动时由各适配器 init 调用）
func Register(kind string, a Adapter) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	adapters[kind] = a
}

// Lookup 按渠道类型获取适配器
func Lookup(kind string) (Adapter, error) {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	a, ok := adapters[kind]
	if !ok {
		return nil, fmt.Errorf("channel: unknown adapter %q", kind)
	}
	return a, nil
}

// RegisteredKinds 返回当前进程已注册的渠道标识（字典序），供管理端下拉与后端 Register 对齐。
func RegisteredKinds() []string {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	out := make([]string, 0, len(adapters))
	for k := range adapters {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
