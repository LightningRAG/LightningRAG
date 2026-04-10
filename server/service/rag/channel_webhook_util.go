package rag

import (
	"crypto/subtle"
	"fmt"
	"strings"
)

func constantTimeSecretEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func extractChannelAgentContent(out map[string]any) string {
	if out == nil {
		return ""
	}
	if v, ok := out["content"].(string); ok {
		return v
	}
	if arr, ok := out["content"].([]any); ok && len(arr) > 0 {
		var sb strings.Builder
		for i, a := range arr {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(strings.TrimSpace(fmt.Sprint(a)))
		}
		return sb.String()
	}
	return ""
}

func truncateForDiscord(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 1800 {
		return s
	}
	return s[:1800] + "…"
}
