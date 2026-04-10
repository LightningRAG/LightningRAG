package rag

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

func TestConversationHistoryItemsToMessages(t *testing.T) {
	msgs := conversationHistoryItemsToMessages([]request.ConversationHistoryItem{
		{Role: "user", Content: " hi "},
		{Role: "assistant", Content: "ok"},
		{Role: "", Content: "x"},
		{Role: "user", Content: ""},
	})
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != interfaces.MessageRoleHuman {
		t.Fatalf("first role: %s", msgs[0].Role)
	}
	if msgs[1].Role != interfaces.MessageRoleAssistant {
		t.Fatalf("second role: %s", msgs[1].Role)
	}
	var many []request.ConversationHistoryItem
	for i := 0; i < maxConversationHistoryItems+10; i++ {
		many = append(many, request.ConversationHistoryItem{Role: "user", Content: "x"})
	}
	if got := len(conversationHistoryItemsToMessages(many)); got != maxConversationHistoryItems {
		t.Fatalf("cap: want %d got %d", maxConversationHistoryItems, got)
	}
}

func TestEffectiveTruncationBudget(t *testing.T) {
	u := uint(1000)
	if g := effectiveTruncationBudget(8000, &u); g != 1000 {
		t.Fatalf("request tighter: want 1000 got %d", g)
	}
	if g := effectiveTruncationBudget(8000, nil); g != 8000 {
		t.Fatalf("no request: want 8000 got %d", g)
	}
	z := uint(0)
	if g := effectiveTruncationBudget(8000, &z); g != 8000 {
		t.Fatalf("zero request: want 8000 got %d", g)
	}
	loose := uint(100000)
	if g := effectiveTruncationBudget(8000, &loose); g != 8000 {
		t.Fatalf("request larger than model: want 8000 got %d", g)
	}
	if g := effectiveTruncationBudget(0, &u); g != 1000 {
		t.Fatalf("model unlimited: want 1000 got %d", g)
	}
}
