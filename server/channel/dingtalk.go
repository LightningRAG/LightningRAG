package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type dingTalkAdapter struct{}

func init() {
	Register("dingtalk", dingTalkAdapter{})
}

type dingTalkInbound struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	ConversationID string `json:"conversationId"`
	SenderStaffID  string `json:"senderStaffId"`
	SenderID       string `json:"senderId"`
	SessionWebhook string `json:"sessionWebhook"`
	MsgID          string `json:"msgId"`
}

type dingTalkEventProbe struct {
	EventType string `json:"EventType"`
}

func (dingTalkAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	if cfg == nil {
		cfg = &ConnectorConfig{}
	}
	body, err := dingTalkDecryptInboundIfNeeded(rawBody, cfg.Extra)
	if err != nil {
		return nil, err
	}

	var probe dingTalkEventProbe
	if json.Unmarshal(body, &probe) == nil && strings.EqualFold(strings.TrimSpace(probe.EventType), "check_url") {
		return &WebhookDispatch{DingTalkCheckURL: true}, nil
	}

	var in dingTalkInbound
	if err := json.Unmarshal(body, &in); err != nil {
		return nil, fmt.Errorf("dingtalk: json: %w", err)
	}
	mt := strings.TrimSpace(strings.ToLower(in.MsgType))
	if mt != "" && mt != "text" {
		return &WebhookDispatch{}, nil
	}
	text := strings.TrimSpace(in.Text.Content)
	if text == "" {
		return &WebhookDispatch{}, nil
	}
	cid := strings.TrimSpace(in.ConversationID)
	sid := strings.TrimSpace(in.SenderStaffID)
	if sid == "" {
		sid = strings.TrimSpace(in.SenderID)
	}
	if cid == "" || sid == "" {
		return nil, errors.New("dingtalk: 需要 conversationId 与 senderStaffId/senderId")
	}
	threadKey := cid + ":" + sid
	ref := ThreadRef{}
	if u := strings.TrimSpace(in.SessionWebhook); u != "" && strings.HasPrefix(u, "https://") {
		ref = ThreadRef{Opaque: map[string]any{"session_webhook": u}}
	}
	return &WebhookDispatch{
		Messages: []NormalizedInbound{{
			ThreadKey: threadKey,
			Text:      text,
			EventID:   strings.TrimSpace(in.MsgID),
			ReplyRef:  ref,
		}},
	}, nil
}

func dingTalkDecryptInboundIfNeeded(rawBody []byte, extra map[string]any) ([]byte, error) {
	var wrapper struct {
		Encrypt string `json:"encrypt"`
	}
	if err := json.Unmarshal(rawBody, &wrapper); err != nil || strings.TrimSpace(wrapper.Encrypt) == "" {
		return rawBody, nil
	}
	key, ok, err := DingTalkAESKeyFromExtra(extra)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("dingtalk: 密文报文需在 extra 配置 dingtalk_encoding_aes_key（43 字符）")
	}
	plain, err := WechatDecryptEncryptBase64(wrapper.Encrypt, key)
	if err != nil {
		return nil, fmt.Errorf("dingtalk decrypt: %w", err)
	}
	suiteKey := extraString(extra, "dingtalk_suite_key")

	var in dingTalkInbound
	if err := json.Unmarshal(plain, &in); err == nil {
		if in.MsgType != "" || in.ConversationID != "" || strings.TrimSpace(in.Text.Content) != "" {
			return plain, nil
		}
	}
	inner, err := WechatUnpackDecryptedPayload(plain, suiteKey)
	if err != nil {
		return nil, fmt.Errorf("dingtalk unpack: %w", err)
	}
	return inner, nil
}

// DingTalkCallbackTokenFromExtra 开放平台回调 Token（用于 URL signature）
func DingTalkCallbackTokenFromExtra(extra map[string]any) string {
	return extraString(extra, "dingtalk_token")
}

func (dingTalkAdapter) SendReply(ctx context.Context, _ string, _ *ConnectorConfig, thread ThreadRef, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if thread.Opaque == nil {
		return errors.New("dingtalk: 缺少 sessionWebhook，请在回调 JSON 中携带 sessionWebhook")
	}
	u, _ := thread.Opaque["session_webhook"].(string)
	u = strings.TrimSpace(u)
	if u == "" || !strings.HasPrefix(u, "https://") {
		return errors.New("dingtalk: session_webhook 无效")
	}
	if len(text) > 20000 {
		text = text[:20000] + "…"
	}
	payload, err := json.Marshal(map[string]any{
		"msgtype": "text",
		"text":    map[string]string{"content": text},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dingtalk sessionWebhook: %s %s", resp.Status, string(raw))
	}
	return nil
}
