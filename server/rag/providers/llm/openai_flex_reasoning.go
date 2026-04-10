package llm

import (
	"bytes"
	"encoding/json"
	"strings"
)

// flexReasoningMaxDepth 防止异常/恶意 JSON 深层嵌套导致栈溢出。
const flexReasoningMaxDepth = 48

// flexOpenAIReasoningFragment 解析 OpenAI 兼容响应中推理相关字段（reasoning_content / reasoning / thought / thinking）：
// 可为 JSON 字符串、对象（如 {"text":"..."}）或字符串/对象数组。
type flexOpenAIReasoningFragment string

func (f *flexOpenAIReasoningFragment) UnmarshalJSON(b []byte) error {
	return f.unmarshalReasoningFragment(bytes.TrimSpace(b), 0)
}

func (f *flexOpenAIReasoningFragment) unmarshalReasoningFragment(b []byte, depth int) error {
	if depth > flexReasoningMaxDepth {
		*f = ""
		return nil
	}
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*f = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexOpenAIReasoningFragment(s)
		return nil
	}
	if b[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal(b, &arr); err != nil {
			*f = ""
			return nil
		}
		var parts []string
		next := depth + 1
		for _, raw := range arr {
			raw = bytes.TrimSpace(raw)
			if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
				continue
			}
			if raw[0] == '"' {
				var s string
				if json.Unmarshal(raw, &s) == nil && s != "" {
					parts = append(parts, s)
				}
				continue
			}
			var sub flexOpenAIReasoningFragment
			if err := sub.unmarshalReasoningFragment(raw, next); err != nil {
				return err
			}
			if sub != "" {
				parts = append(parts, string(sub))
			}
		}
		*f = flexOpenAIReasoningFragment(strings.Join(parts, "\n"))
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		*f = ""
		return nil
	}
	next := depth + 1
	keys := []string{"text", "thinking", "thought", "content", "value", "reasoning", "reasoning_content", "data", "delta", "message"}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				*f = flexOpenAIReasoningFragment(s)
				return nil
			}
			if subm, ok := v.(map[string]any); ok {
				if got := flexParseReasoningSubObjectDepth(subm, next); got != "" {
					*f = got
					return nil
				}
			}
		}
	}
	for _, v := range m {
		if subm, ok := v.(map[string]any); ok {
			if got := flexParseReasoningSubObjectDepth(subm, next); got != "" {
				*f = got
				return nil
			}
		}
	}
	*f = ""
	return nil
}

func flexParseReasoningSubObjectDepth(subm map[string]any, depth int) flexOpenAIReasoningFragment {
	if depth > flexReasoningMaxDepth {
		return ""
	}
	b2, err := json.Marshal(subm)
	if err != nil {
		return ""
	}
	var sub flexOpenAIReasoningFragment
	if err := sub.unmarshalReasoningFragment(b2, depth); err != nil {
		return ""
	}
	return sub
}
