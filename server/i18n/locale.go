package i18n

import (
	"sort"
	"strconv"
	"strings"
)

const DefaultLocale = "en"

// SupportedLocales aligns with web/src/locale/constants.js SUPPORTED_LOCALES.
var SupportedLocales = []string{
	"en", "zh-CN", "zh-TW", "ja", "ko", "fr", "de", "es", "it", "pt-BR", "ru", "vi", "th", "id",
}

// ParseAcceptLanguage picks the best supported locale from an Accept-Language header value.
func ParseAcceptLanguage(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return DefaultLocale
	}
	type tag struct {
		lang string
		q    float64
	}
	var tags []tag
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		sem := strings.Split(part, ";")
		langRaw := strings.TrimSpace(strings.ToLower(sem[0]))
		q := 1.0
		for _, p := range sem[1:] {
			p = strings.TrimSpace(strings.ToLower(p))
			if strings.HasPrefix(p, "q=") {
				if v, err := strconv.ParseFloat(strings.TrimPrefix(p, "q="), 64); err == nil {
					q = v
				}
			}
		}
		tags = append(tags, tag{lang: langRaw, q: q})
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].q > tags[j].q })
	for _, t := range tags {
		if code := normalizeLocaleTag(t.lang); code != "" {
			return code
		}
	}
	return DefaultLocale
}

func normalizeLocaleTag(raw string) string {
	raw = strings.ReplaceAll(raw, "_", "-")
	if raw == "" {
		return ""
	}
	if raw == "zh-cn" || raw == "zh-hans" || raw == "zh" || strings.HasPrefix(raw, "zh-hans") {
		return "zh-CN"
	}
	if raw == "zh-tw" || raw == "zh-hk" || raw == "zh-mo" || strings.HasPrefix(raw, "zh-hant") {
		return "zh-TW"
	}
	if strings.HasPrefix(raw, "ja") {
		return "ja"
	}
	if strings.HasPrefix(raw, "ko") {
		return "ko"
	}
	if strings.HasPrefix(raw, "en") {
		return "en"
	}
	if strings.HasPrefix(raw, "fr") {
		return "fr"
	}
	if strings.HasPrefix(raw, "de") {
		return "de"
	}
	if strings.HasPrefix(raw, "es") {
		return "es"
	}
	if strings.HasPrefix(raw, "it") {
		return "it"
	}
	if raw == "pt-br" || strings.HasPrefix(raw, "pt-br") {
		return "pt-BR"
	}
	if strings.HasPrefix(raw, "pt") {
		return "pt-BR"
	}
	if strings.HasPrefix(raw, "ru") {
		return "ru"
	}
	if strings.HasPrefix(raw, "vi") {
		return "vi"
	}
	if strings.HasPrefix(raw, "th") {
		return "th"
	}
	if strings.HasPrefix(raw, "id") {
		return "id"
	}
	return ""
}

func IsSupportedLocale(code string) bool {
	for _, s := range SupportedLocales {
		if s == code {
			return true
		}
	}
	return false
}
