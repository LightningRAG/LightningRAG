package channel

import (
	"context"
	"net/http"
	"net/url"
)

// ThreadRef 渠道内回复目标（由适配器解释，供 SendReply）
type ThreadRef struct {
	Opaque map[string]any
}

// DiscordDeferredInteraction 斜杠命令等需先 ACK（type 5）再异步编辑原消息
type DiscordDeferredInteraction struct {
	ApplicationID    string
	InteractionToken string
	ChannelID        string
	UserID           string
	Query            string
}

// NormalizedInbound 统一后的入站消息
type NormalizedInbound struct {
	ThreadKey string // 映射 rag_channel_sessions.thread_key
	Text      string
	EventID   string // 可选，用于幂等（后续扩展）
	// ReplyRef 发回渠道时使用（如飞书 chat_id）；为空则 SendReply 可为空操作
	ReplyRef ThreadRef
}

// WebhookDispatch Webhook 解析结果：可能仅为平台握手，或包含多条用户消息
type WebhookDispatch struct {
	// ImmediateJSON 非空时由 HTTP 层直接回写（飞书 challenge、Discord PING、微信 success 等）
	ImmediateJSON []byte
	// ImmediateContentType 为空时默认 application/json; charset=utf-8
	ImmediateContentType string
	// Messages 需要进入 Agent 的消息列表
	Messages []NormalizedInbound
	// DiscordDeferred 与 Messages 互斥：由服务层先回 type 5，再异步跑 Agent 并 PATCH follow-up
	DiscordDeferred *DiscordDeferredInteraction
	// DingTalkCheckURL 解密后为 EventType=check_url，需回加密 JSON success（开放平台 HTTP 回调）
	DingTalkCheckURL bool
}

// WebhookHTTPMeta 请求元数据（签名校验、GET 等）
type WebhookHTTPMeta struct {
	Method  string
	Query   url.Values
	Headers http.Header
}

// ConnectorConfig 运行期配置（由 DB 行 + 解析后的 Extra 组装）
type ConnectorConfig struct {
	Channel string
	Extra   map[string]any
}

// Adapter 单渠道协议适配（新增平台实现该接口并 Register）
type Adapter interface {
	ParseWebhook(ctx context.Context, rawBody []byte, cfg *ConnectorConfig) (*WebhookDispatch, error)
	// SendReply 将 Agent 文本结果发回渠道；Mock 等可空实现
	SendReply(ctx context.Context, secret string, cfg *ConnectorConfig, thread ThreadRef, text string) error
}
