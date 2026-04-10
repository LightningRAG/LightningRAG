# Third-party chat channel connectors (Webhook)

**Language:** English | **中文：** [THIRD_PARTY_CHANNEL_CONNECTORS_zh.md](./THIRD_PARTY_CHANNEL_CONNECTORS_zh.md)

LightningRAG connects **Feishu, DingTalk, WeChat OA, WeCom, Discord, Slack, Telegram, Teams, WhatsApp, LINE** and similar platforms to bound **Agents** via **Webhooks**: public URL, authentication, `extra` JSON, admin APIs, and operations.

> **Note:** **OAuth2 quick login** (GitHub / Google / Microsoft, etc.) is **not** covered here. See [THIRD_PARTY_OAUTH_QUICK_LOGIN.md](./THIRD_PARTY_OAUTH_QUICK_LOGIN.md).

---

## 1. Architecture

| Layer | Path | Role |
|------|------|------|
| Adapters | `server/channel/` | Per-channel `Adapter`: `ParseWebhook`, `SendReply`; `Register("id", adapter)` in `init()` |
| Orchestration | `server/service/rag/channel_webhook_*.go`, `channel_connector_crud.go`, `channel_outbound.go` | Verify/decrypt POST, run Agent, WeChat/WeCom passive XML, Discord defer, outbound retry queue |
| Admin API | `server/api/v1/rag/channel_connector.go` | JWT + Casbin: connector CRUD, channel list, outbound queue |
| Public Webhook | `server/router/rag/channel_connector.go` | **No JWT**: `GET/POST .../open/channel/webhook/:connectorId` |

Outbound HTTP uses `server/channel/httpclient.go` (`ExternalHTTPDo`) with global timeout/proxy settings.

---

## 2. Public Webhook URL

- **POST** (and some **GET** challenges):

  `{System.RouterPrefix}/open/channel/webhook/{connectorId}`

  `connectorId` = `rag_channel_connectors.id`.

- **Auth**
  - Most channels: header **`X-Webhook-Secret`** equals connector **`webhook_secret`** (constant-time compare).
  - Some channels use platform signatures when keys are configured; otherwise **`X-Webhook-Secret`** remains available (dev only).

- **Rate limit** (optional): `rag.channel-webhook-ip-limit-per-minute` — per **connectorId + client IP** when Redis is up; **fail-open** if Redis fails.

---

## 3. Connector `extra` JSON

`extra` must be a **JSON object** (`{}` minimum), not an array, `null`, or a bare string.

### 3.1 Registered channel IDs

**POST** `/rag/channelConnector/channelTypes` returns registered ids (sorted), matching `server/channel` `Register`.

Built-in: `mock`, `feishu`, `dingtalk`, `wechat_mp`, `wecom`, `discord`, `slack`, `telegram`, `teams`, `whatsapp`, `line`.

### 3.2 Per-channel summary

| Channel | Main `extra` keys | POST verification |
|---------|-------------------|-------------------|
| **mock** | — | `X-Webhook-Secret` |
| **feishu** | `app_id`, `app_secret`, optional `encrypt_key`; Lark: `feishu_api_base` / `lark_api_base` | default `X-Webhook-Secret` |
| **dingtalk** | `dingtalk_token`, `dingtalk_encoding_aes_key`, `dingtalk_suite_key`, … | ciphertext + URL `signature` when token present |
| **wechat_mp** | `wechat_token`; secure mode: `wechat_encoding_aes_key`, `wechat_app_id` | AES `msg_signature` or URL `signature` |
| **wecom** | `wecom_token`; secure: `wecom_encoding_aes_key`, `wecom_corp_id`; proactive: `wecom_corp_secret`, `wecom_agent_id` | AES suffix **CorpId** |
| **slack** | `slack_signing_secret` | `X-Slack-Signature` or fallback secret |
| **telegram** | `telegram_bot_token`; optional `telegram_webhook_secret` | `X-Telegram-Bot-Api-Secret-Token` or `X-Webhook-Secret` |
| **teams** | optional `teams_microsoft_app_id` | JWT Bearer or `X-Webhook-Secret` |
| **whatsapp** | `whatsapp_phone_number_id`, `whatsapp_access_token`, optional `whatsapp_app_secret`, `whatsapp_verify_token` | `X-Hub-Signature-256` or secret |
| **line** | `line_channel_secret`, `line_channel_access_token` | `X-Line-Signature` or secret |
| **discord** | per Discord app | usually `X-Webhook-Secret` |

**Telegram** handles: `message`, `edited_message`, `channel_post`, `edited_channel_post` (text).

**LINE** prefers **Reply API** when `replyToken` exists; push/retry uses **Push**.

**WeChat / WeCom:** on success, passive **XML** in the same HTTP response when possible; WeCom avoids duplicate `message/send` in the same round.

### 3.3 Configuration examples (JSON)

Platform-specific `extra` samples, curl probes, and admin create-body examples are **the same JSON values** in any language. For the full copy-paste block list, see **section 3.3** in the Chinese edition:

**[→ THIRD_PARTY_CHANNEL_CONNECTORS_zh.md §3.3](./THIRD_PARTY_CHANNEL_CONNECTORS_zh.md#33-对接配置示例)**

Minimal smoke test (mock / secret-only):

```bash
curl -sS -X POST 'https://api.example.com/open/channel/webhook/12' \
  -H 'Content-Type: application/json' \
  -H 'X-Webhook-Secret: YOUR_WEBHOOK_SECRET' \
  -d '{"ping":true}'
```

---

## 4. GET subscription challenges

| Channel | Behavior |
|---------|----------|
| **wechat_mp** | `echostr` plain or AES |
| **wecom** | same, `ReceiveId` = **CorpId** |
| **whatsapp** | `hub.mode=subscribe`, `hub.challenge`, `hub.verify_token` |

Other channels return “not supported” for GET.

---

## 5. Idempotency & tables

- **`rag_channel_webhook_events`**: dedupe by `connector_id` + platform event key (e.g. Telegram `update_id:message_id`).
- **Retention**: `rag.channel-webhook-event-retention-days` (`0` → 7 days default, `-1` disables pruning).

---

## 6. Outbound retry queue

Failed **`SendReply`** → **`rag_channel_outbounds`**; timer or admin “run once” with backoff.

| Config | Meaning |
|--------|---------|
| `channel-outbound-poll-seconds` | Poll interval; `0` → 30s; `-1` disables timer |
| `channel-outbound-max-attempts` | Max tries; `0` → 8 |
| `channel-outbound-batch-size` | Batch size; `0` → 32 |
| `channel-outbound-claim-lease-seconds` | Lease for multi-instance; `0` → 180 |

---

## 7. Admin APIs (JWT)

Under **`/rag/channelConnector/...`** (mostly **POST**): `create`, `update`, `list`, `get`, `delete`, `channelTypes`, `outbound/list`, `outbound/delete`, `outbound/runOnce`.

Update **`initialize/rag_casbin.go`** and **`initialize/rag_api.go`** when adding routes.

---

## 8. Frontend

- `web/src/view/rag/channelConnector/channelConnector.vue`
- Channel dropdown from **`channelTypes`**, with local fallback.

---

## 9. Security

1. **HTTPS** in production; prefer platform **signature / AES / JWT** over secret header alone.
2. **`webhook_secret`** and tokens stay server-side; generated secrets shown **once** on create.
3. Enable **`channel-webhook-ip-limit-per-minute`** when Redis is available.

---

## 10. Code index

| Area | Files |
|------|-------|
| Registry | `server/channel/registry.go`, `types.go` |
| Adapters | `server/channel/*.go` |
| Types/errors | `server/service/rag/channel_webhook_types.go` |
| GET | `server/service/rag/channel_webhook_get.go` |
| POST verify | `server/service/rag/channel_webhook_prepare.go` |
| Main flow | `server/service/rag/channel_webhook_process.go` |
| Discord | `server/service/rag/channel_webhook_discord.go` |
| CRUD | `server/service/rag/channel_connector_crud.go` |
| Outbound | `server/service/rag/channel_outbound.go` |
| Tasks | `server/task/channel_outbound.go`, `channel_webhook_events_prune.go` |

New channel: implement `Adapter`, `Register`, extend **`channel_webhook_prepare.go`** and **`ProcessOpenChannelWebhookGet`** if needed.

---

*If this document and code diverge, the repository source wins.*
