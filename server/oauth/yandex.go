package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type yandexProvider struct{}

func init() {
	Register(&yandexProvider{})
}

func (yandexProvider) Kind() string               { return "yandex" }
func (yandexProvider) DefaultDisplayName() string { return "Yandex" }
func (yandexProvider) DefaultScopes() []string {
	return []string{"login:info", "login:email"}
}

func (yandexProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.yandex.com/authorize",
			TokenURL: "https://oauth.yandex.com/token",
		},
	}
}

func (yandexProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+tok.AccessToken)
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("yandex info: HTTP %d %s", resp.StatusCode, string(body))
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	sub := yandexStringID(raw["id"])
	login := yandexJSONString(raw["login"])
	displayName := yandexJSONString(raw["display_name"])
	email := yandexJSONString(raw["default_email"])
	if sub == "" {
		sub = strings.TrimSpace(login)
	}
	if sub == "" {
		return nil, fmt.Errorf("yandex: empty subject")
	}
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = strings.TrimSpace(login)
	}
	if name == "" {
		name = sub
	}
	avatar := ""
	if sub != "" {
		avatar = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", sub)
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     strings.TrimSpace(email),
		Name:      name,
		AvatarURL: avatar,
	}, nil
}

func yandexStringID(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil && s != "" {
		return s
	}
	var f float64
	if json.Unmarshal(raw, &f) == nil {
		return fmt.Sprintf("%.0f", f)
	}
	return strings.Trim(string(raw), `"`)
}

func yandexJSONString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return ""
}
