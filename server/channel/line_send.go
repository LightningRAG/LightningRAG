package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	linePushURL  = "https://api.line.me/v2/bot/message/push"
	lineReplyURL = "https://api.line.me/v2/bot/message/reply"
)

func lineLineTextMessage(text string) []map[string]any {
	text = strings.TrimSpace(text)
	if len(text) > 5000 {
		text = text[:4994] + "…"
	}
	return []map[string]any{{"type": "text", "text": text}}
}

// lineReplyText 使用 webhook 自带的 replyToken（一次性、短时有效）；失败时由调用方决定是否改 Push。
func lineReplyText(ctx context.Context, accessToken, replyToken, text string) error {
	replyToken = strings.TrimSpace(replyToken)
	text = strings.TrimSpace(text)
	if replyToken == "" || text == "" {
		return nil
	}
	body := map[string]any{
		"replyToken": replyToken,
		"messages":   lineLineTextMessage(text),
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, lineReplyURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("line reply: %s %s", resp.Status, string(out))
	}
	return nil
}

func linePushText(ctx context.Context, accessToken, to, text string) error {
	to = strings.TrimSpace(to)
	text = strings.TrimSpace(text)
	if to == "" || text == "" {
		return nil
	}
	body := map[string]any{
		"to":       to,
		"messages": lineLineTextMessage(text),
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linePushURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("line push: %s %s", resp.Status, string(out))
	}
	return nil
}
