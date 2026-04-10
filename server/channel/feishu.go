package channel

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type feishuAdapter struct{}

func init() {
	Register("feishu", feishuAdapter{})
}

// 飞书事件体（解密后或明文 url_verification）
type feishuPlainBody struct {
	Challenge string          `json:"challenge"`
	Type      string          `json:"type"`
	Encrypt   string          `json:"encrypt"`
	Schema    string          `json:"schema"`
	Header    json.RawMessage `json:"header"`
	Event     json.RawMessage `json:"event"`
}

type feishuEventHeader struct {
	EventType string `json:"event_type"`
}

type feishuMessageEvent struct {
	Message struct {
		ChatID    string `json:"chat_id"`
		Content   string `json:"content"`
		MessageID string `json:"message_id"`
	} `json:"message"`
	Sender struct {
		SenderID struct {
			OpenID string `json:"open_id"`
			UserID string `json:"user_id"`
		} `json:"sender_id"`
	} `json:"sender"`
}

func (feishuAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	if cfg == nil {
		cfg = &ConnectorConfig{}
	}
	plain, err := feishuMaybeDecrypt(rawBody, cfg.Extra)
	if err != nil {
		return nil, err
	}

	var body feishuPlainBody
	if err := json.Unmarshal(plain, &body); err != nil {
		return nil, fmt.Errorf("feishu: json: %w", err)
	}

	var hdr feishuEventHeader
	if len(body.Header) > 0 {
		_ = json.Unmarshal(body.Header, &hdr)
	}

	if hdr.EventType == "im.message.receive_v1" {
		return feishuParseIMMessage(body.Event)
	}

	// URL 校验 / 订阅验证：challenge 存在且非 IM 消息事件时直接回 challenge
	if body.Type == "url_verification" && body.Challenge != "" {
		b, _ := json.Marshal(map[string]string{"challenge": body.Challenge})
		return &WebhookDispatch{ImmediateJSON: b}, nil
	}
	if body.Challenge != "" && hdr.EventType != "im.message.receive_v1" {
		b, _ := json.Marshal(map[string]string{"challenge": body.Challenge})
		return &WebhookDispatch{ImmediateJSON: b}, nil
	}

	return &WebhookDispatch{}, nil
}

func feishuParseIMMessage(event json.RawMessage) (*WebhookDispatch, error) {
	var ev feishuMessageEvent
	if err := json.Unmarshal(event, &ev); err != nil {
		return nil, fmt.Errorf("feishu: event: %w", err)
	}
	text, err := feishuExtractText(ev.Message.Content)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(text) == "" {
		return &WebhookDispatch{}, nil
	}
	sid := ev.Sender.SenderID.OpenID
	if sid == "" {
		sid = ev.Sender.SenderID.UserID
	}
	threadKey := ev.Message.ChatID + ":" + sid
	ref := ThreadRef{Opaque: map[string]any{"chat_id": ev.Message.ChatID}}
	return &WebhookDispatch{
		Messages: []NormalizedInbound{{
			ThreadKey: threadKey,
			Text:      text,
			EventID:   strings.TrimSpace(ev.Message.MessageID),
			ReplyRef:  ref,
		}},
	}, nil
}

func feishuExtractText(contentJSON string) (string, error) {
	contentJSON = strings.TrimSpace(contentJSON)
	if contentJSON == "" {
		return "", errors.New("feishu: empty message content")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(contentJSON), &m); err != nil {
		return "", fmt.Errorf("feishu: content json: %w", err)
	}
	if t, ok := m["text"].(string); ok {
		return t, nil
	}
	return fmt.Sprint(m["text"]), nil
}

func feishuMaybeDecrypt(rawBody []byte, extra map[string]any) ([]byte, error) {
	var outer feishuPlainBody
	if err := json.Unmarshal(rawBody, &outer); err != nil {
		return nil, err
	}
	if outer.Encrypt == "" {
		return rawBody, nil
	}
	keyStr, _ := extra["encrypt_key"].(string)
	if strings.TrimSpace(keyStr) == "" {
		return nil, errors.New("feishu: encrypted event requires extra.encrypt_key")
	}
	plain, err := feishuDecrypt(outer.Encrypt, keyStr)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

// feishuDecrypt 飞书开放平台事件加密（AES-256-CBC + SHA256 派生密钥）
func feishuDecrypt(b64, encryptKey string) ([]byte, error) {
	cipherBuf, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("feishu: base64: %w", err)
	}
	if len(cipherBuf) < aes.BlockSize {
		return nil, errors.New("feishu: cipher too short")
	}
	keySum := sha256.Sum256([]byte(encryptKey))
	block, err := aes.NewCipher(keySum[:])
	if err != nil {
		return nil, err
	}
	iv := cipherBuf[:aes.BlockSize]
	data := cipherBuf[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("feishu: invalid cipher length")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)
	data = feishuPKCS7Unpad(data, aes.BlockSize)
	return data, nil
}

func feishuPKCS7Unpad(data []byte, blockSize int) []byte {
	if len(data) == 0 {
		return data
	}
	pad := int(data[len(data)-1])
	if pad <= 0 || pad > blockSize || pad > len(data) {
		return data
	}
	return data[:len(data)-pad]
}
