# 第三方平台对话接入说明

本文档说明 LightningRAG 如何将 **飞书、钉钉、微信、企业微信、Discord、Slack、Telegram、Teams、WhatsApp、LINE** 等第三方对话平台通过 **Webhook** 接入已绑定的 **Agent**，包括公开 URL、鉴权方式、`extra` 配置、管理端接口与运维相关配置。

> **说明**：用户 **OAuth2 快捷登录**（GitHub / Google / Microsoft 等）与本文的「对话渠道」无关，请参阅独立文档 [THIRD_PARTY_OAUTH_QUICK_LOGIN.md](./THIRD_PARTY_OAUTH_QUICK_LOGIN_zh.md)。

---

## 1. 架构概览

| 层次 | 路径 | 职责 |
|------|------|------|
| 协议适配 | `server/channel/` | 各平台 `Adapter`：`ParseWebhook`（解析入站）、`SendReply`（主动回消息）；`init()` 中 `Register("渠道标识", adapter)` |
| 业务编排 | `server/service/rag/channel_webhook_*.go`、`channel_connector_crud.go`、`channel_outbound.go` | POST 验签/解密、调用 `ParseWebhook`、多消息跑 Agent、微信/企微被动回复 XML、Discord defer、出站重试队列 |
| 管理 API | `server/api/v1/rag/channel_connector.go` | JWT + Casbin 下的连接器 CRUD、已注册渠道列表、出站队列 |
| 公开 Webhook | `server/router/rag/channel_connector.go` | **无 JWT**：`GET/POST .../open/channel/webhook/:connectorId` |

出站 HTTP 一般经 `server/channel/httpclient.go` 的 `ExternalHTTPDo`，与全局超时/代理策略一致。

---

## 2. 公开 Webhook URL

- **POST**（及少数渠道的 **GET** 校验）地址形如：

  `{System.RouterPrefix}/open/channel/webhook/{connectorId}`

  其中 `connectorId` 为数据库中 `rag_channel_connectors.id`（创建连接器后由列表/详情可见）。

- **鉴权**
  - 多数渠道：请求头 **`X-Webhook-Secret`** 与连接器上的 **`webhook_secret`** 一致（常量时间比较）。
  - 部分渠道在配置了平台侧密钥后，改为校验平台签名（见下文各渠道表）；未配置时仍可退回 **`X-Webhook-Secret`**（仅建议开发环境）。

- **限流**（可选）：配置项 `rag.channel-webhook-ip-limit-per-minute`（>0 且 Redis 可用时，按 **connectorId + 客户端 IP** 滑动窗口限流；Redis 异常时 **fail-open**）。

---

## 3. 连接器与 `extra`

- 管理端创建/编辑连接器时，**扩展配置 `extra` 必须为合法 JSON 对象**（如 `{}`），不能为数组、`null` 或顶层字符串。
- 各平台专有字段均放在该 JSON 内；下列为常用键名（与代码中 `extraString` / 各 `*FromExtra` 一致）。

### 3.1 已注册渠道一览

可通过 **POST** `/rag/channelConnector/channelTypes`（需登录与权限）获取当前进程已注册的渠道标识列表（与 `server/channel` 内 `Register` 一致，字典序）。

当前内置标识包括：`mock`、`feishu`、`dingtalk`、`wechat_mp`、`wecom`、`discord`、`slack`、`telegram`、`teams`、`whatsapp`、`line`。

### 3.2 分渠道配置要点

| 渠道 | 主要 `extra` 键 | POST 鉴权说明 |
|------|------------------|----------------|
| **mock** | 无特殊要求 | `X-Webhook-Secret` |
| **feishu** | `app_id`、`app_secret`、加密事件需 `encrypt_key`；国际 **Lark** 可加 `feishu_api_base` 或 `lark_api_base`（如 `https://open.larksuite.com`） | 默认 `X-Webhook-Secret` |
| **dingtalk** | `dingtalk_token`、`dingtalk_encoding_aes_key`、`dingtalk_suite_key` 等（见开放平台 HTTP 回调） | 密文体会校验 URL `signature`（在具备 token 时） |
| **wechat_mp** | `wechat_token`；安全模式另需 `wechat_encoding_aes_key`（43 字符）、`wechat_app_id` | AES 时校验 `msg_signature` 并解密 `Encrypt`；否则 URL `signature` |
| **wecom** | `wecom_token`；安全模式另需 `wecom_encoding_aes_key`、`wecom_corp_id`；主动发消息另需 `wecom_corp_secret`、`wecom_agent_id`（数字或字符串） | 同微信算法，AES 尾缀为 **CorpId** |
| **slack** | `slack_signing_secret` | `X-Slack-Signature` 等；未配置则 `X-Webhook-Secret` |
| **telegram** | `telegram_bot_token`；可选 `telegram_webhook_secret`（与 `setWebhook` 的 `secret_token` 一致） | 头 **`X-Telegram-Bot-Api-Secret-Token`** 或 `X-Webhook-Secret` |
| **teams** | 可选 `teams_microsoft_app_id`（JWT 校验 **Authorization: Bearer**） | 配置 AppId 时验 JWT；否则 `X-Webhook-Secret` |
| **whatsapp** | `whatsapp_phone_number_id`、`whatsapp_access_token` 等；签名校验可选 `whatsapp_app_secret`（`X-Hub-Signature-256`） | 有 `app_secret` 时验 Meta 签名；否则 `X-Webhook-Secret` |
| **line** | `line_channel_secret`（`X-Line-Signature`）；发消息需 `line_channel_access_token` | 有 channel secret 时验签；否则 `X-Webhook-Secret` |
| **discord** | 按机器人/Webhook 文档配置；斜杠延迟响应走专用逻辑 | 通常 `X-Webhook-Secret`（或平台要求） |

**Telegram** 支持：`message`、`edited_message`、`channel_post`、`edited_channel_post`（文本）。

**LINE** 同步回复优先 **Reply API**（`replyToken`），异步/出站重试等走 **Push**。

**微信 / 企微**：单条文本成功时优先在 **同一次 HTTP 响应** 中返回 **被动回复 XML**（可加密）；企微在本轮请求内 **不** 再调用 `message/send`，避免与被动回复重复。

### 3.3 对接配置示例

以下示例中 **`connectorId`** 请换成管理端列表中的连接器 ID；**`webhook_secret`** 为创建连接器时设置或由服务端生成（与请求头 `X-Webhook-Secret` 一致）。**`extra`** 填入管理端「扩展配置」文本框（**合法 JSON 对象**）。密钥类字段请替换为你方真实值，勿提交到版本库。

#### Webhook 地址与通用请求头

假设网关对外地址为 `https://api.example.com`，路由前缀为空，连接器 ID 为 `12`：

```text
Webhook URL（POST/GET）: https://api.example.com/open/channel/webhook/12
```

多数渠道在 **未启用平台签名** 时，POST 需带：

```http
X-Webhook-Secret: <与连接器 webhook_secret 相同>
Content-Type: application/json   # 或平台要求的类型，如微信为 XML
```

用 curl 探测（**mock** 或已配置密钥的渠道）：

```bash
curl -sS -X POST 'https://api.example.com/open/channel/webhook/12' \
  -H 'Content-Type: application/json' \
  -H 'X-Webhook-Secret: YOUR_WEBHOOK_SECRET' \
  -d '{"ping":true}'
```

#### 飞书 `feishu`（国内）

```json
{
  "app_id": "cli_xxxxxxxx",
  "app_secret": "YOUR_APP_SECRET",
  "encrypt_key": "YOUR_EVENT_ENCRYPT_KEY"
}
```

- 事件订阅若开启 **Encrypt Key**，`encrypt_key` 必填；否则可先省略，仅用 `X-Webhook-Secret` 做开发联调（生产建议走加密）。

#### 飞书国际版 Lark（同一渠道 `feishu`，改 API 根域名）

```json
{
  "app_id": "cli_xxxxxxxx",
  "app_secret": "YOUR_APP_SECRET",
  "encrypt_key": "YOUR_EVENT_ENCRYPT_KEY",
  "lark_api_base": "https://open.larksuite.com"
}
```

也可使用键名 `feishu_api_base`，效果相同。

#### 钉钉 `dingtalk`（开放平台 HTTP 回调）

```json
{
  "dingtalk_token": "YOUR_CALLBACK_TOKEN",
  "dingtalk_encoding_aes_key": "YOUR_43_CHAR_AES_KEY",
  "dingtalk_suite_key": "YOUR_SUITE_OR_APP_KEY"
}
```

具体键名以钉钉开放平台当前应用「事件订阅」文档为准；密文体会带 `encrypt` 字段时，需正确配置上述项以便验签与解密。

#### 微信公众号 `wechat_mp`（明文模式示例）

```json
{
  "wechat_token": "YOUR_WECHAT_TOKEN"
}
```

#### 微信公众号（安全模式：消息加解密）

```json
{
  "wechat_token": "YOUR_WECHAT_TOKEN",
  "wechat_encoding_aes_key": "YOUR_43_CHARACTER_ENCODING_AES_KEY",
  "wechat_app_id": "wxXXXXXXXXXXXXXXXX"
}
```

`wechat_encoding_aes_key` 为公众平台后台提供的 **43 位** EncodingAESKey（按平台说明填写，勿随意截断）。

#### 企业微信 `wecom`（自建应用：接收消息 + 被动回复 + 可选主动发消息）

```json
{
  "wecom_token": "YOUR_CALLBACK_TOKEN",
  "wecom_encoding_aes_key": "YOUR_43_CHARACTER_ENCODING_AES_KEY",
  "wecom_corp_id": "wwXXXXXXXXXXXXXXXX",
  "wecom_corp_secret": "YOUR_APP_SECRET",
  "wecom_agent_id": 1000002
}
```

- 仅依赖 **被动回复**、不需要队列走 `message/send` 时，可暂时不配 `wecom_corp_secret` / `wecom_agent_id`（仍建议生产配齐以便出站重试）。
- `wecom_agent_id` 可写 **数字或字符串**（如 `"1000002"`）。

#### Slack `slack`（Events API）

```json
{
  "slack_signing_secret": "YOUR_SIGNING_SECRET"
}
```

未配置 `slack_signing_secret` 时，POST 可退回仅校验 `X-Webhook-Secret`（**不推荐生产**）。

#### Telegram `telegram`

```json
{
  "telegram_bot_token": "123456789:AAFxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "telegram_webhook_secret": "YOUR_SETWEBHOOK_SECRET_TOKEN"
}
```

- `setWebhook` 时若设置了 `secret_token`，请把同一值写入 `telegram_webhook_secret`；请求头将校验 **`X-Telegram-Bot-Api-Secret-Token`**。
- 不配 `telegram_webhook_secret` 时，使用 `X-Webhook-Secret` 与连接器密钥一致即可（开发用）。

#### Microsoft Teams `teams`

仅 **`X-Webhook-Secret`**（与多数示例相同）：

```json
{}
```

若入站需校验 **JWT**（按 Bot Framework 要求配置）：

```json
{
  "teams_microsoft_app_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

发消息等出站能力另需 `teams_microsoft_app_password`（或别名 `microsoft_app_password`），见 `server/channel/teams_outbound.go` 注释。

#### WhatsApp Cloud API `whatsapp`

```json
{
  "whatsapp_phone_number_id": "123456789012345",
  "whatsapp_access_token": "EAAxxxxxxxx...",
  "whatsapp_app_secret": "YOUR_META_APP_SECRET",
  "whatsapp_verify_token": "YOUR_HUB_VERIFY_TOKEN"
}
```

- **GET 订阅**：`hub.verify_token` 与 `whatsapp_verify_token` 或连接器 **WebhookSecret** 一致。
- **POST**：配置 `whatsapp_app_secret` 时校验 **`X-Hub-Signature-256`**；否则退回 `X-Webhook-Secret`。

#### LINE `line`

```json
{
  "line_channel_secret": "YOUR_CHANNEL_SECRET",
  "line_channel_access_token": "YOUR_CHANNEL_ACCESS_TOKEN"
}
```

- 配置 `line_channel_secret` 时校验 **`X-Line-Signature`**；否则可退回 `X-Webhook-Secret`（开发用）。

#### Discord / mock

- **discord**：按 Discord 开发者门户配置 Webhook 或 Bot；请求头通常带 **`X-Webhook-Secret`** 与连接器一致（与具体接入方式有关）。
- **mock**：`extra` 可为 `{}`，用于联调 Agent 与路由。

#### 管理端创建连接器（JSON 请求体示例）

创建时除 `extra` 外还需 **name、channel、agentId** 等，示例：

```json
{
  "name": "飞书客服机器人",
  "channel": "feishu",
  "agentId": 1,
  "enabled": true,
  "extra": "{\"app_id\":\"cli_xxx\",\"app_secret\":\"***\",\"encrypt_key\":\"***\"}"
}
```

注意：若前端/API 直接传对象而非字符串，以实际 API 定义为准；入库的 `extra` 字段在数据库中为 **JSON 字符串**。

---

## 4. GET 校验（订阅 URL）

以下渠道在平台配置「验证 URL」时会访问 **GET** 同一 Webhook 路径：

| 渠道 | 行为 |
|------|------|
| **wechat_mp** | `echostr` 明文或 AES 解密（与公众号一致） |
| **wecom** | 同算法，`ReceiveId` 使用 **CorpId** |
| **whatsapp** | `hub.mode=subscribe`、`hub.challenge`、`hub.verify_token`（与 `whatsapp_verify_token` 或 `WebhookSecret` 一致） |

其他渠道对该 GET 返回「不支持」类错误。

---

## 5. 幂等与数据表

- **`rag_channel_webhook_events`**：按 `connector_id` + 平台事件键去重（如 Telegram 的 `update_id:message_id`）。  
- **每日清理**：`rag.channel-webhook-event-retention-days`（`0` 默认保留 7 天，`-1` 关闭清理）。

---

## 6. 出站重试队列

当某渠道 **`SendReply`** 失败时，可将回复任务写入 **`rag_channel_outbounds`**，由定时任务或管理端「立即重试一轮」执行，带指数退避与最大次数上限。

**相关配置**（`server/config.yaml` → `rag`）：

| 配置项 | 含义 |
|--------|------|
| `channel-outbound-poll-seconds` | 扫描间隔；`0` 默认 30 秒；`-1` 不注册定时任务（仍可入队 + 手动 runOnce） |
| `channel-outbound-max-attempts` | 单条最大尝试次数；`0` 默认 8 |
| `channel-outbound-batch-size` | 每轮最多处理条数；`0` 默认 32 |
| `channel-outbound-claim-lease-seconds` | 多实例下认领租约（秒）；`0` 默认 180 |

管理端列表可查看 **认领租约至**（`lease_until`）便于排障。

---

## 7. 管理端 API（需 JWT）

路径均在 **`/rag/channelConnector/...`**（方法多为 **POST**），包括但不限于：

- `create`、`update`、`list`、`get`、`delete`
- `channelTypes`：已注册渠道列表
- `outbound/list`、`outbound/delete`、`outbound/runOnce`

权限由 Casbin 与初始化规则维护；新增路由需同步 **`initialize/rag_casbin.go`** 与 **`initialize/rag_api.go`**（若使用 API 元数据同步）。

---

## 8. 前端

- 页面：`web/src/view/rag/channelConnector/channelConnector.vue`
- 新建连接器时渠道下拉默认从 **`channelTypes`** 拉取；失败时使用本地兜底列表。

---

## 9. 安全建议

1. **生产环境**务必使用 **HTTPS**，并正确配置各平台 **签名 / AES / JWT**，避免仅依赖 `X-Webhook-Secret`。
2. **`webhook_secret`** 与 `extra` 内令牌仅保存在服务端；创建时若服务端生成密钥，**仅在创建响应中明文展示一次**。
3. 合理开启 **`channel-webhook-ip-limit-per-minute`** 以降低恶意刷 Webhook 的风险（依赖 Redis）。

---

## 10. 代码索引（便于二次开发）

| 模块 | 文件 |
|------|------|
| 适配器注册与类型 | `server/channel/registry.go`、`server/channel/types.go` |
| 单渠道实现 | `server/channel/*.go`（如 `feishu.go`、`wecom.go`） |
| Webhook 类型与错误 | `server/service/rag/channel_webhook_types.go` |
| GET 校验 | `server/service/rag/channel_webhook_get.go` |
| POST 验签/解密 | `server/service/rag/channel_webhook_prepare.go` |
| 主流程 | `server/service/rag/channel_webhook_process.go` |
| Discord 异步 | `server/service/rag/channel_webhook_discord.go` |
| 连接器 CRUD | `server/service/rag/channel_connector_crud.go` |
| 出站队列 | `server/service/rag/channel_outbound.go` |
| 定时任务 | `server/task/channel_outbound.go`、`server/task/channel_webhook_events_prune.go` |

新增渠道：在 `server/channel` 实现 `Adapter` 并在 `init()` 中 **`Register`**，必要时在 **`channel_webhook_prepare.go`** 增加 `case` 分支，并在 **`ProcessOpenChannelWebhookGet`** 中处理 GET（若平台需要）。

---

*文档版本与代码同步维护；若行为与实现不一致，以当前仓库源码为准。*
