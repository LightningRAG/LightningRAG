// Package component 参数解析工具
package component

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// standaloneVarTokenRe 与画布变量 token 形态一致：裸 sys.* / 组件@字段 / iteration.* / env.*
var standaloneVarTokenRe = regexp.MustCompile(`^(?i)(sys\.[a-z0-9_.]+|[a-z][a-z0-9_]*@[a-z0-9_.-]+|iteration\.[a-z0-9_.]+|env\.[a-z0-9_.]+)$`)

// NormalizeSingleRefForResolve 将「整段字符串」规范为 ResolveString 可替换的形式：
// 已含 { 视为模板或已带花括号的引用，原样返回；否则若为独立变量 token，则外包一层 { }；普通字面量（如自然语言检索词）不变。
func NormalizeSingleRefForResolve(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if strings.Contains(raw, "{") {
		return raw
	}
	if standaloneVarTokenRe.MatchString(raw) {
		return "{" + raw + "}"
	}
	return raw
}

// getBoolParam 解析布尔参数，缺省为 false
func getBoolParam(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(b, "true") || b == "1"
	}
	return false
}

// getOptionalBoolPtrParam 从 params 解析可选布尔；键均不存在或无法解析时返回 nil（表示不覆盖默认）
func getOptionalBoolPtrParam(m map[string]any, keys ...string) *bool {
	for _, key := range keys {
		v, ok := m[key]
		if !ok || v == nil {
			continue
		}
		switch b := v.(type) {
		case bool:
			x := b
			return &x
		case string:
			l := strings.ToLower(strings.TrimSpace(b))
			if l == "true" || l == "1" || l == "yes" {
				t := true
				return &t
			}
			if l == "false" || l == "0" || l == "no" {
				f := false
				return &f
			}
		case float64:
			x := b != 0
			return &x
		}
	}
	return nil
}

// getBoolParamDefault 解析布尔参数，缺省为 def
func getBoolParamDefault(m map[string]any, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(b, "true") || b == "1"
	}
	return def
}

// resolveWorkflowQuery 解析搜索类 query：支持含 {…} 的模板，或裸变量名如 sys.query（读 globals）
func resolveWorkflowQuery(cv Canvas, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "{") {
		return strings.TrimSpace(cv.ResolveString(raw))
	}
	if v, ok := cv.GetVariableValue(raw); ok && v != nil {
		s := strings.TrimSpace(fmt.Sprint(v))
		if s != "" {
			return s
		}
	}
	return strings.TrimSpace(cv.ResolveString(NormalizeSingleRefForResolve(raw)))
}

// getStrParam 从 params 获取字符串
func getStrParam(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

// getIntParam 从 params 获取整数
func getIntParam(m map[string]any, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return def
}

// getFloatParam 从 params 获取浮点数
func getFloatParam(m map[string]any, key string, def float64) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return def
}

// resolveComponentLLMConfig 从组件 params 中解析 LLM 配置，
// 当 api_key 为空且 Canvas 有 LLMConfigResolver 时，自动从数据库查询补全
func resolveComponentLLMConfig(cv Canvas, params map[string]any) (provider, modelName, baseURL, apiKey string) {
	provider = getStrParam(params, "provider")
	modelName = getStrParam(params, "model_name")
	baseURL = getStrParam(params, "base_url")
	apiKey = getStrParam(params, "api_key")

	if provider == "" || modelName == "" {
		llmID := getStrParam(params, "llm_id")
		if llmID != "" {
			if idx := strings.Index(llmID, "@"); idx > 0 {
				provider = strings.TrimSpace(llmID[:idx])
				modelName = strings.TrimSpace(llmID[idx+1:])
			} else {
				modelName = llmID
				provider = "ollama"
			}
		}
	}
	if provider == "" {
		provider = "ollama"
	}
	if modelName == "" {
		modelName = "llama3.2"
	}

	if apiKey == "" && cv != nil {
		if resolver := cv.GetLLMConfigResolver(); resolver != nil {
			cfg, err := resolver(context.Background(), cv.GetTenantID(), provider, modelName)
			if err == nil && cfg != nil {
				apiKey = cfg.APIKey
				if baseURL == "" && cfg.BaseURL != "" {
					baseURL = cfg.BaseURL
				}
			}
		}
	}
	return
}

// getUintSliceParam 从 params 获取 uint 切片
func getUintSliceParam(m map[string]any, key string) []uint {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []uint
	for _, item := range arr {
		switch n := item.(type) {
		case float64:
			out = append(out, uint(n))
		case string:
			u, _ := strconv.ParseUint(n, 10, 32)
			out = append(out, uint(u))
		}
	}
	return out
}

// getStrSliceParam 从 params 获取字符串切片（JSON 数组）
func getStrSliceParam(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range arr {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}
