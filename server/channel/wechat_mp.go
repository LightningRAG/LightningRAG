package channel

import (
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/xml"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type wechatMPAdapter struct{}

func init() {
	Register("wechat_mp", wechatMPAdapter{})
}

// WechatVerifySignature 微信公众平台 URL 校验（token + timestamp + nonce 字典序后 SHA1）
func WechatVerifySignature(q url.Values, token string) bool {
	token = strings.TrimSpace(token)
	if token == "" || q == nil {
		return false
	}
	sig := q.Get("signature")
	ts := q.Get("timestamp")
	nonce := q.Get("nonce")
	if sig == "" {
		return false
	}
	arr := []string{token, ts, nonce}
	sort.Strings(arr)
	sum := sha1.Sum([]byte(arr[0] + arr[1] + arr[2]))
	hex := fmt.Sprintf("%x", sum)
	if len(hex) != len(sig) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hex), []byte(sig)) == 1
}

// WechatTextReplyXML 被动回复文本（明文模式）
func WechatTextReplyXML(toUser, fromUser, content string) []byte {
	toUser = strings.TrimSpace(toUser)
	fromUser = strings.TrimSpace(fromUser)
	content = wechatSanitizeCDATA(content)
	b := fmt.Appendf(nil, `<xml>
<ToUserName><![CDATA[%s]]></ToUserName>
<FromUserName><![CDATA[%s]]></FromUserName>
<CreateTime>%d</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[%s]]></Content>
</xml>`, toUser, fromUser, time.Now().Unix(), content)
	return b
}

func wechatSanitizeCDATA(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 2048 {
		s = s[:2048] + "…"
	}
	return strings.ReplaceAll(s, "]]>", "」」>")
}

type wechatXMLIn struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgId        int64    `xml:"MsgId"`
	Event        string   `xml:"Event"`
}

func extraString(extra map[string]any, key string) string {
	if extra == nil {
		return ""
	}
	v, ok := extra[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

func (wechatMPAdapter) ParseWebhook(_ context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error) {
	if cfg == nil {
		cfg = &ConnectorConfig{}
	}
	rawBody = trimBOMXML(rawBody)
	if len(rawBody) == 0 {
		return &WebhookDispatch{
			ImmediateJSON:        []byte("success"),
			ImmediateContentType: "text/plain; charset=utf-8",
		}, nil
	}
	var in wechatXMLIn
	if err := xml.Unmarshal(rawBody, &in); err != nil {
		return nil, fmt.Errorf("wechat_mp: xml: %w", err)
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
	threadKey := in.FromUserName + ":" + in.ToUserName
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

func trimBOMXML(b []byte) []byte {
	b = []byte(strings.TrimSpace(string(b)))
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	return b
}

func (wechatMPAdapter) SendReply(context.Context, string, *ConnectorConfig, ThreadRef, string) error {
	// 被动回复走 HTTP XML，由服务层组装 FinalBody
	return nil
}

// WechatTokenFromExtra 读取公众号接口配置里的 Token
func WechatTokenFromExtra(extra map[string]any) string {
	return extraString(extra, "wechat_token")
}
