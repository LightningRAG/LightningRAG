package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
)

// ProcessOpenChannelWebhookGet 公开 GET：微信 echostr / 企微 echostr / WhatsApp hub.challenge 等
func (s *ChannelConnectorService) ProcessOpenChannelWebhookGet(ctx context.Context, connectorID uint, q url.Values) (*WebhookOutcome, error) {
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ?", connectorID).First(&conn).Error; err != nil {
		return nil, err
	}
	if !conn.Enabled {
		return nil, ErrConnectorDisabled
	}
	switch conn.Channel {
	case "wechat_mp":
		return s.ProcessWeChatWebhookVerify(ctx, connectorID, q)
	case "wecom":
		return s.ProcessWeComWebhookVerify(ctx, connectorID, q)
	case "whatsapp":
		return s.ProcessWhatsAppWebhookVerify(ctx, &conn, q)
	default:
		return nil, ErrChannelGETVerifyNotSupported
	}
}

// ProcessWhatsAppWebhookVerify Meta Cloud API 订阅 URL 校验（hub.verify_token）
func (s *ChannelConnectorService) ProcessWhatsAppWebhookVerify(ctx context.Context, conn *rag.RagChannelConnector, q url.Values) (*WebhookOutcome, error) {
	if conn.Channel != "whatsapp" {
		return nil, ErrChannelGETVerifyNotSupported
	}
	extra := map[string]any{}
	if strings.TrimSpace(conn.Extra) != "" {
		_ = json.Unmarshal([]byte(conn.Extra), &extra)
	}
	mode := strings.TrimSpace(q.Get("hub.mode"))
	challenge := q.Get("hub.challenge")
	verifyTok := strings.TrimSpace(q.Get("hub.verify_token"))
	if mode != "subscribe" || strings.TrimSpace(challenge) == "" {
		return nil, fmt.Errorf("whatsapp: 缺少 hub.mode=subscribe 或 hub.challenge")
	}
	expect := channel.WhatsAppVerifyTokenFromExtra(extra)
	if expect == "" {
		expect = conn.WebhookSecret
	}
	if expect == "" {
		return nil, fmt.Errorf("whatsapp: 请在 extra 配置 whatsapp_verify_token 或设置连接器 WebhookSecret")
	}
	if !constantTimeSecretEqual(verifyTok, expect) {
		return nil, ErrWebhookSecretMismatch
	}
	return &WebhookOutcome{
		ImmediateBody:        []byte(challenge),
		ImmediateContentType: "text/plain; charset=utf-8",
		ImmediateStatus:      200,
	}, nil
}

func (s *ChannelConnectorService) ProcessWeChatWebhookVerify(ctx context.Context, connectorID uint, q url.Values) (*WebhookOutcome, error) {
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ?", connectorID).First(&conn).Error; err != nil {
		return nil, err
	}
	if conn.Channel != "wechat_mp" {
		return nil, ErrWeChatConnectorExpected
	}
	if !conn.Enabled {
		return nil, ErrConnectorDisabled
	}
	extra := map[string]any{}
	if strings.TrimSpace(conn.Extra) != "" {
		_ = json.Unmarshal([]byte(conn.Extra), &extra)
	}
	token := channel.WechatTokenFromExtra(extra)
	if token == "" {
		return nil, fmt.Errorf("wechat_mp 未配置 extra.wechat_token")
	}
	aesKey, aesOK, err := channel.WechatAESKeyFromExtra(extra)
	if err != nil {
		return nil, err
	}
	appID := channel.WechatAppIDFromExtra(extra)
	msgSig := q.Get("msg_signature")
	encEcho := q.Get("echostr")
	if encEcho == "" {
		return nil, fmt.Errorf("missing echostr")
	}
	if aesOK {
		if appID == "" {
			return nil, fmt.Errorf("安全模式 URL 验证需配置 wechat_app_id")
		}
		if msgSig == "" {
			return nil, fmt.Errorf("安全模式 URL 验证需参数 msg_signature")
		}
		if !channel.WechatMsgSignatureMatch(token, q.Get("timestamp"), q.Get("nonce"), encEcho, msgSig) {
			return nil, ErrWebhookSecretMismatch
		}
		dec, err := channel.WechatDecryptEncryptBase64(encEcho, aesKey)
		if err != nil {
			return nil, err
		}
		plainEcho, err := channel.WechatUnpackDecryptedPayload(dec, appID)
		if err != nil {
			return nil, err
		}
		return &WebhookOutcome{
			ImmediateBody:        plainEcho,
			ImmediateContentType: "text/plain; charset=utf-8",
			ImmediateStatus:      200,
		}, nil
	}
	if !channel.WechatVerifySignature(q, token) {
		return nil, ErrWebhookSecretMismatch
	}
	return &WebhookOutcome{
		ImmediateBody:        []byte(encEcho),
		ImmediateContentType: "text/plain; charset=utf-8",
		ImmediateStatus:      200,
	}, nil
}

func (s *ChannelConnectorService) ProcessWeComWebhookVerify(ctx context.Context, connectorID uint, q url.Values) (*WebhookOutcome, error) {
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ?", connectorID).First(&conn).Error; err != nil {
		return nil, err
	}
	if conn.Channel != "wecom" {
		return nil, ErrWeComConnectorExpected
	}
	if !conn.Enabled {
		return nil, ErrConnectorDisabled
	}
	extra := map[string]any{}
	if strings.TrimSpace(conn.Extra) != "" {
		_ = json.Unmarshal([]byte(conn.Extra), &extra)
	}
	token := channel.WecomTokenFromExtra(extra)
	if token == "" {
		return nil, fmt.Errorf("wecom 未配置 extra.wecom_token")
	}
	aesKey, aesOK, err := channel.WecomAESKeyFromExtra(extra)
	if err != nil {
		return nil, err
	}
	corpID := channel.WecomCorpIDFromExtra(extra)
	msgSig := q.Get("msg_signature")
	encEcho := q.Get("echostr")
	if encEcho == "" {
		return nil, fmt.Errorf("missing echostr")
	}
	if aesOK {
		if corpID == "" {
			return nil, fmt.Errorf("安全模式 URL 验证需配置 wecom_corp_id")
		}
		if msgSig == "" {
			return nil, fmt.Errorf("安全模式 URL 验证需参数 msg_signature")
		}
		if !channel.WechatMsgSignatureMatch(token, q.Get("timestamp"), q.Get("nonce"), encEcho, msgSig) {
			return nil, ErrWebhookSecretMismatch
		}
		dec, err := channel.WechatDecryptEncryptBase64(encEcho, aesKey)
		if err != nil {
			return nil, err
		}
		plainEcho, err := channel.WechatUnpackDecryptedPayload(dec, corpID)
		if err != nil {
			return nil, err
		}
		return &WebhookOutcome{
			ImmediateBody:        plainEcho,
			ImmediateContentType: "text/plain; charset=utf-8",
			ImmediateStatus:      200,
		}, nil
	}
	if !channel.WechatVerifySignature(q, token) {
		return nil, ErrWebhookSecretMismatch
	}
	return &WebhookOutcome{
		ImmediateBody:        []byte(encEcho),
		ImmediateContentType: "text/plain; charset=utf-8",
		ImmediateStatus:      200,
	}, nil
}
