package rag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"go.uber.org/zap"
)

// prepareChannelWebhookBody 按渠道校验 POST 签名（或解密），返回供 Adapter.ParseWebhook 使用的 body。
func prepareChannelWebhookBody(conn *rag.RagChannelConnector, extra map[string]any, rawBody []byte, secretHeader string, meta channel.WebhookHTTPMeta) ([]byte, error) {
	bodyIn := rawBody
	switch conn.Channel {
	case "slack":
		ss := channel.SlackSigningSecretFromExtra(extra)
		if ss != "" {
			if meta.Headers == nil || !channel.SlackVerifyRequest(meta.Headers, rawBody, ss) {
				return nil, ErrWebhookSecretMismatch
			}
		} else {
			if !constantTimeSecretEqual(secretHeader, conn.WebhookSecret) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	case "telegram":
		tsec := channel.TelegramWebhookSecretFromExtra(extra)
		if tsec != "" {
			if meta.Headers == nil {
				return nil, ErrWebhookSecretMismatch
			}
			got := meta.Headers.Get("X-Telegram-Bot-Api-Secret-Token")
			if !constantTimeSecretEqual(got, tsec) {
				return nil, ErrWebhookSecretMismatch
			}
		} else {
			if !constantTimeSecretEqual(secretHeader, conn.WebhookSecret) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	case "line":
		lsec := channel.LineChannelSecretFromExtra(extra)
		if lsec != "" {
			if meta.Headers == nil || !channel.LineVerifySignature(meta.Headers, rawBody, lsec) {
				return nil, ErrWebhookSecretMismatch
			}
		} else {
			if !constantTimeSecretEqual(secretHeader, conn.WebhookSecret) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	case "whatsapp":
		appSec := channel.WhatsAppAppSecretFromExtra(extra)
		if appSec != "" {
			if meta.Headers == nil || !channel.WhatsAppVerifySignature256(meta.Headers, rawBody, appSec) {
				return nil, ErrWebhookSecretMismatch
			}
		} else {
			if !constantTimeSecretEqual(secretHeader, conn.WebhookSecret) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	case "teams":
		appID := channel.TeamsMicrosoftAppIDFromExtra(extra)
		if appID != "" {
			if meta.Headers == nil {
				return nil, ErrWebhookSecretMismatch
			}
			auth := strings.TrimSpace(meta.Headers.Get("Authorization"))
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return nil, ErrWebhookSecretMismatch
			}
			rawTok := strings.TrimSpace(parts[1])
			if err := channel.TeamsVerifyBearerToken(rawTok, appID); err != nil {
				global.LRAG_LOG.Debug("teams jwt verify failed", zap.Error(err))
				return nil, ErrWebhookSecretMismatch
			}
		} else {
			if !constantTimeSecretEqual(secretHeader, conn.WebhookSecret) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	case "wechat_mp":
		token := channel.WechatTokenFromExtra(extra)
		if token == "" {
			return nil, fmt.Errorf("wechat_mp 未配置 extra.wechat_token")
		}
		aesKey, aesOK, err := channel.WechatAESKeyFromExtra(extra)
		if err != nil {
			return nil, err
		}
		appID := channel.WechatAppIDFromExtra(extra)
		ts := meta.Query.Get("timestamp")
		nonce := meta.Query.Get("nonce")
		msgSig := meta.Query.Get("msg_signature")
		encType := meta.Query.Get("encrypt_type")
		if aesOK {
			if appID == "" {
				return nil, fmt.Errorf("配置 wechat_encoding_aes_key 时需同时配置 wechat_app_id")
			}
			inner, cryptoUsed, err := channel.WechatDecryptInboundXML(rawBody, token, ts, nonce, msgSig, aesKey, appID)
			if err != nil {
				return nil, err
			}
			if cryptoUsed {
				bodyIn = inner
			} else {
				if encType == "aes" {
					return nil, fmt.Errorf("wechat encrypt_type=aes 但报文无 Encrypt 节点")
				}
				if !channel.WechatVerifySignature(meta.Query, token) {
					return nil, ErrWebhookSecretMismatch
				}
			}
		} else {
			if !channel.WechatVerifySignature(meta.Query, token) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	case "wecom":
		token := channel.WecomTokenFromExtra(extra)
		if token == "" {
			return nil, fmt.Errorf("wecom 未配置 extra.wecom_token")
		}
		aesKey, aesOK, err := channel.WecomAESKeyFromExtra(extra)
		if err != nil {
			return nil, err
		}
		corpID := channel.WecomCorpIDFromExtra(extra)
		ts := meta.Query.Get("timestamp")
		nonce := meta.Query.Get("nonce")
		msgSig := meta.Query.Get("msg_signature")
		encType := meta.Query.Get("encrypt_type")
		if aesOK {
			if corpID == "" {
				return nil, fmt.Errorf("配置 wecom_encoding_aes_key 时需同时配置 wecom_corp_id")
			}
			inner, cryptoUsed, err := channel.WechatDecryptInboundXML(rawBody, token, ts, nonce, msgSig, aesKey, corpID)
			if err != nil {
				return nil, err
			}
			if cryptoUsed {
				bodyIn = inner
			} else {
				if encType == "aes" {
					return nil, fmt.Errorf("wecom encrypt_type=aes 但报文无 Encrypt 节点")
				}
				if !channel.WechatVerifySignature(meta.Query, token) {
					return nil, ErrWebhookSecretMismatch
				}
			}
		} else {
			if !channel.WechatVerifySignature(meta.Query, token) {
				return nil, ErrWebhookSecretMismatch
			}
		}
	default:
		if !constantTimeSecretEqual(secretHeader, conn.WebhookSecret) {
			return nil, ErrWebhookSecretMismatch
		}
	}

	if conn.Channel == "dingtalk" {
		var wrap struct {
			Encrypt string `json:"encrypt"`
		}
		_ = json.Unmarshal(rawBody, &wrap)
		if strings.TrimSpace(wrap.Encrypt) != "" {
			tok := channel.DingTalkCallbackTokenFromExtra(extra)
			if tok != "" && meta.Query != nil && strings.TrimSpace(meta.Query.Get("signature")) != "" {
				if !channel.DingTalkVerifyURLSignature(tok, meta.Query, wrap.Encrypt) {
					return nil, ErrWebhookSecretMismatch
				}
			}
		}
	}

	return bodyIn, nil
}
