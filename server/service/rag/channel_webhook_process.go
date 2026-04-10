package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *ChannelConnectorService) tryConsumeWebhookEvent(ctx context.Context, connectorID uint, eventKey string) bool {
	eventKey = strings.TrimSpace(eventKey)
	if eventKey == "" {
		return true
	}
	ev := rag.RagChannelWebhookEvent{ConnectorID: connectorID, EventKey: eventKey}
	if err := global.LRAG_DB.WithContext(ctx).Create(&ev).Error; err != nil {
		global.LRAG_LOG.Debug("webhook event dedup/skip", zap.Uint("connector", connectorID), zap.String("key", eventKey), zap.Error(err))
		return false
	}
	return true
}

// ProcessWebhook 公开 POST：鉴权、ParseWebhook、握手/延迟响应、多消息 Agent、微信系被动回复等
func (s *ChannelConnectorService) ProcessWebhook(ctx context.Context, connectorID uint, secretHeader string, rawBody []byte, meta channel.WebhookHTTPMeta) (*WebhookOutcome, error) {
	if meta.Method == "" {
		meta.Method = http.MethodPost
	}
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ?", connectorID).First(&conn).Error; err != nil {
		return nil, err
	}
	if !conn.Enabled {
		return nil, ErrConnectorDisabled
	}

	extra := map[string]any{}
	if strings.TrimSpace(conn.Extra) != "" {
		_ = json.Unmarshal([]byte(conn.Extra), &extra)
	}

	bodyIn, err := prepareChannelWebhookBody(&conn, extra, rawBody, secretHeader, meta)
	if err != nil {
		return nil, err
	}

	cfg := &channel.ConnectorConfig{Channel: conn.Channel, Extra: extra}

	ad, err := channel.Lookup(conn.Channel)
	if err != nil {
		return nil, err
	}

	dispatch, err := ad.ParseWebhook(ctx, bodyIn, cfg)
	if err != nil {
		return nil, err
	}
	if dispatch.DingTalkCheckURL {
		tok := channel.DingTalkCallbackTokenFromExtra(extra)
		key, kOk, err := channel.DingTalkAESKeyFromExtra(extra)
		if err != nil {
			return nil, err
		}
		if !kOk || tok == "" {
			return nil, fmt.Errorf("钉钉 check_url 需在 extra 配置 dingtalk_token、dingtalk_encoding_aes_key、dingtalk_suite_key")
		}
		sk := channel.DingTalkSuiteKeyFromExtra(extra)
		encOK, err := channel.DingTalkBuildEncryptedSuccessResponse(tok, key, sk)
		if err != nil {
			return nil, err
		}
		return &WebhookOutcome{
			FinalBody:        encOK,
			FinalContentType: "application/json; charset=utf-8",
		}, nil
	}
	if len(dispatch.ImmediateJSON) > 0 {
		out := &WebhookOutcome{
			ImmediateBody:   dispatch.ImmediateJSON,
			ImmediateStatus: 200,
		}
		if dispatch.ImmediateContentType != "" {
			out.ImmediateContentType = dispatch.ImmediateContentType
		}
		return out, nil
	}

	if dispatch.DiscordDeferred != nil {
		dd := *dispatch.DiscordDeferred
		connCopy := conn
		return &WebhookOutcome{
			ImmediateBody:   []byte(`{"type":5}`),
			ImmediateStatus: 200,
			DeferredAfter: func(bg context.Context) error {
				return s.discordDeferredFollowUp(bg, &connCopy, &dd)
			},
		}, nil
	}

	if (conn.Channel == "wechat_mp" || conn.Channel == "wecom") && len(dispatch.Messages) == 0 {
		return &WebhookOutcome{
			ImmediateBody:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
			ImmediateStatus:      200,
		}, nil
	}

	type msgResult struct {
		ThreadKey      string `json:"threadKey"`
		Reply          string `json:"reply"`
		ConversationID uint   `json:"conversationId"`
		Error          string `json:"error,omitempty"`
		WorkflowPaused bool   `json:"workflowPaused,omitempty"`
		SendReplyError string `json:"sendReplyError,omitempty"`
	}

	results := make([]msgResult, 0, len(dispatch.Messages))
	globals := map[string]any{
		"channel":      conn.Channel,
		"connector_id": conn.ID,
	}

	var wechatFan, wechatMP string

	for _, msg := range dispatch.Messages {
		if (conn.Channel == "wechat_mp" || conn.Channel == "wecom") && msg.ReplyRef.Opaque != nil {
			if v, ok := msg.ReplyRef.Opaque["wechat_from_user"].(string); ok {
				wechatFan = v
			}
			if v, ok := msg.ReplyRef.Opaque["wechat_to_user"].(string); ok {
				wechatMP = v
			}
		}

		threadKey := strings.TrimSpace(msg.ThreadKey)
		text := strings.TrimSpace(msg.Text)
		if threadKey == "" || text == "" {
			continue
		}

		if !s.tryConsumeWebhookEvent(ctx, conn.ID, msg.EventID) {
			continue
		}

		var convID uint
		var sess rag.RagChannelSession
		q := global.LRAG_DB.Where("connector_id = ? AND thread_key = ?", conn.ID, threadKey)
		err := q.First(&sess).Error
		if err == nil {
			convID = sess.ConversationID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		out, newConvID, err := (&AgentService{}).Run(ctx, conn.OwnerID, conn.AuthorityID, conn.AgentID, nil, text, convID, globals)
		if err != nil {
			results = append(results, msgResult{ThreadKey: threadKey, Error: err.Error()})
			continue
		}

		reply := extractChannelAgentContent(out)
		paused := false
		if out != nil {
			if v, ok := out["workflowPausedAtEntry"].(bool); ok {
				paused = v
			}
		}

		if convID == 0 && newConvID > 0 {
			ns := rag.RagChannelSession{
				ConnectorID:    conn.ID,
				ThreadKey:      threadKey,
				ConversationID: newConvID,
			}
			if err := global.LRAG_DB.Create(&ns).Error; err != nil {
				var existing rag.RagChannelSession
				if err2 := global.LRAG_DB.Where("connector_id = ? AND thread_key = ?", conn.ID, threadKey).First(&existing).Error; err2 != nil {
					results = append(results, msgResult{ThreadKey: threadKey, Error: err.Error()})
					continue
				}
				newConvID = existing.ConversationID
			}
		}

		ref := msg.ReplyRef
		if ref.Opaque == nil {
			ref = channel.ThreadRef{}
		}
		var sendErr error
		if conn.Channel == "wecom" {
			sendErr = nil
		} else {
			sendErr = ad.SendReply(ctx, conn.WebhookSecret, cfg, ref, reply)
		}
		sendNote := ""
		if sendErr != nil {
			sendNote = sendErr.Error()
			global.LRAG_LOG.Warn("channel SendReply", zap.String("channel", conn.Channel), zap.Error(sendErr))
			if ref.Opaque != nil {
				if qErr := s.enqueueChannelOutbound(ctx, conn.ID, conn.Channel, reply, ref.Opaque); qErr != nil {
					global.LRAG_LOG.Warn("channel outbound enqueue", zap.Error(qErr))
				} else {
					sendNote = sendNote + "；已入队重试"
				}
			}
		}

		results = append(results, msgResult{
			ThreadKey:      threadKey,
			Reply:          reply,
			ConversationID: newConvID,
			WorkflowPaused: paused,
			SendReplyError: sendNote,
		})
	}

	if conn.Channel == "wechat_mp" || conn.Channel == "wecom" {
		if len(results) == 1 && results[0].Error == "" && results[0].Reply != "" && wechatFan != "" && wechatMP != "" {
			plain := channel.WechatTextReplyXML(wechatFan, wechatMP, results[0].Reply)
			var aesKey []byte
			var aesOK bool
			var tok, receiveID string
			if conn.Channel == "wechat_mp" {
				aesKey, aesOK, err = channel.WechatAESKeyFromExtra(extra)
				tok = channel.WechatTokenFromExtra(extra)
				receiveID = channel.WechatAppIDFromExtra(extra)
			} else {
				aesKey, aesOK, err = channel.WecomAESKeyFromExtra(extra)
				tok = channel.WecomTokenFromExtra(extra)
				receiveID = channel.WecomCorpIDFromExtra(extra)
			}
			if err != nil {
				return nil, err
			}
			if aesOK {
				encOut, err := channel.WechatEncryptedPassiveXML(plain, tok, aesKey, receiveID)
				if err != nil {
					return nil, err
				}
				return &WebhookOutcome{
					FinalBody:        encOut,
					FinalContentType: "application/xml; charset=utf-8",
				}, nil
			}
			return &WebhookOutcome{
				FinalBody:        plain,
				FinalContentType: "application/xml; charset=utf-8",
			}, nil
		}
		return &WebhookOutcome{
			ImmediateBody:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
			ImmediateStatus:      200,
		}, nil
	}

	return &WebhookOutcome{
		JSONResponse: map[string]any{"ok": true, "results": results},
	}, nil
}
