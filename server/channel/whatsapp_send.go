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

const whatsappGraphAPIBase = "https://graph.facebook.com/v21.0"

func whatsappSendText(ctx context.Context, phoneNumberID, accessToken, toPhone, text string) error {
	phoneNumberID = strings.TrimSpace(phoneNumberID)
	accessToken = strings.TrimSpace(accessToken)
	toPhone = strings.TrimSpace(toPhone)
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if len(text) > 4096 {
		text = text[:4090] + "…"
	}
	u := whatsappGraphAPIBase + "/" + phoneNumberID + "/messages"
	body := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toPhone,
		"type":              "text",
		"text": map[string]string{
			"preview_url": "false",
			"body":        text,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp send: %s %s", resp.Status, string(out))
	}
	return nil
}
