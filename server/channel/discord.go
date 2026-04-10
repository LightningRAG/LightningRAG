package channel

import (
	"context"
	"encoding/json"
	"strings"
)

type discordAdapter struct{}

func init() {
	Register("discord", discordAdapter{})
}

type discordInteraction struct {
	Type          int             `json:"type"`
	ApplicationID string          `json:"application_id"`
	Token         string          `json:"token"`
	ChannelID     string          `json:"channel_id"`
	Data          json.RawMessage `json:"data"`
	Member        *discordMember  `json:"member"`
	User          *discordUser    `json:"user"`
	GuildID       string          `json:"guild_id"`
}

type discordMember struct {
	User discordUser `json:"user"`
}

type discordUser struct {
	ID string `json:"id"`
}

type discordAppCmdData struct {
	Name    string          `json:"name"`
	Options []discordOption `json:"options"`
}

type discordOption struct {
	Name    string          `json:"name"`
	Type    int             `json:"type"`
	Value   json.RawMessage `json:"value"`
	Options []discordOption `json:"options"`
}

// Discord 交互类型常量（部分）
const (
	DiscordInteractionPing             = 1
	DiscordInteractionApplicationCmd   = 2
	DiscordInteractionMessageComponent = 3
	discordOptionTypeString            = 3
	discordOptionTypeInteger           = 4
	discordOptionTypeBoolean           = 5
)

func (discordAdapter) ParseWebhook(_ context.Context, rawBody []byte, _ *ConnectorConfig) (*WebhookDispatch, error) {
	var it discordInteraction
	if err := json.Unmarshal(rawBody, &it); err != nil {
		return nil, err
	}
	if it.Type == DiscordInteractionPing {
		return &WebhookDispatch{ImmediateJSON: []byte(`{"type":1}`)}, nil
	}
	if it.Type != DiscordInteractionApplicationCmd {
		return &WebhookDispatch{}, nil
	}

	userID := ""
	if it.Member != nil {
		userID = it.Member.User.ID
	}
	if userID == "" && it.User != nil {
		userID = it.User.ID
	}
	if strings.TrimSpace(it.ChannelID) == "" || userID == "" {
		ephemeral, _ := json.Marshal(map[string]any{
			"type": 4,
			"data": map[string]any{
				"content": "无法识别用户或频道。",
				"flags":   64,
			},
		})
		return &WebhookDispatch{ImmediateJSON: ephemeral}, nil
	}

	q := discordExtractQuery(it.Data)
	if strings.TrimSpace(q) == "" {
		ephemeral, _ := json.Marshal(map[string]any{
			"type": 4,
			"data": map[string]any{
				"content": "请在命令中附带文本参数（例如名为 query 的字符串选项）。",
				"flags":   64,
			},
		})
		return &WebhookDispatch{ImmediateJSON: ephemeral}, nil
	}

	return &WebhookDispatch{
		DiscordDeferred: &DiscordDeferredInteraction{
			ApplicationID:    it.ApplicationID,
			InteractionToken: it.Token,
			ChannelID:        it.ChannelID,
			UserID:           userID,
			Query:            strings.TrimSpace(q),
		},
	}, nil
}

func discordExtractQuery(data json.RawMessage) string {
	var d discordAppCmdData
	if err := json.Unmarshal(data, &d); err != nil {
		return ""
	}
	if s := discordWalkOptions(d.Options); s != "" {
		return s
	}
	return ""
}

func discordWalkOptions(opts []discordOption) string {
	for _, o := range opts {
		if len(o.Options) > 0 {
			if s := discordWalkOptions(o.Options); s != "" {
				return s
			}
			continue
		}
		switch o.Type {
		case discordOptionTypeString:
			var s string
			if json.Unmarshal(o.Value, &s) == nil && strings.TrimSpace(s) != "" {
				return s
			}
		case discordOptionTypeInteger:
			var n json.Number
			if json.Unmarshal(o.Value, &n) == nil {
				return n.String()
			}
		case discordOptionTypeBoolean:
			var b bool
			if json.Unmarshal(o.Value, &b) == nil {
				if b {
					return "true"
				}
				return "false"
			}
		}
	}
	return ""
}

func (discordAdapter) SendReply(context.Context, string, *ConnectorConfig, ThreadRef, string) error {
	return nil
}
