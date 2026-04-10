package channel

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type slackAdapter struct{}

func init() {
	Register("slack", slackAdapter{})
}

// SlackSigningSecretFromExtra Events API 签名校验密钥
func SlackSigningSecretFromExtra(extra map[string]any) string {
	return extraString(extra, "slack_signing_secret")
}

// SlackVerifyRequest 校验 X-Slack-Signature（v0:timestamp:body 的 HMAC-SHA256），并做简单重放窗口
func SlackVerifyRequest(h http.Header, body []byte, signingSecret string) bool {
	signingSecret = strings.TrimSpace(signingSecret)
	if signingSecret == "" || h == nil {
		return false
	}
	tsStr := h.Get("X-Slack-Request-Timestamp")
	sig := h.Get("X-Slack-Signature")
	if tsStr == "" || sig == "" {
		return false
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return false
	}
	now := time.Now().Unix()
	if ts < now-300 || ts > now+60 {
		return false
	}
	base := "v0:" + tsStr + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = mac.Write([]byte(base))
	want := "v0=" + hex.EncodeToString(mac.Sum(nil))
	// 长度不等时无法常量时间比较
	if len(want) != len(sig) {
		return false
	}
	var diff byte
	for i := range want {
		diff |= want[i] ^ sig[i]
	}
	return diff == 0
}

type slackEnvelope struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge"`
	EventID   string          `json:"event_id"`
	Event     json.RawMessage `json:"event"`
}

type slackMessageEvent struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	User     string `json:"user"`
	Channel  string `json:"channel"`
	BotID    string `json:"bot_id"`
	Subtype  string `json:"subtype"`
	ThreadTS string `json:"thread_ts"`
	TS       string `json:"ts"`
}

func (slackAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	_ = cfg
	var env slackEnvelope
	if err := json.Unmarshal(rawBody, &env); err != nil {
		return nil, fmt.Errorf("slack: json: %w", err)
	}
	if env.Type == "url_verification" && strings.TrimSpace(env.Challenge) != "" {
		out, err := json.Marshal(map[string]string{"challenge": env.Challenge})
		if err != nil {
			return nil, err
		}
		return &WebhookDispatch{ImmediateJSON: out}, nil
	}
	if env.Type != "event_callback" || len(env.Event) == 0 {
		return &WebhookDispatch{}, nil
	}
	var ev slackMessageEvent
	if err := json.Unmarshal(env.Event, &ev); err != nil {
		return &WebhookDispatch{}, nil
	}
	if ev.Type != "message" {
		return &WebhookDispatch{}, nil
	}
	st := strings.TrimSpace(ev.Subtype)
	if st != "" && st != "thread_broadcast" && st != "file_share" {
		// 跳过 message_changed、bot_message 等
		if st == "bot_message" || st == "message_changed" || st == "message_deleted" {
			return &WebhookDispatch{}, nil
		}
	}
	text := strings.TrimSpace(ev.Text)
	if text == "" || ev.Channel == "" || ev.User == "" {
		return &WebhookDispatch{}, nil
	}
	threadKey := slackThreadKey(ev.Channel, ev.User, ev.ThreadTS)
	ref := ThreadRef{Opaque: map[string]any{
		"slack_channel_id": ev.Channel,
		"slack_thread_ts":  strings.TrimSpace(ev.ThreadTS),
	}}
	return &WebhookDispatch{
		Messages: []NormalizedInbound{{
			ThreadKey: threadKey,
			Text:      text,
			EventID:   strings.TrimSpace(env.EventID),
			ReplyRef:  ref,
		}},
	}, nil
}

func slackThreadKey(channel, user, threadTS string) string {
	if strings.TrimSpace(threadTS) != "" {
		return channel + ":t:" + threadTS
	}
	return channel + ":u:" + user
}

func (slackAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if cfg == nil || cfg.Extra == nil {
		return errors.New("slack: no config")
	}
	tok := extraString(cfg.Extra, "slack_bot_token")
	if tok == "" {
		return errors.New("slack: extra 需配置 slack_bot_token")
	}
	ch, _ := thread.Opaque["slack_channel_id"].(string)
	ch = strings.TrimSpace(ch)
	if ch == "" {
		return errors.New("slack: missing channel id")
	}
	if len(text) > 4000 {
		text = text[:4000] + "…"
	}
	body := map[string]any{"channel": ch, "text": text}
	if ts, ok := thread.Opaque["slack_thread_ts"].(string); ok && strings.TrimSpace(ts) != "" {
		body["thread_ts"] = strings.TrimSpace(ts)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var api struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	_ = json.Unmarshal(raw, &api)
	if !api.OK {
		return fmt.Errorf("slack chat.postMessage: %s (%s)", api.Error, string(raw))
	}
	return nil
}
