package rag

import (
	"regexp"
	"strings"
	"unicode"
)

// NormalizeLightningRAGRetrieverMode 校验并规范化检索模式（与 references/LightRAG QueryParam.mode + 本项目扩展一致）
func NormalizeLightningRAGRetrieverMode(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "global_" {
		return "global"
	}
	switch s {
	case "naive":
		// 与 vector 合并为同一向量切片检索（对齐 references/ragflow 向量腿语义；兼容旧 API / LightRAG 别名）
		return "vector"
	case "vector", "local", "global", "hybrid", "mix", "bypass", "keyword", "pageindex":
		return s
	default:
		return ""
	}
}

var lightragBracketPrefixRE = regexp.MustCompile(`^/([a-z]*)\[(.*?)\](.*)`)

// ParseLightningRAGQueryPrefix 解析 Ollama 兼容前缀（见 references/LightRAG/lightrag/api/routers/ollama_api.py parse_query_mode）
// 返回：去掉前缀后的查询正文、由前缀得到的模式（空表示未指定）、是否仅要上下文（对应 *context 与 /context）
func ParseLightningRAGQueryPrefix(query string) (cleanedQuery string, modeFromPrefix string, onlyNeedContext bool) {
	q := query
	if m := lightragBracketPrefixRE.FindStringSubmatch(q); len(m) == 4 {
		modePrefix := m[1]
		rest := strings.TrimLeft(m[3], " ")
		if modePrefix != "" {
			q = "/" + modePrefix + " " + rest
		} else {
			q = rest
		}
	}

	type pref struct {
		prefix string
		mode   string
		ctx    bool
	}
	// 较长前缀优先，与 Python 3.7+ dict 插入顺序一致
	prefs := []pref{
		{"/localcontext", "local", true},
		{"/globalcontext", "global", true},
		{"/hybridcontext", "hybrid", true},
		{"/naivecontext", "vector", true},
		{"/mixcontext", "mix", true},
		{"/local ", "local", false},
		{"/global ", "global", false},
		{"/naive ", "vector", false},
		{"/hybrid ", "hybrid", false},
		{"/mix ", "mix", false},
		{"/bypass ", "bypass", false},
	}
	for _, e := range prefs {
		if strings.HasPrefix(q, e.prefix) {
			rest := strings.TrimSpace(q[len(e.prefix):])
			return rest, e.mode, e.ctx
		}
	}
	// /context → mix + only_need_context（须避免误匹配 /contextual）
	if strings.HasPrefix(q, "/context") {
		if len(q) == len("/context") {
			return "", "mix", true
		}
		if unicode.IsSpace(rune(q[len("/context")])) {
			return strings.TrimSpace(q[len("/context"):]), "mix", true
		}
	}

	return query, "", false
}

// ResolveLightningRAGQueryModeAndQuestion 合并 JSON 字段 queryMode 与消息前缀；返回用于检索的 query、模式覆盖、是否仅上下文
func ResolveLightningRAGQueryModeAndQuestion(rawContent string, queryModeField string) (question string, modeOverride string, onlyNeedContext bool) {
	cleaned, prefixMode, onlyCtx := ParseLightningRAGQueryPrefix(rawContent)
	modeOverride = NormalizeLightningRAGRetrieverMode(strings.TrimSpace(queryModeField))
	if modeOverride == "" {
		modeOverride = prefixMode
	}
	if prefixMode != "" || onlyCtx {
		question = cleaned
	} else {
		question = rawContent
	}
	onlyNeedContext = onlyCtx
	return question, modeOverride, onlyNeedContext
}
