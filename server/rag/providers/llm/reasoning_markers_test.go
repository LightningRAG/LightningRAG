package llm

import "testing"

func TestStripAssistantReasoningMarkers(t *testing.T) {
	open, close := AssistantReasoningOpenTag, AssistantReasoningCloseTag
	in := "a" + open + "inner" + close + "b"
	if g := StripAssistantReasoningMarkers(in); g != "ab" {
		t.Fatalf("got %q want ab", g)
	}
	two := "x" + open + "1" + close + "y" + open + "2" + close + "z"
	if g := StripAssistantReasoningMarkers(two); g != "xyz" {
		t.Fatalf("nested pairs: got %q", g)
	}
	if g := StripAssistantReasoningMarkers("plain"); g != "plain" {
		t.Fatalf("plain: got %q", g)
	}
	unclosed := "p" + open + "no close"
	if g := StripAssistantReasoningMarkers(unclosed); g != unclosed {
		t.Fatalf("unclosed: got %q want unchanged", g)
	}
}
