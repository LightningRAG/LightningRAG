package channel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

type whatsappAdapter struct{}

func init() {
	Register("whatsapp", whatsappAdapter{})
}

// WhatsAppAppSecretFromExtra Meta 应用密钥，用于校验 X-Hub-Signature-256
func WhatsAppAppSecretFromExtra(extra map[string]any) string {
	return extraString(extra, "whatsapp_app_secret")
}

// WhatsAppVerifyTokenFromExtra Webhook 订阅时 GET 校验的 hub.verify_token（可与连接器 WebhookSecret 二选一）
func WhatsAppVerifyTokenFromExtra(extra map[string]any) string {
	return extraString(extra, "whatsapp_verify_token")
}

// WhatsAppPhoneNumberIDFromExtra Cloud API 电话号码 ID（发消息路径）
func WhatsAppPhoneNumberIDFromExtra(extra map[string]any) string {
	return extraString(extra, "whatsapp_phone_number_id")
}

// WhatsAppAccessTokenFromExtra Graph API 长期令牌
func WhatsAppAccessTokenFromExtra(extra map[string]any) string {
	return extraString(extra, "whatsapp_access_token")
}

// WhatsAppVerifySignature256 校验 Meta Webhook POST 签名（sha256= 小写十六进制）
func WhatsAppVerifySignature256(h http.Header, body []byte, appSecret string) bool {
	appSecret = strings.TrimSpace(appSecret)
	if appSecret == "" || h == nil {
		return false
	}
	sig := strings.TrimSpace(h.Get("X-Hub-Signature-256"))
	low := strings.ToLower(sig)
	const p = "sha256="
	if len(low) < len(p) || low[:len(p)] != p {
		return false
	}
	wantHex := strings.TrimSpace(sig[len(p):])
	want, err := hex.DecodeString(wantHex)
	if err != nil || len(want) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write(body)
	got := mac.Sum(nil)
	return subtle.ConstantTimeCompare(want, got) == 1
}

type waInboundMessage struct {
	From string `json:"from"`
	ID   string `json:"id"`
	Type string `json:"type"`
	Text struct {
		Body string `json:"body"`
	} `json:"text"`
	Button struct {
		Text string `json:"text"`
	} `json:"button"`
	Interactive struct {
		Type        string `json:"type"`
		ButtonReply struct {
			Title string `json:"title"`
		} `json:"button_reply"`
		ListReply struct {
			Title string `json:"title"`
		} `json:"list_reply"`
	} `json:"interactive"`
}

type waWebhook struct {
	Object string `json:"object"`
	Entry  []struct {
		Changes []struct {
			Field string `json:"field"`
			Value struct {
				MessagingProduct string             `json:"messaging_product"`
				Messages         []waInboundMessage `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func normalizeWhatsAppPhone(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractWhatsAppText(m *waInboundMessage) string {
	if m == nil {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(m.Type)) {
	case "text":
		return strings.TrimSpace(m.Text.Body)
	case "button":
		return strings.TrimSpace(m.Button.Text)
	case "interactive":
		if t := strings.TrimSpace(m.Interactive.ButtonReply.Title); t != "" {
			return t
		}
		return strings.TrimSpace(m.Interactive.ListReply.Title)
	default:
		return ""
	}
}

func (whatsappAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	_ = cfg
	var w waWebhook
	if err := json.Unmarshal(rawBody, &w); err != nil {
		return nil, fmt.Errorf("whatsapp: json: %w", err)
	}
	if w.Object != "" && w.Object != "whatsapp_business_account" {
		return &WebhookDispatch{}, nil
	}
	var msgs []NormalizedInbound
	for _, ent := range w.Entry {
		for _, ch := range ent.Changes {
			if ch.Field != "" && ch.Field != "messages" {
				continue
			}
			for i := range ch.Value.Messages {
				m := &ch.Value.Messages[i]
				text := extractWhatsAppText(m)
				from := normalizeWhatsAppPhone(m.From)
				if text == "" || from == "" {
					continue
				}
				if len(text) > 4096 {
					text = text[:4090] + "…"
				}
				msgs = append(msgs, NormalizedInbound{
					ThreadKey: "wa:" + from,
					Text:      text,
					EventID:   strings.TrimSpace(m.ID),
					ReplyRef: ThreadRef{Opaque: map[string]any{
						"whatsapp_from": from,
					}},
				})
			}
		}
	}
	return &WebhookDispatch{Messages: msgs}, nil
}

func (whatsappAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	if cfg == nil || cfg.Extra == nil {
		return errors.New("whatsapp SendReply: no config")
	}
	phoneID := WhatsAppPhoneNumberIDFromExtra(cfg.Extra)
	tok := WhatsAppAccessTokenFromExtra(cfg.Extra)
	if phoneID == "" || tok == "" {
		return errors.New("whatsapp SendReply: extra 需 whatsapp_phone_number_id 与 whatsapp_access_token")
	}
	if thread.Opaque == nil {
		return errors.New("whatsapp SendReply: missing ReplyRef")
	}
	to, _ := thread.Opaque["whatsapp_from"].(string)
	to = normalizeWhatsAppPhone(to)
	if to == "" {
		return errors.New("whatsapp SendReply: missing whatsapp_from")
	}
	return whatsappSendText(ctx, phoneID, tok, to, text)
}
