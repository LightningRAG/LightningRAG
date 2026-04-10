package oauth

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	mu        sync.RWMutex
	providers = make(map[string]Provider)
)

// Register 由各平台 init 调用，注册 Provider 实现
func Register(p Provider) {
	if p == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	k := strings.ToLower(strings.TrimSpace(p.Kind()))
	providers[k] = p
}

// Lookup 按 kind 获取已注册实现（大小写不敏感）
func Lookup(kind string) (Provider, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[kind]
	if !ok {
		return nil, fmt.Errorf("oauth: unknown provider %q", kind)
	}
	return p, nil
}

// RegisteredKinds 返回已注册平台标识（字典序）
func RegisteredKinds() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(providers))
	for k := range providers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
