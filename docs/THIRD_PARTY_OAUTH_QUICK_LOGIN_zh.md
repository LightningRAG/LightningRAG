# 第三方平台快捷登录（OAuth2）对接说明

本文档说明 LightningRAG 如何通过 **OAuth2 授权码**（多数 IdP 另启用 **PKCE**）对接 GitHub、Google、Microsoft、**微信开放平台网站应用（扫码）** 等身份提供商（IdP），实现登录页的「第三方快捷登录」。配置主要落在**数据库**与后台界面，与 **对话渠道 Webhook**（见 [THIRD_PARTY_CHANNEL_CONNECTORS.md](./THIRD_PARTY_CHANNEL_CONNECTORS_zh.md)）相互独立。

---

## 1. 架构概览

| 层次 | 路径 | 职责 |
|------|------|------|
| IdP 协议适配 | `server/oauth/` | 各平台实现 `Provider`：`Kind`、`OAuth2Config`、`FetchProfile`；`init()` 中 `Register`；**无数据库、无用户业务**。少数平台另实现 `AuthorizeURLBuilder` / `CodeExchanger`（如微信扫码：非 `client_id` 参数、固定 `#wechat_redirect`、GET 换票、**不使用 PKCE**） |
| 业务编排 | `server/service/oauthapp/` | 读写 `sys_oauth_*`、state / 一次性换票缓存、创建或绑定用户、完成登录会话 |
| 数据模型 | `server/model/system/` | `SysOAuthProvider`、`SysOAuthSetting`、`SysUserOAuthBinding` 等（表名以 `sys_oauth` 为前缀） |
| 管理 API | `server/api/v1/oauthapp/` | 提供商 CRUD、全局设置、公开端点的处理器 |
| 管理路由 | `server/router/oauthapp/` | 需 JWT + Casbin：`/sysOAuthProvider/*`、`/sysOAuthSetting/*` |
| 公开路由 | `server/router/system/sys_base.go` | **无 JWT**：`/base/oauth/*`（含限流中间件，见下文） |
| 运行时缓存 | `server/global/oauth_runtime.go` | 进程内缓存：前端回跳地址、加密主密钥材料（来自库表） |
| 密钥加解密 | `server/utils/oauth_crypt.go` | Client Secret 等对称加密；材料优先 DB，否则回退 JWT 签名密钥 |

前端管理页：`web/src/view/oauth/settings.vue`（菜单一般在 **系统工具 → 第三方快捷登录**）。

---

## 2. 数据表与权限

- **`sys_oauth_providers`**：每种 `kind` 至多一条；存 Client ID、密文密钥、是否启用、`extra` JSON 等。
- **`sys_oauth_settings`**：单行（主键固定为 1）；全局 **登录完成后的前端回跳 URL**、可选 **加密主密钥**。
- **`sys_user_oauth_bindings`**：用户与 IdP `subject` 的绑定关系。

Casbin 中已种子化 `888` 等角色对 `/sysOAuthProvider/*`、`/sysOAuthSetting/*` 的访问规则；公开 `/base/oauth/*` 不走 Casbin。

---

## 3. 在 IdP 控制台需要配置的回调地址

授权码回调必须指向**后端对外可访问的 API**，路径模板为（含 `system.router-prefix`）：

```text
{公网 API 根路径}{RouterPrefix}/base/oauth/callback/{kind}
```

- `{kind}` 为平台英文标识，与后台「平台」列及代码注册名一致，**小写**（如 `github`、`google`）。
- `RouterPrefix` 来自 `config.yaml` 的 `system.router-prefix`（空则路径以 `/base/oauth/...` 开头）。

后台 **全局设置** 接口会返回字段 **`callbackPathPattern`**（仅路径、含前缀），管理页会展示完整 URL 模板供复制。登录页拉取 **`GET /base/oauth/providers`** 时也会在 `data.callbackPathPattern` 中返回同一路径，便于前端与 `VITE_BASE_API` 的 origin 拼接授权链接。

**授权跳转入口**（由前端或浏览器访问）：

```text
{API 根 URL}{RouterPrefix}/base/oauth/authorize/{kind}
```

可选查询参数：`redirect` — 登录成功后希望回到前端的站内路径（须以 `/` 开头，且经过服务端白名单式校验，见 `sanitizeOAuthReturnPath`）。

---

## 4. 公开 HTTP 接口（无需登录）

均在 `{RouterPrefix}/base/...` 下注册，`GET` 为主：

| 路径 | 说明 |
|------|------|
| `GET /base/oauth/providers` | 返回已启用且配置完整的提供商列表 + `callbackPathPattern` |
| `GET /base/oauth/authorize/:kind` | 302 跳转至 IdP 授权页 |
| `GET /base/oauth/callback/:kind` | IdP 回调；换票成功后 302 到全局配置的前端地址，并带 `oauth_ex=...` |
| `GET /base/oauth/exchange?oauth_ex=...` | 前端用一次性 `oauth_ex` 换取登录 JSON（与站内登录接口响应结构衔接） |

流程要点：

1. 默认使用 **PKCE**；`state` 与 `code_verifier` 存在全局 **BlackCache**，有时效。**微信开放平台网站应用（`wechat_open`）等平台不走 PKCE**，仅依赖 `state` 与一次性授权码。
2. 回调成功后重定向到 **`sys_oauth_settings.frontend_redirect`**（未配置时有开发用默认）；URL 上带 **`oauth_ex`**。
3. 前端应用 **`exchange`** 一次即失效，防止重放。

若 IdP 返回错误查询参数 `error`（如用户拒绝授权），会重定向到前端并带 **`oauth_err`**（见下一节）。

---

## 5. 登录失败：`oauth_err` 查询参数

回调或授权失败时，浏览器会被重定向到全局配置的前端登录地址，并附带 `oauth_err=<code>`（URL 编码）。前端可按码展示文案，常见取值：

| code | 含义（简述） |
|------|----------------|
| `denied` | 用户在 IdP 侧取消或拒绝授权（含常见 `access_denied` 等） |
| `idp` | IdP 返回其它 `error` 参数 |
| `authorize` | 无法构造授权跳转（未启用、配置不全、未知 kind 等） |
| `missing` | 缺少 `code` 或 `state` |
| `state` | state 过期、篡改或 kind 不一致 |
| `provider` | 库中无启用配置或 Lookup 失败 |
| `token` | 换票或拉取用户资料失败 |
| `user` | 建用户 / 绑定时失败或账号被禁用 |
| `session` | 建立登录会话失败 |
| `server` | 服务端生成换票 ID 等异常 |
| `unknown` | 其它未分类错误 |

历史兼容：部分链接可能仍为 `oauth_err=1`，前端可按未知错误处理。

---

## 6. 限流（可选）

配置项 **`rag.oauth-ip-limit-per-minute`**（`config.yaml` 中 `rag` 段）：

- `> 0` 且 **Redis 可用** 时，对 `/base/oauth/providers`、`authorize`、`callback`、`exchange` 按 **客户端 IP** 做滑动窗口限流。
- Redis 异常时 **fail-open**（不阻断请求）。

与对话 Webhook 的 `rag.channel-webhook-ip-limit-per-minute` 相互独立。

---

## 7. 已注册的 `kind` 与扩展 `extra`（JSON）

代码中已注册的 `kind`（字典序，与 `server/oauth` 内 `Register` 一致）包括：

`auth0`，`bitbucket`，`cognito`，`discord`，`dropbox`，`facebook`，`figma`，`gitee`，`github`，`gitlab`，`google`，`kakao`，`line`，`linkedin`，`microsoft`，`okta`，`paypal`，`slack`，`spotify`，`strava`，`twitch`，`twitter`，`wechat_open`，`yandex`，`zoom`。

管理端新建提供商时 **`kind` 会规范为小写**；按 `kind` 查询与绑定支持大小写不敏感。

**扩展 JSON**（写入 `sys_oauth_providers.extra`，对象类型）常用键（与 `server/oauth/doc.go` 一致）：

| 平台 / 场景 | 键名 | 说明 |
|-------------|------|------|
| Microsoft | `tenant` | 租户，如 `common`、`organizations`、具体租户 ID |
| GitLab 自建 | `gitlab_base_url` | GitLab 实例根 URL |
| Facebook | `facebook_graph_version` | Graph API 版本，如 `v21.0` |
| GitHub Enterprise | `github_enterprise_host` | 企业版主机名 |
| Auth0 | `auth0_domain` | 租户域 |
| PayPal | `paypal_sandbox` | `true` 时使用沙箱端点 |
| Okta | `okta_issuer` 或 `okta_domain` + `okta_auth_server` | 签发者或与域名、授权服务器组合 |
| Cognito | `cognito_domain` | 托管域主机名 |

### 7.1 微信扫码快捷登录（`wechat_open`）

对应 [微信开放平台](https://open.weixin.qq.com/) — **网站应用**（非微信公众号、非小程序）。用户在 PC 登录页点击后跳转微信扫码页，授权后回调本系统换票并完成登录。

| 配置项 | 说明 |
|--------|------|
| **kind** | 固定填 `wechat_open`（小写） |
| **Client ID** | 开放平台网站应用的 **AppID** |
| **Client Secret** | 网站应用的 **AppSecret** |
| **Scopes** | 可留空；默认 `snsapi_login`（网页扫码仅支持此 scope） |
| **授权回调域** | 在微信开放平台控制台「网站应用 → 开发信息」中配置；须与回调 URL 的 **域名** 一致（不含路径）。实际回调路径仍为本文第 3 节模板：`…/base/oauth/callback/wechat_open` |

**用户标识**：绑定用户时优先使用 **`unionid`**（若开放平台返回且该应用已绑定到同一开放平台账号）；否则使用 **`openid`**。微信通常**不返回邮箱**，新用户将仅有昵称与头像 URL。

**与公众号/企微的区别**：微信公众号、企业微信的对话 Webhook 见 [THIRD_PARTY_CHANNEL_CONNECTORS.md](./THIRD_PARTY_CHANNEL_CONNECTORS_zh.md)；与本节 OAuth 登录为不同产品与不同凭证，勿混用。

---

## 8. 管理端 API 摘要（需 JWT + 权限）

- **提供商**：`/sysOAuthProvider/createOAuthProvider`（POST）、`updateOAuthProvider`（PUT）、`deleteOAuthProvider`（DELETE）、`getOAuthProviderList`（GET）、`findOAuthProvider`（GET）、`getRegisteredOAuthKinds`（GET）等。
- **全局设置**：`/sysOAuthSetting/getOAuthSetting`（GET）、`/sysOAuthSetting/updateOAuthSetting`（PUT）。

具体 Method 与权限以 `server/source/system/casbin.go`、`server/source/system/api.go` 为准。

---

## 9. 运维与升级提示

- **加密主密钥**：若单独配置 `sys_oauth_settings` 中的密钥，修改后已加密的各平台 Client Secret 需**重新录入**。
- **前端环境变量**：`VITE_BASE_API` 应与浏览器能访问到的 API 根一致；`providers` 接口返回的 `callbackPathPattern` 可纠正「仅域名、漏写 router-prefix」类问题。
- **菜单组件路径**：若数据库中菜单 `component` 仍为历史值 `view/superAdmin/oauth/oauthProvider.vue`，请改为 **`view/oauth/settings.vue`**，否则动态路由无法加载页面。

---

## 10. 扩展新 IdP（开发）

1. 在 `server/oauth/` 新增文件，实现 `Provider` 接口，在 `init()` 中调用 `oauth.Register`。
2. 若需租户/域名等，在 `MergeScopes` / `OAuth2Config` / `FetchProfile` 中读取 `RuntimeConfig.Extra`。
3. 若授权 URL 或换票流程与标准 OAuth2+PKCE 不兼容，可额外实现 `AuthorizeURLBuilder`、`CodeExchanger`（定义见 `server/oauth/types.go`），参考 `wechat_open.go`。
4. 无需改表结构；后台即可选择新 `kind`（来自 `getRegisteredOAuthKinds`）。

更多实现细节见包注释：`server/oauth/doc.go`、`server/service/oauthapp/doc.go`。
