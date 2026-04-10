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

type lineProvider struct{}

func init() {
	Register(&lineProvider{})
}

func (lineProvider) Kind() string               { return "line" }
func (lineProvider) DefaultDisplayName() string { return "LINE" }
func (lineProvider) DefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}

func (lineProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://access.line.me/oauth2/v2.1/authorize",
			TokenURL: "https://api.line.me/oauth2/v2.1/token",
		},
	}
}

func (lineProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.line.me/oauth2/v2.1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
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
		return nil, fmt.Errorf("line userinfo: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Email   string `json:"email"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	return &NormalizedProfile{
		Subject:   strings.TrimSpace(u.Sub),
		Email:     strings.TrimSpace(u.Email),
		Name:      strings.TrimSpace(u.Name),
		AvatarURL: strings.TrimSpace(u.Picture),
	}, nil
}
