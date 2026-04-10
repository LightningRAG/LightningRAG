// Package channel 实现各第三方对话平台的 Webhook 协议适配（Adapter 接口 + Register 注册表）。
//
// 每个渠道一个或若干源文件（如 feishu.go + feishu_send.go），在 init() 中 Register(kind, adapter)。
// HTTP 出站统一可走 httpclient.go 的 ExternalHTTPDo（超时/代理等与项目一致）。
//
// 业务编排（鉴权分支、多消息 Agent、微信被动回复、出站队列）在 service/rag 的 channel_webhook_* 与 channel_outbound.go。

package channel
