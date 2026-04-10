package channel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

var errMockMissingFields = errors.New("mock webhook: threadKey and text (or query) required")

type mockAdapter struct{}

func init() {
	Register("mock", mockAdapter{})
}

// MockWebhookBody 文档化 Mock Webhook JSON（便于对接测试）
type MockWebhookBody struct {
	ThreadKey string `json:"threadKey"`
	Text      string `json:"text"`
	Query     string `json:"query"`
	EventID   string `json:"eventId"`
}

func (mockAdapter) ParseWebhook(_ context.Context, rawBody []byte, _ *ConnectorConfig) (*WebhookDispatch, error) {
	var body MockWebhookBody
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(body.Text)
	if text == "" {
		text = strings.TrimSpace(body.Query)
	}
	if body.ThreadKey == "" || text == "" {
		return nil, errMockMissingFields
	}
	return &WebhookDispatch{
		Messages: []NormalizedInbound{{
			ThreadKey: body.ThreadKey,
			Text:      text,
			EventID:   body.EventID,
		}},
	}, nil
}

func (mockAdapter) SendReply(context.Context, string, *ConnectorConfig, ThreadRef, string) error {
	return nil
}
