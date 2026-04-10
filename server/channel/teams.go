package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type teamsAdapter struct{}

func init() {
	Register("teams", teamsAdapter{})
}

// TeamsMicrosoftAppIDFromExtra Bot 的 Microsoft App ID（用于校验入站 JWT 的 aud）
func TeamsMicrosoftAppIDFromExtra(extra map[string]any) string {
	s := extraString(extra, "teams_microsoft_app_id")
	if s != "" {
		return s
	}
	return extraString(extra, "microsoft_app_id")
}

// TeamsMicrosoftAppPasswordFromExtra 用于 OAuth client_credentials
func TeamsMicrosoftAppPasswordFromExtra(extra map[string]any) string {
	s := extraString(extra, "teams_microsoft_app_password")
	if s != "" {
		return s
	}
	return extraString(extra, "microsoft_app_password")
}

type bfChannelAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type bfConversation struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
}

type bfActivity struct {
	Type         string            `json:"type"`
	ID           string            `json:"id"`
	ServiceURL   string            `json:"serviceUrl"`
	ChannelID    string            `json:"channelId"`
	Text         string            `json:"text"`
	From         *bfChannelAccount `json:"from"`
	Recipient    *bfChannelAccount `json:"recipient"`
	Conversation *bfConversation   `json:"conversation"`
}

func (teamsAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	_ = cfg
	var act bfActivity
	if err := json.Unmarshal(rawBody, &act); err != nil {
		return nil, fmt.Errorf("teams: json: %w", err)
	}
	t := strings.TrimSpace(act.Type)
	if t == "" {
		return nil, errors.New("teams: missing activity type")
	}
	if t != "message" {
		return &WebhookDispatch{}, nil
	}
	text := strings.TrimSpace(act.Text)
	if text == "" {
		return &WebhookDispatch{}, nil
	}
	if act.From == nil || act.Recipient == nil || act.Conversation == nil {
		return &WebhookDispatch{}, nil
	}
	fromID := strings.TrimSpace(act.From.ID)
	recID := strings.TrimSpace(act.Recipient.ID)
	convID := strings.TrimSpace(act.Conversation.ID)
	if convID == "" {
		convID = strings.TrimSpace(act.Conversation.ConversationID)
	}
	svc := strings.TrimSpace(act.ServiceURL)
	if fromID == "" || recID == "" || convID == "" || svc == "" {
		return &WebhookDispatch{}, nil
	}
	if fromID == recID {
		return &WebhookDispatch{}, nil
	}

	threadKey := convID + ":" + fromID
	eventID := strings.TrimSpace(act.ID)

	msg := NormalizedInbound{
		ThreadKey: threadKey,
		Text:      text,
		EventID:   eventID,
		ReplyRef: ThreadRef{Opaque: map[string]any{
			"teams_service_url":     svc,
			"teams_conversation_id": convID,
			"teams_from_bot_id":     recID,
			"teams_channel_id":      strings.TrimSpace(act.ChannelID),
		}},
	}
	return &WebhookDispatch{Messages: []NormalizedInbound{msg}}, nil
}

func (teamsAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	if cfg == nil || cfg.Extra == nil {
		return errors.New("teams SendReply: no config")
	}
	appID := TeamsMicrosoftAppIDFromExtra(cfg.Extra)
	secret := TeamsMicrosoftAppPasswordFromExtra(cfg.Extra)
	if appID == "" || secret == "" {
		return errors.New("teams SendReply: extra 需 teams_microsoft_app_id 与 teams_microsoft_app_password")
	}
	op := thread.Opaque
	if op == nil {
		return errors.New("teams SendReply: missing ReplyRef")
	}
	svc, _ := op["teams_service_url"].(string)
	conv, _ := op["teams_conversation_id"].(string)
	botID, _ := op["teams_from_bot_id"].(string)
	tok, err := teamsGetAccessToken(ctx, appID, secret)
	if err != nil {
		return err
	}
	return teamsSendMessage(ctx, tok, svc, conv, botID, text)
}
