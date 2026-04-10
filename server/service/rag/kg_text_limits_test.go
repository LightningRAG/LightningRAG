package rag

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/LightningRAG/LightningRAG/server/global"
)

func TestKgTruncateRunes(t *testing.T) {
	if g := kgTruncateRunes("ab", 10); g != "ab" {
		t.Fatalf("%q", g)
	}
	long := strings.Repeat("α", 5) // 5 runes
	out := kgTruncateRunes(long, 3)
	if utf8.RuneCountInString(out) != 3 {
		t.Fatalf("want 3 runes got %d: %q", utf8.RuneCountInString(out), out)
	}
	if !strings.HasSuffix(out, "…") {
		t.Fatal("expected ellipsis")
	}
}

func TestEffectiveKgMaxEntityNameRunes(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.KgEntityNameMaxRunes
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.KgEntityNameMaxRunes = old })
	global.LRAG_CONFIG.Rag.KgEntityNameMaxRunes = 0
	if n := effectiveKgMaxEntityNameRunes(); n != kgFallbackMaxEntityNameRunes {
		t.Fatalf("default: %d", n)
	}
	global.LRAG_CONFIG.Rag.KgEntityNameMaxRunes = 100
	if n := effectiveKgMaxEntityNameRunes(); n != 100 {
		t.Fatalf("override: %d", n)
	}
}
