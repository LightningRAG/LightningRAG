package llm

import "testing"

func TestPickOpenAIReasoning(t *testing.T) {
	if g := pickOpenAIReasoning("a", "b", "c"); g != "a" {
		t.Fatalf("priority 1: got %q want a", g)
	}
	if g := pickOpenAIReasoning("", "b", "c"); g != "b" {
		t.Fatalf("priority 2: got %q want b", g)
	}
	if g := pickOpenAIReasoning("", "", "c"); g != "c" {
		t.Fatalf("priority 3: got %q want c", g)
	}
	if g := pickOpenAIReasoning("", "", "", "d"); g != "d" {
		t.Fatalf("priority 4 thinking: got %q want d", g)
	}
	if g := pickOpenAIReasoning(); g != "" {
		t.Fatalf("empty: got %q", g)
	}
}

func TestBedrockModelSupportsAnthropicThinking(t *testing.T) {
	if !bedrockModelSupportsAnthropicThinking("anthropic.claude-3-5-sonnet-20240620-v1:0") {
		t.Fatal("claude id")
	}
	if !bedrockModelSupportsAnthropicThinking("us.anthropic.claude-sonnet-4-20250514-v1:0") {
		t.Fatal("inference profile id")
	}
	if bedrockModelSupportsAnthropicThinking("meta.llama3-70b-instruct-v1:0") {
		t.Fatal("llama should be false")
	}
	if bedrockModelSupportsAnthropicThinking("") {
		t.Fatal("empty")
	}
}

func TestOllamaThinkFromReasoningEffort(t *testing.T) {
	if _, ok := ollamaThinkFromReasoningEffort(""); ok {
		t.Fatal("empty")
	}
	v, ok := ollamaThinkFromReasoningEffort("high")
	if !ok || v != "high" {
		t.Fatalf("high: %v %v", v, ok)
	}
	v, ok = ollamaThinkFromReasoningEffort("none")
	if !ok || v != false {
		t.Fatalf("none: %v %v", v, ok)
	}
}

func TestAnthropicThinkingBlockFromReasoningEffort(t *testing.T) {
	mt := 4096
	if g := anthropicThinkingBlockFromReasoningEffort("", &mt); g != nil {
		t.Fatalf("empty effort: %#v", g)
	}
	if mt != 4096 {
		t.Fatalf("empty effort changed maxTokens: %d", mt)
	}
	mt = 4096
	g := anthropicThinkingBlockFromReasoningEffort("high", &mt)
	if g == nil {
		t.Fatal("high: nil block")
	}
	if g["type"] != "enabled" {
		t.Fatalf("high type: %#v", g)
	}
	if mt < 20_000+anthropicThinkingMinAnswerRoom {
		t.Fatalf("high should raise maxTokens, got %d", mt)
	}
}

func TestCohereThinkingFromReasoningEffort(t *testing.T) {
	if g := cohereThinkingFromReasoningEffort(""); g != nil {
		t.Fatalf("empty: got %#v", g)
	}
	if g := cohereThinkingFromReasoningEffort("disabled"); g["type"] != "disabled" {
		t.Fatalf("disabled: %#v", g)
	}
	if g := cohereThinkingFromReasoningEffort("high"); g["type"] != "enabled" || g["token_budget"] != nil {
		t.Fatalf("high: %#v", g)
	}
	glow := cohereThinkingFromReasoningEffort("low")
	if glow["type"] != "enabled" {
		t.Fatalf("low: %#v", glow)
	}
	if tb, ok := glow["token_budget"].(int); !ok || tb != 2048 {
		t.Fatalf("low token_budget: %#v", glow)
	}
}

func TestCohereStripSSEDataLine(t *testing.T) {
	tests := []struct {
		line string
		want string
		ok   bool
	}{
		{`data: {"type":"content-delta"}`, `{"type":"content-delta"}`, true},
		{`{"type":"message"}`, `{"type":"message"}`, true},
		{"", "", false},
		{"  ", "", false},
		{": ping", "", false},
		{"event: message", "", false},
		{"data: [DONE]", "", false},
		{`data: data: {"k":1}`, `{"k":1}`, true},
	}
	for _, tc := range tests {
		got, ok := cohereStripSSEDataLine(tc.line)
		if ok != tc.ok || got != tc.want {
			t.Errorf("line %q: got (%q, %v) want (%q, %v)", tc.line, got, ok, tc.want, tc.ok)
		}
	}
}
