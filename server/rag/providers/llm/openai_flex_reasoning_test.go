package llm

import (
	"encoding/json"
	"testing"
)

func TestFlexOpenAIReasoningFragmentThinking(t *testing.T) {
	var v struct {
		T flexOpenAIReasoningFragment `json:"thinking"`
	}
	cases := []struct {
		raw  string
		want string
	}{
		{`{"thinking":"hello"}`, "hello"},
		{`{"thinking":{"text":"x"}}`, "x"},
		{`{"thinking":{"thinking":"y"}}`, "y"},
		{`{"thinking":null}`, ""},
		{`{"thinking":["a","b"]}`, "a\nb"},
		{`{"thinking":[{"text":"p1"},{"text":"p2"}]}`, "p1\np2"},
	}
	for _, tc := range cases {
		if err := json.Unmarshal([]byte(tc.raw), &v); err != nil {
			t.Fatalf("%s: %v", tc.raw, err)
		}
		if string(v.T) != tc.want {
			t.Fatalf("%s: got %q want %q", tc.raw, v.T, tc.want)
		}
	}
}

func TestFlexOpenAIReasoningFragmentMaxDepth(t *testing.T) {
	var v struct {
		R flexOpenAIReasoningFragment `json:"reasoning_content"`
	}
	deep := `{"text":"x"}`
	for i := 0; i < flexReasoningMaxDepth+3; i++ {
		deep = `{"w":` + deep + `}`
	}
	if err := json.Unmarshal([]byte(`{"reasoning_content":`+deep+`}`), &v); err != nil {
		t.Fatal(err)
	}
	if string(v.R) != "" {
		t.Fatalf("over max depth: got %q want empty", v.R)
	}
	shallow := `{"text":"y"}`
	for i := 0; i < flexReasoningMaxDepth; i++ {
		shallow = `{"w":` + shallow + `}`
	}
	if err := json.Unmarshal([]byte(`{"reasoning_content":`+shallow+`}`), &v); err != nil {
		t.Fatal(err)
	}
	if string(v.R) != "y" {
		t.Fatalf("within max depth: got %q want y", v.R)
	}
}

func TestFlexOpenAIReasoningFragmentReasoningContent(t *testing.T) {
	var v struct {
		R flexOpenAIReasoningFragment `json:"reasoning_content"`
	}
	for _, tc := range []struct{ raw, want string }{
		{`{"reasoning_content":{"text":"rc"}}`, "rc"},
		{`{"reasoning_content":["x","y"]}`, "x\ny"},
		{`{"reasoning_content":{"data":{"text":"nested"}}}`, "nested"},
		{`{"reasoning_content":{"outer":{"text":"deep"}}}`, "deep"},
	} {
		if err := json.Unmarshal([]byte(tc.raw), &v); err != nil {
			t.Fatalf("%s: %v", tc.raw, err)
		}
		if string(v.R) != tc.want {
			t.Fatalf("%s: got %q want %q", tc.raw, v.R, tc.want)
		}
	}
}
