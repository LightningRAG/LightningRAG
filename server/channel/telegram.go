package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type telegramAdapter struct{}

func init() {
	Register("telegram", telegramAdapter{})
}

// TelegramWebhookSecretFromExtra 与 setWebhook 时 secret_token 一致，对应请求头 X-Telegram-Bot-Api-Secret-Token
func TelegramWebhookSecretFromExtra(extra map[string]any) string {
	return extraString(extra, "telegram_webhook_secret")
}

type tgMsg struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	From      *struct {
		ID int64 `json:"id"`
	} `json:"from"`
	SenderChat *struct {
		ID int64 `json:"id"`
	} `json:"sender_chat"`
	Chat *struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

type tgUpdate struct {
	UpdateID          int64  `json:"update_id"`
	Message           *tgMsg `json:"message"`
	EditedMessage     *tgMsg `json:"edited_message"`
	ChannelPost       *tgMsg `json:"channel_post"`
	EditedChannelPost *tgMsg `json:"edited_channel_post"`
}

func tgPickMessage(u *tgUpdate) *tgMsg {
	if u == nil {
		return nil
	}
	switch {
	case u.Message != nil:
		return u.Message
	case u.EditedMessage != nil:
		return u.EditedMessage
	case u.ChannelPost != nil:
		return u.ChannelPost
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost
	default:
		return nil
	}
}

func tgInboundFromMessage(updateID int64, m *tgMsg) (*NormalizedInbound, bool) {
	if m == nil || m.Chat == nil {
		return nil, false
	}
	text := strings.TrimSpace(m.Text)
	if text == "" {
		return nil, false
	}
	cid := m.Chat.ID
	var threadKey string
	switch {
	case m.From != nil:
		threadKey = strconv.FormatInt(cid, 10) + ":" + strconv.FormatInt(m.From.ID, 10)
	case m.SenderChat != nil:
		threadKey = strconv.FormatInt(cid, 10) + ":sc:" + strconv.FormatInt(m.SenderChat.ID, 10)
	default:
		// 频道匿名发帖等：无 from
		threadKey = strconv.FormatInt(cid, 10) + ":channel"
	}
	ref := ThreadRef{Opaque: map[string]any{
		"telegram_chat_id": cid,
	}}
	ev := ""
	if m.MessageID != 0 {
		ev = strconv.FormatInt(updateID, 10) + ":" + strconv.FormatInt(m.MessageID, 10)
	}
	return &NormalizedInbound{
		ThreadKey: threadKey,
		Text:      text,
		EventID:   ev,
		ReplyRef:  ref,
	}, true
}

func (telegramAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	_ = cfg
	var u tgUpdate
	if err := json.Unmarshal(rawBody, &u); err != nil {
		return nil, fmt.Errorf("telegram: json: %w", err)
	}
	m := tgPickMessage(&u)
	inb, ok := tgInboundFromMessage(u.UpdateID, m)
	if !ok {
		return &WebhookDispatch{}, nil
	}
	return &WebhookDispatch{Messages: []NormalizedInbound{*inb}}, nil
}

func (telegramAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if cfg == nil || cfg.Extra == nil {
		return errors.New("telegram: no config")
	}
	tok := extraString(cfg.Extra, "telegram_bot_token")
	if tok == "" {
		return errors.New("telegram: extra 需配置 telegram_bot_token")
	}
	var chatID int64
	if thread.Opaque != nil {
		if v, ok := thread.Opaque["telegram_chat_id"].(float64); ok {
			chatID = int64(v)
		}
		if chatID == 0 {
			if v, ok := thread.Opaque["telegram_chat_id"].(int64); ok {
				chatID = v
			}
		}
	}
	if chatID == 0 {
		return errors.New("telegram: missing chat id")
	}
	if len(text) > 4096 {
		text = text[:4090] + "…"
	}
	payload, err := json.Marshal(map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return err
	}
	url := "https://api.telegram.org/bot" + tok + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.Unmarshal(raw, &out)
	if !out.OK {
		return fmt.Errorf("telegram sendMessage: %s", out.Description)
	}
	return nil
}
