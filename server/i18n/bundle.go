package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	bundle     *Bundle
	bundleOnce sync.Once
	bundleErr  error
)

type Bundle struct {
	messages map[string]map[string]string
}

func loadBundle() (*Bundle, error) {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return nil, err
	}
	out := &Bundle{messages: make(map[string]map[string]string)}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := localeFS.ReadFile("locales/" + e.Name())
		if err != nil {
			return nil, err
		}
		lang := strings.TrimSuffix(e.Name(), ".json")
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("i18n locales/%s: %w", e.Name(), err)
		}
		out.messages[lang] = m
	}
	return out, nil
}

func ensureBundle() error {
	bundleOnce.Do(func() {
		bundle, bundleErr = loadBundle()
	})
	return bundleErr
}

// T resolves a message for locale with fallbacks: exact locale → zh-CN (if zh-TW) → en → key.
func T(locale, key string) string {
	if err := ensureBundle(); err != nil || bundle == nil {
		return key
	}
	if s := lookup(bundle.messages, locale, key); s != "" {
		return s
	}
	if locale == "zh-TW" {
		if s := lookup(bundle.messages, "zh-CN", key); s != "" {
			return s
		}
	}
	if s := lookup(bundle.messages, DefaultLocale, key); s != "" {
		return s
	}
	return key
}

func lookup(messages map[string]map[string]string, locale, key string) string {
	if m, ok := messages[locale]; ok {
		if s, ok2 := m[key]; ok2 && s != "" {
			return s
		}
	}
	return ""
}

// Tf formats a translated template with fmt.Sprintf.
func Tf(locale, key string, args ...any) string {
	return fmt.Sprintf(T(locale, key), args...)
}
