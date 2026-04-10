// 第三方渠道 Webhook：GET 见 channel_webhook_get.go；POST 校验与解密见 channel_webhook_prepare.go；
// 主流程见 channel_webhook_process.go；Discord 异步见 channel_webhook_discord.go；连接器 CRUD 见 channel_connector_crud.go。
// 各渠道协议适配器在 server/channel（Adapter + Register）。
package rag

import (
	"context"
	"errors"
)

// WebhookOutcome 公开 Webhook 一次请求的 HTTP 写出策略（由 API 层消费）。
type WebhookOutcome struct {
	ImmediateBody        []byte // 非空则原样作为 HTTP body（飞书 challenge、Discord ACK 等）
	ImmediateContentType string // 空则默认 application/json; charset=utf-8
	ImmediateStatus      int    // 默认 200
	JSONResponse         any    // 与 ImmediateBody 二选一：返回 JSON 时
	DeferredAfter        func(context.Context) error
	FinalBody            []byte // 同步处理完成后回写（如微信公众号/企微 XML）
	FinalContentType     string
}

// ChannelConnectorService 第三方渠道连接器：CRUD 见 channel_connector_crud.go，Webhook 见 channel_webhook_*.go。
type ChannelConnectorService struct{}

// ErrWebhookSecretMismatch Webhook 请求头密钥与库中不一致
var ErrWebhookSecretMismatch = errors.New("webhook secret mismatch")

// ErrConnectorDisabled 连接器已禁用
var ErrConnectorDisabled = errors.New("connector disabled")

// ErrWeChatConnectorExpected 当前连接器不是 wechat_mp，不能使用微信 URL 校验
var ErrWeChatConnectorExpected = errors.New("connector is not wechat_mp")

// ErrWeComConnectorExpected 当前连接器不是 wecom，不能使用企业微信 URL 校验
var ErrWeComConnectorExpected = errors.New("connector is not wecom")

// ErrChannelGETVerifyNotSupported 当前渠道不支持在公开 Webhook URL 上使用 GET 校验（仅部分平台需要）
var ErrChannelGETVerifyNotSupported = errors.New("channel does not support GET webhook verification")
