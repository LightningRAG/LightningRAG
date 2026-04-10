package rag

import (
	"strings"
	"unicode/utf8"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// 知识图谱写入侧长度约束，借鉴 references/LightRAG/lightrag/constants.py：
// DEFAULT_ENTITY_NAME_MAX_LENGTH、描述/关键词在合并后无限增长会影响库表与向量文本。

const (
	kgFallbackMaxEntityNameRunes       = 256 // DEFAULT_ENTITY_NAME_MAX_LENGTH
	kgFallbackStoredDescriptionRunes = 16384
	kgFallbackStoredKeywordsRunes    = 4096
	kgMaxEntityTypeRunes             = 128 // 与 gorm size:128 一致
)

func kgTruncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) > max {
		if max == 1 {
			return string(r[:1])
		}
		return string(r[:max-1]) + "…"
	}
	return s
}

func effectiveKgMaxEntityNameRunes() int {
	c := global.LRAG_CONFIG.Rag.KgEntityNameMaxRunes
	if c <= 0 {
		return kgFallbackMaxEntityNameRunes
	}
	if c > 512 {
		return 512
	}
	return c
}

func effectiveKgStoredDescriptionMaxRunes() int {
	c := global.LRAG_CONFIG.Rag.KgStoredDescriptionMaxRunes
	if c <= 0 {
		return kgFallbackStoredDescriptionRunes
	}
	if c > 200000 {
		return 200000
	}
	return c
}

func effectiveKgStoredKeywordsMaxRunes() int {
	c := global.LRAG_CONFIG.Rag.KgStoredKeywordsMaxRunes
	if c <= 0 {
		return kgFallbackStoredKeywordsRunes
	}
	if c > 32000 {
		return 32000
	}
	return c
}

func kgClampEntityType(s string) string {
	return kgTruncateRunes(strings.TrimSpace(s), kgMaxEntityTypeRunes)
}

func kgClampStoredDescription(s string) string {
	return kgTruncateRunes(strings.TrimSpace(s), effectiveKgStoredDescriptionMaxRunes())
}

func kgClampStoredKeywords(s string) string {
	return kgTruncateRunes(strings.TrimSpace(s), effectiveKgStoredKeywordsMaxRunes())
}
