package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/oauth2"
)

type giteeProvider struct{}

func init() {
	Register(&giteeProvider{})
}

func (giteeProvider) Kind() string               { return "gitee" }
func (giteeProvider) DefaultDisplayName() string { return "Gitee" }
func (giteeProvider) DefaultScopes() []string {
	return []string{"user_info", "emails"}
}

func (giteeProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://gitee.com/oauth/authorize",
			TokenURL: "https://gitee.com/oauth/token",
		},
	}
}

func (giteeProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	u := url.Values{}
	u.Set("access_token", tok.AccessToken)
	reqURL := "https://gitee.com/api/v5/user?" + u.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitee user: HTTP %d %s", resp.StatusCode, string(body))
	}
	var raw struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	sub := strconv.Itoa(raw.ID)
	name := raw.Name
	if name == "" {
		name = raw.Login
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     raw.Email,
		Name:      name,
		AvatarURL: raw.AvatarURL,
	}, nil
}
