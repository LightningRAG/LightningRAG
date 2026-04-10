// Package oauth 仅负责各 IdP 的 OAuth2 协议侧适配（无数据库、无用户业务）。
//
// 与「快捷登录」产品相关的其它层次（不要写在本包）：
//   - server/service/oauthapp — 库表、授权/回调/换票、与 User/Menu 服务协作
//   - server/api/v1/oauthapp — HTTP 处理器
//   - server/router/oauthapp — 管理端路由；公开 /base/oauth/* 在 router/system/sys_base.go 注册
//
// 新增平台：在本包新增 xxx.go，实现 Provider 接口，在 init() 中调用 Register。
//
// 已注册 kind：github, google, gitee, gitlab, microsoft, discord, facebook, linkedin, slack, bitbucket,
// twitch, spotify, line, zoom, twitter, kakao, dropbox, auth0, yandex, paypal, okta, cognito, strava, figma,
// wechat_open（微信开放平台网站应用扫码登录）。
//
// 扩展 JSON：microsoft→tenant；gitlab→gitlab_base_url；facebook→facebook_graph_version；
// github→github_enterprise_host；auth0→auth0_domain；paypal→paypal_sandbox（bool，沙箱端点）；
// okta→okta_issuer 或 okta_domain[+okta_auth_server]；cognito→cognito_domain（托管域主机名）。
// 需非标准授权 URL 或换票方式的 IdP 可实现 AuthorizeURLBuilder / CodeExchanger（见 types.go）。
package oauth
