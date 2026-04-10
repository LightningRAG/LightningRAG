// Package oauthapp 实现「第三方快捷登录」业务层：库表配置、授权/回调/换票流程、全局设置。
//
// 分层说明：
//   - server/oauth          — 各 IdP 的 OAuth2 协议适配（Register/Lookup、换票、拉取用户资料），无业务状态
//   - server/service/oauthapp — 本包：读写 sys_oauth_*、会话 state、与用户/菜单服务协作
//   - server/model/system   — 数据模型 SysOAuthProvider、SysOAuthSetting、SysUserOAuthBinding（表名不变）
//   - server/api/v1/oauthapp — HTTP 处理器（公开 /base/oauth/* 与管理端 sysOAuth*）
//   - server/global         — OAuth 全局运行时缓存（回跳 URL、加密主密钥材料）
//   - server/utils          — Client Secret 加解密（oauth_crypt）
package oauthapp
