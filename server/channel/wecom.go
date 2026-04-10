package channel

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

type wecomAdapter struct{}

func init() {
	Register("wecom", wecomAdapter{})
}

// WecomTokenFromExtra 接收消息回调 Token（与公众号一致）
func WecomTokenFromExtra(extra map[string]any) string {
	return extraString(extra, "wecom_token")
}

// WecomCorpIDFromExtra 企业 ID；AES 包尾校验与加密回复与公众号 AppID 位置相同
func WecomCorpIDFromExtra(extra map[string]any) string {
	return extraString(extra, "wecom_corp_id")
}

// WecomAESKeyFromExtra wecom_encoding_aes_key（43 字符）
func WecomAESKeyFromExtra(extra map[string]any) (key []byte, ok bool, err error) {
	return EncodingAESKeyFromExtra(extra, "wecom_encoding_aes_key")
}

// WecomCorpSecretFromExtra 应用 Secret，用于 gettoken / 主动发消息
func WecomCorpSecretFromExtra(extra map[string]any) string {
	return extraString(extra, "wecom_corp_secret")
}

// WecomAgentIDFromExtra 自建应用 AgentId（整数），message/send 必填
func WecomAgentIDFromExtra(extra map[string]any) int {
	if extra == nil {
		return 0
	}
	switch v := extra["wecom_agent_id"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

type wecomXMLIn struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgId        int64    `xml:"MsgId"`
	Event        string   `xml:"Event"`
	AgentID      int      `xml:"AgentId"`
}

func (wecomAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	_ = cfg
	rawBody = trimBOMXML(rawBody)
	if len(rawBody) == 0 {
		return &WebhookDispatch{
			ImmediateJSON:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
		}, nil
	}
	var in wecomXMLIn
	if err := xml.Unmarshal(rawBody, &in); err != nil {
		return nil, fmt.Errorf("wecom: xml: %w", err)
	}
	mt := strings.TrimSpace(strings.ToLower(in.MsgType))
	if mt == "event" {
		return &WebhookDispatch{
			ImmediateJSON:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
		}, nil
	}
	if mt != "text" {
		return &WebhookDispatch{
			ImmediateJSON:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
		}, nil
	}
	text := strings.TrimSpace(in.Content)
	if text == "" {
		return &WebhookDispatch{
			ImmediateJSON:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
		}, nil
	}
	agentPart := ""
	if in.AgentID > 0 {
		agentPart = fmt.Sprintf("%d:", in.AgentID)
	}
	threadKey := "wecom:" + agentPart + in.FromUserName + ":" + in.ToUserName
	// 与公众号被动回复共用字段名，便于 channel_connector 组装 FinalBody
	ref := ThreadRef{Opaque: map[string]any{
		"wechat_from_user": in.FromUserName,
		"wechat_to_user":   in.ToUserName,
	}}
	eventID := fmt.Sprintf("%d", in.MsgId)
	if in.MsgId == 0 {
		eventID = ""
	}
	return &WebhookDispatch{
		Messages: []NormalizedInbound{{
			ThreadKey: threadKey,
			Text:      text,
			EventID:   eventID,
			ReplyRef:  ref,
		}},
	}, nil
}

func (wecomAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	if cfg == nil || cfg.Extra == nil {
		return fmt.Errorf("wecom SendReply: no config")
	}
	corpID := WecomCorpIDFromExtra(cfg.Extra)
	sec := WecomCorpSecretFromExtra(cfg.Extra)
	agentID := WecomAgentIDFromExtra(cfg.Extra)
	if corpID == "" || sec == "" || agentID == 0 {
		return fmt.Errorf("wecom SendReply: extra 需 wecom_corp_id、wecom_corp_secret、wecom_agent_id")
	}
	if thread.Opaque == nil {
		return fmt.Errorf("wecom SendReply: missing ReplyRef")
	}
	to, _ := thread.Opaque["wechat_from_user"].(string)
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("wecom SendReply: missing touser")
	}
	return wecomSendText(ctx, corpID, sec, agentID, to, text)
}
