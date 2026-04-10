package initialize

import (
	"strings"
	"testing"
)

func TestRagApis_matchRagCasbin888(t *testing.T) {
	casbinKeys := make(map[string]struct{})
	for _, r := range ragCasbinRules {
		if r.Ptype != "p" || r.V0 != "888" {
			continue
		}
		casbinKeys[r.V2+"\x00"+r.V1] = struct{}{}
	}
	apiKeys := make(map[string]struct{})
	for _, a := range ragApis {
		apiKeys[a.Method+"\x00"+a.Path] = struct{}{}
	}
	for _, a := range ragApis {
		k := a.Method + "\x00" + a.Path
		if _, ok := casbinKeys[k]; !ok {
			t.Errorf("ragApis 缺少对应 Casbin 规则 (888): %s %s", a.Method, a.Path)
		}
	}
	for _, r := range ragCasbinRules {
		if r.Ptype != "p" || r.V0 != "888" {
			continue
		}
		k := r.V2 + "\x00" + r.V1
		if _, ok := apiKeys[k]; !ok {
			t.Errorf("ragCasbinRules(888) 在 ragApis 中无对应项: %s %s", r.V2, r.V1)
		}
	}
}

func TestRagCasbinRules_uniqueP888(t *testing.T) {
	seen := make(map[string]struct{})
	for _, r := range ragCasbinRules {
		if r.Ptype != "p" || r.V0 != "888" {
			continue
		}
		k := r.V2 + "\x00" + r.V1
		if _, ok := seen[k]; ok {
			t.Fatalf("duplicate Casbin rule (888): %s %s", r.V2, r.V1)
		}
		seen[k] = struct{}{}
	}
}

func TestRagApis_uniquePathMethod(t *testing.T) {
	seen := make(map[string]struct{})
	for _, a := range ragApis {
		key := a.Method + "\x00" + a.Path
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate path+method: %s %s", a.Method, a.Path)
		}
		seen[key] = struct{}{}
		if a.Path == "" || !strings.HasPrefix(a.Path, "/") {
			t.Fatalf("path must be non-empty and start with /: method=%q path=%q", a.Method, a.Path)
		}
		if a.Method == "" {
			t.Fatalf("method must be non-empty: path=%s", a.Path)
		}
	}
}
