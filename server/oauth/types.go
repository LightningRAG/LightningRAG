package oauth

import (
	"context"
	"strings"

	"golang.org/x/oauth2"
)

// NormalizedProfile 各平台归一化后的用户信息
type NormalizedProfile struct {
	Subject   string
	Email     string
	Name      string
	AvatarURL string
}

// RuntimeConfig 单次授权流程中的运行时配置（来自库表 + 解密 + 回调 URL）
type RuntimeConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	Extra        map[string]any
}

// Provider 第三方 OAuth 平台契约：在独立文件中 init 注册
type Provider interface {
	Kind() string
	DefaultDisplayName() string
	// DefaultScopes 当库表 scopes 为空时使用
	DefaultScopes() []string
	OAuth2Config(rc RuntimeConfig) *oauth2.Config
	// FetchProfile 用 access_token 拉取用户信息
	FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error)
}

// AuthorizeURLBuilder 若 Provider 实现此接口，授权跳转使用该 URL，且不启用 PKCE（state 中不保存 code_verifier）。
// 适用于授权参数非标准（如 appid）、或 URL 需固定 fragment 的 IdP（例如微信开放平台网站应用扫码登录）。
type AuthorizeURLBuilder interface {
	BuildAuthorizeURL(rc RuntimeConfig, state string) (string, error)
}

// CodeExchanger 若 Provider 实现此接口，用授权码换 token 时走自定义逻辑（如 GET 换票、非标准 JSON）。
type CodeExchanger interface {
	ExchangeCode(ctx context.Context, code string, rc RuntimeConfig) (*oauth2.Token, error)
}

// MergeScopes 库表 scopes 字符串（空格分隔）与默认值合并
func MergeScopes(dbScopes string, def []string) []string {
	s := strings.TrimSpace(dbScopes)
	if s == "" {
		return def
	}
	return strings.Fields(s)
}
