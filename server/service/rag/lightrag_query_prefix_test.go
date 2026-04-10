package rag

import "testing"

func TestParseLightningRAGQueryPrefix(t *testing.T) {
	tests := []struct {
		in        string
		wantClean string
		wantMode  string
		wantCtx   bool
	}{
		{"/local 什么是 RAG", "什么是 RAG", "local", false},
		{"/naive foo", "foo", "vector", false},
		{"/mix bar", "bar", "mix", false},
		{"/bypass x", "x", "bypass", false},
		{"/localcontext q", "q", "local", true},
		{"/naivecontext q", "q", "vector", true},
		{"/context", "", "mix", true},
		{"/context hello", "hello", "mix", true},
		{"/context\thello", "hello", "mix", true},
		{"no prefix", "no prefix", "", false},
	}
	for _, tt := range tests {
		c, m, ctx := ParseLightningRAGQueryPrefix(tt.in)
		if c != tt.wantClean || m != tt.wantMode || ctx != tt.wantCtx {
			t.Errorf("ParseLightningRAGQueryPrefix(%q) = (%q,%q,%v), want (%q,%q,%v)",
				tt.in, c, m, ctx, tt.wantClean, tt.wantMode, tt.wantCtx)
		}
	}
}

func TestResolveLightningRAGQueryModeAndQuestion(t *testing.T) {
	q, mode, ctx := ResolveLightningRAGQueryModeAndQuestion("/local  hi", "hybrid")
	if q != "hi" || mode != "hybrid" || ctx {
		t.Fatalf("queryMode field should win: got q=%q mode=%q ctx=%v", q, mode, ctx)
	}
	q2, mode2, ctx2 := ResolveLightningRAGQueryModeAndQuestion("plain", "mix")
	if q2 != "plain" || mode2 != "mix" || ctx2 {
		t.Fatalf("explicit mode without prefix: got q=%q mode=%q ctx=%v", q2, mode2, ctx2)
	}
}
