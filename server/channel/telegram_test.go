package channel

import (
	"context"
	"testing"
)

func TestTelegramParseWebhookEditedMessage(t *testing.T) {
	var ad telegramAdapter
	raw := []byte(`{"update_id":10,"edited_message":{"message_id":20,"from":{"id":30},"chat":{"id":40},"text":" edited "}}`)
	d, err := ad.ParseWebhook(context.Background(), raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Messages) != 1 {
		t.Fatalf("messages: %d", len(d.Messages))
	}
	if d.Messages[0].ThreadKey != "40:30" || d.Messages[0].Text != "edited" {
		t.Fatalf("got %+v", d.Messages[0])
	}
	if d.Messages[0].EventID != "10:20" {
		t.Fatalf("event id %q", d.Messages[0].EventID)
	}
}

func TestTelegramParseWebhookChannelPostAnonymous(t *testing.T) {
	var ad telegramAdapter
	raw := []byte(`{"update_id":1,"channel_post":{"message_id":2,"chat":{"id":-10099},"text":"news"}}`)
	d, err := ad.ParseWebhook(context.Background(), raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Messages) != 1 {
		t.Fatalf("messages: %d", len(d.Messages))
	}
	if d.Messages[0].ThreadKey != "-10099:channel" {
		t.Fatalf("thread %q", d.Messages[0].ThreadKey)
	}
	v, _ := d.Messages[0].ReplyRef.Opaque["telegram_chat_id"].(int64)
	if v != -10099 {
		t.Fatalf("chat id %v", d.Messages[0].ReplyRef.Opaque["telegram_chat_id"])
	}
}
