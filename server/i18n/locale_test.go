package i18n

import (
	"errors"
	"testing"
)

func TestParseAcceptLanguage(t *testing.T) {
	if ParseAcceptLanguage("") != "en" {
		t.Fatal("empty header should default to en")
	}
	if ParseAcceptLanguage("zh-CN,en;q=0.8") != "zh-CN" {
		t.Fatalf("got %q", ParseAcceptLanguage("zh-CN,en;q=0.8"))
	}
	if ParseAcceptLanguage("en-US,en;q=0.9") != "en" {
		t.Fatalf("got %q", ParseAcceptLanguage("en-US,en;q=0.9"))
	}
	if ParseAcceptLanguage("ja,en;q=0.8") != "ja" {
		t.Fatalf("got %q", ParseAcceptLanguage("ja,en;q=0.8"))
	}
}

func TestT(t *testing.T) {
	if s := T("en", "response.success"); s == "" || s == "response.success" {
		t.Fatalf("missing en bundle: %q", s)
	}
	if s := T("zh-CN", "response.success"); s == "" {
		t.Fatal("missing zh-CN translation")
	}
	if s := T("ja", "response.success"); s == "" || s == "response.success" {
		t.Fatal("ja should fall back to en")
	}
}

func TestIsProviderAPIKeyError(t *testing.T) {
	silicon := errors.New(`openai embedding error: {"error":{"message":"当前 API-KEY 无效，请检查密钥是否正确配置或已过期"}}`)
	if !IsProviderAPIKeyError(silicon) {
		t.Fatal("expected Chinese SiliconFlow-style message to match")
	}
	if IsProviderAPIKeyError(errors.New("unrelated validation failed")) {
		t.Fatal("expected unrelated error not to match")
	}
	if !IsProviderAPIKeyError(errors.New(`openai embedding error: {"error":{"type":"authentication_error"}}`)) {
		t.Fatal("expected authentication_error in embedding response to match")
	}
}

func TestLocaleJSONKeyParityWithEnglish(t *testing.T) {
	if err := ensureBundle(); err != nil {
		t.Fatal(err)
	}
	if bundle == nil {
		t.Fatal("nil bundle")
	}
	en := bundle.messages["en"]
	if len(en) == 0 {
		t.Fatal("empty en bundle")
	}
	for code, m := range bundle.messages {
		if code == DefaultLocale {
			continue
		}
		for k := range en {
			if _, ok := m[k]; !ok {
				t.Errorf("locale %q missing key %q (present in en)", code, k)
			}
		}
		for k := range m {
			if _, ok := en[k]; !ok {
				t.Errorf("locale %q has extra key %q (not in en)", code, k)
			}
		}
	}
}
