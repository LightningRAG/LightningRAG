# Third-party quick login (OAuth2)

**Language:** English | **中文：** [THIRD_PARTY_OAUTH_QUICK_LOGIN_zh.md](./THIRD_PARTY_OAUTH_QUICK_LOGIN_zh.md)

This document explains how LightningRAG integrates **OAuth2 authorization code** (most IdPs also use **PKCE**) with identity providers such as GitHub, Google, Microsoft, and **WeChat Open Platform web apps (QR login)** for the “third-party quick login” on the sign-in page. Configuration lives mainly in the **database** and the admin UI. It is independent of **chat channel Webhooks** (see [THIRD_PARTY_CHANNEL_CONNECTORS.md](./THIRD_PARTY_CHANNEL_CONNECTORS.md)).

---

## 1. Architecture overview

| Layer | Path | Responsibility |
|------|------|------------------|
| IdP adapters | `server/oauth/` | Each platform implements `Provider`: `Kind`, `OAuth2Config`, `FetchProfile`; `Register` in `init()`; **no DB or user logic**. Some platforms add `AuthorizeURLBuilder` / `CodeExchanger` (e.g. WeChat QR: non-`client_id` params, fixed `#wechat_redirect`, GET token exchange, **no PKCE**) |
| Orchestration | `server/service/oauthapp/` | `sys_oauth_*`, state / one-time ticket cache, user create/bind, login session |
| Models | `server/model/system/` | `SysOAuthProvider`, `SysOAuthSetting`, `SysUserOAuthBinding`, etc. (tables prefixed `sys_oauth`) |
| Admin API | `server/api/v1/oauthapp/` | Provider CRUD, global settings, public handlers |
| Admin routes | `server/router/oauthapp/` | JWT + Casbin: `/sysOAuthProvider/*`, `/sysOAuthSetting/*` |
| Public routes | `server/router/system/sys_base.go` | **No JWT**: `/base/oauth/*` (rate limits below) |
| Runtime cache | `server/global/oauth_runtime.go` | In-process: frontend return URL, encryption key material from DB |
| Crypto | `server/utils/oauth_crypt.go` | Symmetric encryption for client secrets; key from DB or JWT signing key |

Admin UI: `web/src/view/oauth/settings.vue` (menu: **System tools → Third-party quick login**).

---

## 2. Tables and permissions

- **`sys_oauth_providers`**: at most one row per `kind`; stores client ID, encrypted secret, enabled flag, `extra` JSON, etc.
- **`sys_oauth_settings`**: single row (PK `1`); global **post-login frontend return URL**, optional **encryption master key**.
- **`sys_user_oauth_bindings`**: binds users to IdP `subject`.

Casbin seeds roles such as `888` for `/sysOAuthProvider/*` and `/sysOAuthSetting/*`; public `/base/oauth/*` bypasses Casbin.

---

## 3. Redirect URIs in the IdP console

The authorization callback must hit a **publicly reachable backend API** (including `system.router-prefix`):

```text
{Public API root}{RouterPrefix}/base/oauth/callback/{kind}
```

- `{kind}` is the platform id (lowercase, e.g. `github`, `google`).
- `RouterPrefix` comes from `config.yaml` → `system.router-prefix` (empty means paths start with `/base/oauth/...`).

The admin **global settings** API returns **`callbackPathPattern`** (path only, with prefix). The login page **`GET /base/oauth/providers`** also returns `data.callbackPathPattern` for building full URLs with `VITE_BASE_API`.

**Authorization entry** (browser or frontend):

```text
{API root URL}{RouterPrefix}/base/oauth/authorize/{kind}
```

Optional query: `redirect` — in-app path after login (must start with `/` and pass server whitelist `sanitizeOAuthReturnPath`).

---

## 4. Public HTTP endpoints (no login)

Under `{RouterPrefix}/base/...`, mostly `GET`:

| Path | Description |
|------|-------------|
| `GET /base/oauth/providers` | Enabled providers + `callbackPathPattern` |
| `GET /base/oauth/authorize/:kind` | 302 to IdP |
| `GET /base/oauth/callback/:kind` | IdP callback; 302 to frontend with `oauth_ex=...` |
| `GET /base/oauth/exchange?oauth_ex=...` | Exchange one-time ticket for login JSON |

Flow notes:

1. **PKCE** by default; `state` and `code_verifier` in **BlackCache** with TTL. **WeChat Open `wechat_open`** and similar flows skip PKCE.
2. Success redirects to **`sys_oauth_settings.frontend_redirect`** with **`oauth_ex`**.
3. **`exchange`** is single-use.

IdP `error` query params redirect to the frontend with **`oauth_err`** (next section).

---

## 5. Login failure: `oauth_err` query parameter

| code | Meaning |
|------|---------|
| `denied` | User cancelled / `access_denied` |
| `idp` | Other IdP `error` |
| `authorize` | Cannot build authorize URL |
| `missing` | Missing `code` or `state` |
| `state` | Invalid/expired `state` or kind mismatch |
| `provider` | No enabled config |
| `token` | Token or profile fetch failed |
| `user` | User create/bind failed or disabled |
| `session` | Session creation failed |
| `server` | Server-side ticket error |
| `unknown` | Uncategorized |

Legacy: some links may use `oauth_err=1`.

---

## 6. Rate limiting (optional)

**`rag.oauth-ip-limit-per-minute`** in `config.yaml`:

- If `> 0` and **Redis** is up, sliding window per client IP on `/base/oauth/providers`, `authorize`, `callback`, `exchange`.
- **Fail-open** if Redis is down.

Independent of chat Webhook limit `rag.channel-webhook-ip-limit-per-minute`.

---

## 7. Registered `kind` values and `extra` JSON

Registered kinds (lowercase) include:  
`auth0`, `bitbucket`, `cognito`, `discord`, `dropbox`, `facebook`, `figma`, `gitee`, `github`, `gitlab`, `google`, `kakao`, `line`, `linkedin`, `microsoft`, `okta`, `paypal`, `slack`, `spotify`, `strava`, `twitch`, `twitter`, `wechat_open`, `yandex`, `zoom`.

Admin normalizes **`kind`** to lowercase; lookups are case-insensitive.

Common **`extra`** keys (see `server/oauth/doc.go`): Microsoft `tenant`, GitLab `gitlab_base_url`, Facebook `facebook_graph_version`, GitHub Enterprise `github_enterprise_host`, Auth0 `auth0_domain`, PayPal `paypal_sandbox`, Okta `okta_issuer` / `okta_domain` + `okta_auth_server`, Cognito `cognito_domain`.

### 7.1 WeChat QR login (`wechat_open`)

[WeChat Open Platform](https://open.weixin.qq.com/) **website app** (not Official Account / mini program). Configure **authorized callback domain** to match the callback host; path remains `…/base/oauth/callback/wechat_open`.

Binding prefers **`unionid`**, else **`openid`**. Email is usually absent; new users get nickname and avatar only.

**vs Official Account / WeCom:** chat Webhooks are documented in [THIRD_PARTY_CHANNEL_CONNECTORS.md](./THIRD_PARTY_CHANNEL_CONNECTORS.md); credentials differ from OAuth login.

---

## 8. Admin APIs (JWT + permissions)

Providers: `createOAuthProvider`, `updateOAuthProvider`, `deleteOAuthProvider`, `getOAuthProviderList`, `findOAuthProvider`, `getRegisteredOAuthKinds`, etc.  
Settings: `getOAuthSetting`, `updateOAuthSetting`.

See `server/source/system/casbin.go` and `server/source/system/api.go` for exact methods.

---

## 9. Operations

- Rotating the **encryption master key** requires **re-entering** encrypted client secrets.
- **`VITE_BASE_API`** must match the browser-reachable API root.
- If menu `component` still points to `view/superAdmin/oauth/oauthProvider.vue`, update to **`view/oauth/settings.vue`**.

---

## 10. Adding a new IdP

1. Add `server/oauth/<platform>.go`, implement `Provider`, `oauth.Register` in `init()`.
2. Read `RuntimeConfig.Extra` for tenant/host-specific behavior.
3. For non-standard OAuth2, implement `AuthorizeURLBuilder` / `CodeExchanger` (see `wechat_open.go`).
4. No schema change; new `kind` appears via `getRegisteredOAuthKinds`.

See `server/oauth/doc.go` and `server/service/oauthapp/doc.go`.
