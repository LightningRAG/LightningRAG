package i18n

import (
	"strings"
)

// IsProviderAPIKeyError reports whether err likely indicates an invalid or expired
// upstream LLM/embedding API key. Upstream bodies are often localized (e.g. Chinese);
// we map these to rag.error.model_api_key_rejected in FailWithError.
func IsProviderAPIKeyError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	lower := strings.ToLower(s)

	if strings.Contains(s, "请检查密钥是否正确配置或已过期") {
		return true
	}
	if strings.Contains(s, "当前") && strings.Contains(s, "无效") {
		if strings.Contains(lower, "api-key") || strings.Contains(lower, "api key") {
			return true
		}
	}

	if strings.Contains(lower, "invalid_api_key") || strings.Contains(lower, "authentication_error") {
		return true
	}
	if strings.Contains(lower, "incorrect api key") || strings.Contains(lower, "invalid api key") {
		return true
	}

	modelHTTP := strings.Contains(lower, "embedding error") ||
		strings.Contains(lower, "openai api error") ||
		strings.Contains(lower, "azure embedding error") ||
		strings.Contains(lower, "cohere embedding error") ||
		strings.Contains(lower, "voyage embedding error") ||
		strings.Contains(lower, "jina embedding error") ||
		strings.Contains(lower, "nvidia embedding error") ||
		strings.Contains(lower, "volcengine embedding error") ||
		strings.Contains(lower, "ollama embedding error") ||
		strings.Contains(lower, "replicate embed") ||
		strings.Contains(lower, "replicate:")
	if modelHTTP {
		if strings.Contains(lower, " 401") || strings.Contains(lower, ": 401") ||
			strings.HasSuffix(lower, "401") || strings.Contains(lower, "\"status\":401") {
			return true
		}
		if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "access denied") {
			return true
		}
		if strings.Contains(lower, "invalid subscription key") || strings.Contains(lower, "wrong api endpoint") {
			return true
		}
	}
	return false
}
