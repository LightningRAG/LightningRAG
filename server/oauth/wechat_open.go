package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// 微信开放平台「网站应用」扫码登录（OAuth2 授权码），非公众号/小程序。
// 文档：https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html

type wechatOpenProvider struct{}

func init() {
	Register(&wechatOpenProvider{})
}

func (wechatOpenProvider) Kind() string               { return "wechat_open" }
func (wechatOpenProvider) DefaultDisplayName() string { return "WeChat" }
func (wechatOpenProvider) DefaultScopes() []string {
	return []string{"snsapi_login"}
}

func (wechatOpenProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://open.weixin.qq.com/connect/qrconnect",
			TokenURL: "https://api.weixin.qq.com/sns/oauth2/access_token",
		},
	}
}

func (wechatOpenProvider) BuildAuthorizeURL(rc RuntimeConfig, state string) (string, error) {
	appid := strings.TrimSpace(rc.ClientID)
	if appid == "" {
		return "", fmt.Errorf("wechat_open: missing AppID (client id)")
	}
	redir := strings.TrimSpace(rc.RedirectURI)
	if redir == "" {
		return "", fmt.Errorf("wechat_open: missing redirect_uri")
	}
	scope := strings.TrimSpace(strings.Join(rc.Scopes, " "))
	if scope == "" {
		scope = "snsapi_login"
	}
	v := url.Values{}
	v.Set("appid", appid)
	v.Set("redirect_uri", redir)
	v.Set("response_type", "code")
	v.Set("scope", scope)
	v.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + v.Encode() + "#wechat_redirect", nil
}

func (wechatOpenProvider) ExchangeCode(ctx context.Context, code string, rc RuntimeConfig) (*oauth2.Token, error) {
	appid := strings.TrimSpace(rc.ClientID)
	secret := strings.TrimSpace(rc.ClientSecret)
	if appid == "" || secret == "" {
		return nil, fmt.Errorf("wechat_open: missing AppID or AppSecret")
	}
	u := "https://api.weixin.qq.com/sns/oauth2/access_token?" + url.Values{
		"appid":      {appid},
		"secret":     {secret},
		"code":       {strings.TrimSpace(code)},
		"grant_type": {"authorization_code"},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var raw struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		Scope        string `json:"scope"`
		UnionID      string `json:"unionid"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("wechat_open token: %w", err)
	}
	if raw.ErrCode != 0 {
		return nil, fmt.Errorf("wechat_open token: errcode=%d errmsg=%s", raw.ErrCode, raw.ErrMsg)
	}
	if raw.AccessToken == "" || strings.TrimSpace(raw.OpenID) == "" {
		return nil, fmt.Errorf("wechat_open: empty access_token or openid")
	}
	t := &oauth2.Token{
		AccessToken:  raw.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: raw.RefreshToken,
	}
	if raw.ExpiresIn > 0 {
		t.Expiry = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second)
	}
	ex := map[string]interface{}{
		"openid":  raw.OpenID,
		"unionid": raw.UnionID,
	}
	return t.WithExtra(ex), nil
}

func (wechatOpenProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	openID, _ := tok.Extra("openid").(string)
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, fmt.Errorf("wechat_open userinfo: missing openid on token")
	}
	unionID, _ := tok.Extra("unionid").(string)
	unionID = strings.TrimSpace(unionID)

	u := "https://api.weixin.qq.com/sns/userinfo?" + url.Values{
		"access_token": {tok.AccessToken},
		"openid":       {openID},
		"lang":         {"zh_CN"},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info struct {
		OpenID     string `json:"openid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("wechat_open userinfo: %w", err)
	}
	if info.ErrCode != 0 {
		return nil, fmt.Errorf("wechat_open userinfo: errcode=%d errmsg=%s", info.ErrCode, info.ErrMsg)
	}
	sub := strings.TrimSpace(info.UnionID)
	if sub == "" {
		sub = strings.TrimSpace(unionID)
	}
	if sub == "" {
		sub = strings.TrimSpace(info.OpenID)
	}
	if sub == "" {
		sub = openID
	}
	name := strings.TrimSpace(info.Nickname)
	return &NormalizedProfile{
		Subject:   sub,
		Email:     "",
		Name:      name,
		AvatarURL: strings.TrimSpace(info.HeadImgURL),
	}, nil
}
