package channel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type lineAdapter struct{}

func init() {
	Register("line", lineAdapter{})
}

// LineChannelSecretFromExtra 用于校验 Webhook X-Line-Signature
func LineChannelSecretFromExtra(extra map[string]any) string {
	return extraString(extra, "line_channel_secret")
}

// LineChannelAccessTokenFromExtra Messaging API channel access token（发 Push）
func LineChannelAccessTokenFromExtra(extra map[string]any) string {
	return extraString(extra, "line_channel_access_token")
}

// LineVerifySignature 校验 X-Line-Signature（Base64(HMAC-SHA256(body, channel_secret))）
func LineVerifySignature(h http.Header, body []byte, channelSecret string) bool {
	channelSecret = strings.TrimSpace(channelSecret)
	if channelSecret == "" || h == nil {
		return false
	}
	sigB64 := strings.TrimSpace(h.Get("X-Line-Signature"))
	got, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil || len(got) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(channelSecret))
	_, _ = mac.Write(body)
	want := mac.Sum(nil)
	return subtle.ConstantTimeCompare(got, want) == 1
}

type lineSource struct {
	Type    string `json:"type"`
	UserID  string `json:"userId"`
	GroupID string `json:"groupId"`
	RoomID  string `json:"roomId"`
}

type lineEvent struct {
	Type       string          `json:"type"`
	ReplyToken string          `json:"replyToken"`
	Source     json.RawMessage `json:"source"`
	Message    *struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"message"`
}

type lineWebhook struct {
	Events []lineEvent `json:"events"`
}

func lineThreadKeyAndTo(src lineSource) (threadKey, to string) {
	u := strings.TrimSpace(src.UserID)
	switch strings.TrimSpace(src.Type) {
	case "group":
		g := strings.TrimSpace(src.GroupID)
		if g == "" {
			return "", ""
		}
		if u != "" {
			return "line:g:" + g + ":u:" + u, g
		}
		return "line:g:" + g, g
	case "room":
		r := strings.TrimSpace(src.RoomID)
		if r == "" {
			return "", ""
		}
		if u != "" {
			return "line:r:" + r + ":u:" + u, r
		}
		return "line:r:" + r, r
	case "user":
		if u == "" {
			return "", ""
		}
		return "line:u:" + u, u
	default:
		return "", ""
	}
}

func (lineAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	_ = cfg
	var w lineWebhook
	if err := json.Unmarshal(rawBody, &w); err != nil {
		return nil, fmt.Errorf("line: json: %w", err)
	}
	var msgs []NormalizedInbound
	for _, ev := range w.Events {
		if strings.TrimSpace(ev.Type) != "message" || ev.Message == nil {
			continue
		}
		if strings.ToLower(strings.TrimSpace(ev.Message.Type)) != "text" {
			continue
		}
		text := strings.TrimSpace(ev.Message.Text)
		if text == "" {
			continue
		}
		var src lineSource
		if len(ev.Source) > 0 {
			_ = json.Unmarshal(ev.Source, &src)
		}
		tk, to := lineThreadKeyAndTo(src)
		if tk == "" || to == "" {
			continue
		}
		if len(text) > 4800 {
			text = text[:4794] + "…"
		}
		evID := strings.TrimSpace(ev.Message.ID)
		if evID == "" {
			evID = strings.TrimSpace(ev.ReplyToken)
		}
		opaque := map[string]any{"line_to": to}
		if rt := strings.TrimSpace(ev.ReplyToken); rt != "" {
			opaque["line_reply_token"] = rt
		}
		msgs = append(msgs, NormalizedInbound{
			ThreadKey: tk,
			Text:      text,
			EventID:   evID,
			ReplyRef:  ThreadRef{Opaque: opaque},
		})
	}
	return &WebhookDispatch{Messages: msgs}, nil
}

func (lineAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	if cfg == nil || cfg.Extra == nil {
		return errors.New("line SendReply: no config")
	}
	tok := LineChannelAccessTokenFromExtra(cfg.Extra)
	if tok == "" {
		return errors.New("line SendReply: extra 需 line_channel_access_token")
	}
	if thread.Opaque == nil {
		return errors.New("line SendReply: missing ReplyRef")
	}
	to, _ := thread.Opaque["line_to"].(string)
	to = strings.TrimSpace(to)
	rt, _ := thread.Opaque["line_reply_token"].(string)
	rt = strings.TrimSpace(rt)
	if rt != "" {
		if err := lineReplyText(ctx, tok, rt, text); err == nil {
			return nil
		} else if to == "" {
			return err
		}
		// replyToken 已用过或已过期（如异步/出站重试）时回退 Push
	}
	if to == "" {
		return errors.New("line SendReply: missing line_to（且无可用 replyToken）")
	}
	return linePushText(ctx, tok, to, text)
}
